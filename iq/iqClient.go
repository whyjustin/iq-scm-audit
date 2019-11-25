package iq

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	auditHttp "iq-scm-audit/http"
	"iq-scm-audit/sbom"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const apiEndpoint = "/api/v2/"
const applicationsEndpoint = apiEndpoint + "applications/"
const applicationScmEndpoint = apiEndpoint + "sourceControl/application/"
const organizationsEndpoint = apiEndpoint + "organizations/"
const organizationScmEndpoint = apiEndpoint + "sourceControl/organization/"
const scanEndpoint = apiEndpoint + "scan/applications/"

type IqClient struct {
	IqServerUrl string
	Username string
	Password string
}

type Applications struct {
	Applications[] Application
}

type Application struct {
	Id string
	PublicId string
	Name string
	RepositoryUrl string `json:"-"`
	ReportUrl string `json:"-"`
}

type ApplicationScm struct {
	OwnerId string
	RepositoryUrl string
}

type Organizations struct {
	Organizations[] Organization
}

type Organization struct {
	Id string
	Name string
}

type SbomScanTicket struct {
	StatusUrl string
}

type SbomScanResult struct {
	PolicyAction       string
	ReportHtmlUrl string
	IsError            bool
}

type ApplicationEvaluationResult struct {
	ReportHtmlUrl string
}

func NewIqClient(iqServerUrl string, username string, password string) *IqClient {
	var iqClient = new(IqClient)
	iqClient.IqServerUrl = iqServerUrl
	iqClient.Username = username
	iqClient.Password = password
	return iqClient
}

func (client *IqClient) GetApplications() *Applications {
	var getBytes = client.getHttpClient().HttpGet(client.IqServerUrl + applicationsEndpoint)
	var applications = new(Applications)
	getError := json.Unmarshal(getBytes, &applications)
	if getError != nil {
		log.Fatal(string(getBytes))
	}
	for index := range applications.Applications {
		application := &applications.Applications[index]
		var repositoryUrl = client.GetApplicationScm(application.Id).RepositoryUrl
		application.RepositoryUrl = repositoryUrl
	}
	return applications
}

func (client *IqClient) GetOrCreateOrganization(organizationName string) *Organization {
	getBytes := client.getHttpClient().HttpGet(client.IqServerUrl + organizationsEndpoint)
	organizations := new(Organizations)
	getError := json.Unmarshal(getBytes, &organizations)
	if getError != nil {
		log.Fatal(string(getBytes))
	}
	for _, organization := range organizations.Organizations {
		if organization.Name == organizationName {
			log.Println("Found existing organization - " + organization.Name + ":" + organization.Id)
			return &organization
		}
	}

	var postBytes = client.getHttpClient().HttpPost(client.IqServerUrl + organizationsEndpoint, map[string]string {
		"name": organizationName,
	})
	organization := new(Organization)
	postError := json.Unmarshal(postBytes, &organization)
	if postError != nil {
		log.Fatal(string(postBytes))
	}
	return organization
}

func (client *IqClient) GetOrCreateApplication(organizationId string, publicId string, name string) *Application {
	getBytes := client.getHttpClient().HttpGet(client.IqServerUrl + applicationsEndpoint + "?publicId=" + publicId)
	applications := new(Applications)
	var application Application
	getError := json.Unmarshal(getBytes, &applications)
	if getError != nil {
		log.Fatal(string(getBytes))
	}
	if len(applications.Applications) > 0 {
		application = applications.Applications[0]
		log.Println("Found existing application - " + application.Name + ":" + application.PublicId)
		return &application
	}

	postBytes := client.getHttpClient().HttpPost(client.IqServerUrl + applicationsEndpoint, map[string]string {
		"publicId": publicId,
		"name": name,
		"organizationId": organizationId,
	})
	postError := json.Unmarshal(postBytes, &application)
	if postError != nil {
		log.Fatal(string(postBytes))
	}
	return &application
}

func (client *IqClient) GetApplicationScm(applicationId string) *ApplicationScm {
	getBytes := client.getHttpClient().HttpGet(client.IqServerUrl + applicationScmEndpoint + applicationId)
	var applicationScm = new(ApplicationScm)
	getError := json.Unmarshal(getBytes, &applicationScm)
	if getError != nil {
		// IQ Server returns error if SCM is not configured
		return applicationScm
	}
	return applicationScm
}

func(client *IqClient) SetOrganizationScm(organizationId string, token string) {
	client.getHttpClient().HttpPost(client.IqServerUrl + organizationScmEndpoint + organizationId, map[string]string {
		"token": token,
		"provider": "GitHub",
	})
}

func (client *IqClient) SetApplicationScm(applicationId string, repositoryUrl string) {
	client.getHttpClient().HttpPost(client.IqServerUrl + applicationScmEndpoint + applicationId, map[string]string {
		"repositoryUrl": repositoryUrl,
	})
}

func (client *IqClient) ScanSbom(applicationId string, sbom sbom.Sbom) *SbomScanTicket {
	postBytes := client.getHttpClient().HttpPostXml(client.IqServerUrl + scanEndpoint + applicationId + "/sources/cyclone", sbom)
	sbomTicket := new(SbomScanTicket)
	postError := json.Unmarshal(postBytes, &sbomTicket)
	if postError != nil {
		log.Fatal(string(postBytes))
	}
	return sbomTicket
}

func (client *IqClient) GetSbomScanResult(statusUrl string) *SbomScanResult {
	sbomScanResult := new(SbomScanResult)
	var errorQueue []string
	for {
		getBytes := client.getHttpClient().HttpGet(client.IqServerUrl + "/" + statusUrl)
		getError := json.Unmarshal(getBytes, &sbomScanResult)
		if getError == nil {
			break
		} else {
			errorQueue = append(errorQueue, string(getBytes))
			if len(errorQueue) > 30 {
				for _, errorMessage := range errorQueue {
					log.Println(errorMessage)
				}
				os.Exit(1)
			}
		}
		time.Sleep(1 * time.Second)
	}
	return sbomScanResult
}

func (client *IqClient) Evaluate(path string, applicationId string, stage string) *ApplicationEvaluationResult {
	jarLocation, _ := filepath.Abs("./iq/nexus-iq-cli-1.78.0-02.jar")
	resultsFilePath := filepath.Join(path, "evaluation-results.json")
	evaluateCommand := exec.Command("java", "-jar", jarLocation, "-s", client.IqServerUrl, "-a", client.Username + ":" + client.Password, "-i", applicationId, "-t", stage, "-r", resultsFilePath, path)
	var stdout, stderr bytes.Buffer
	evaluateCommand.Stdout = &stdout
	evaluateCommand.Stderr = &stderr
	exitError := evaluateCommand.Run()
	if exitError != nil {
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		log.Println(outStr)
		log.Fatal(errStr)
	}
	jsonFile, openError := os.Open(resultsFilePath)
	if openError != nil {
		log.Fatal(openError.Error())
	}
	defer jsonFile.Close()

	resultBytes, readError := ioutil.ReadAll(jsonFile)
	if readError != nil {
		log.Fatal(readError.Error())
	}

	applicationEvaluationResult := new(ApplicationEvaluationResult)
	unmarshalError := json.Unmarshal(resultBytes, applicationEvaluationResult)
	if unmarshalError != nil {
		log.Fatal(unmarshalError.Error())
	}

	return applicationEvaluationResult
}


func (client *IqClient) getHttpClient() *auditHttp.HttpClient {
	httpClient := new(auditHttp.HttpClient)
	httpClient.Username = client.Username
	httpClient.Password = client.Password
	return httpClient
}