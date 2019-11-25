package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"iq-scm-audit/github"
	"iq-scm-audit/iq"
	"iq-scm-audit/sbom"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type IssueData struct {
	IqServerUrl string
	AuditReportUrl string
	ReleaseReportUrl string
	PackageReportUrl string
	Repository string
	Contact string
	NameWithOwner string
}

type AuditConfiguration struct {
	GitHubToken              *string
	GitHubQuery              *string
	IqServerUrl              *string
	IqUsername               *string
	IqPassword               *string
	IqOrganization           *string
	IqContact                *string
	SkipIssueCreation        bool
	SkipExistingApplications bool
	SkipIQEvaluations		 bool
}

type RequiredFlag struct {
	Field *string
	Name string
	Usage string
	EnvironmentalVariable string
}

func main() {
	configuration := new(AuditConfiguration)
	configuration.GitHubToken = new(string)
	configuration.GitHubQuery = new(string)
	configuration.IqServerUrl = new(string)
	configuration.IqUsername = new(string)
	configuration.IqPassword = new(string)
	configuration.IqOrganization = new(string)
	configuration.IqContact = new(string)

	var requiredFlags []RequiredFlag
	requiredFlags = appendFlag(requiredFlags, configuration.GitHubToken, "gitHubToken", "GitHub Token", "GITHUB_TOKEN")
	requiredFlags = appendFlag(requiredFlags, configuration.GitHubQuery, "gitHubQuery", "Query String for GitHub graphql repository search", "GITHUB_QUERY")
	requiredFlags = appendFlag(requiredFlags, configuration.IqServerUrl, "iqServerUrl", "Nexus IQ Server Url", "IQ_SERVER_URL")
	requiredFlags = appendFlag(requiredFlags, configuration.IqUsername, "iqUsername", "Nexus IQ Username", "IQ_USERNAME")
	requiredFlags = appendFlag(requiredFlags, configuration.IqPassword, "iqPassword", "Nexus IQ Password", "IQ_PASSWORD")
	requiredFlags = appendFlag(requiredFlags, configuration.IqOrganization, "iqOrganization", "Organization to create new applications", "IQ_ORGANIZATION")
	requiredFlags = appendFlag(requiredFlags, configuration.IqContact, "iqcontact", "Email of person to contact for access to Nexus IQ", "IQ_CONTACT")

	flag.BoolVar(&configuration.SkipIssueCreation,"skipIssueCreation", false, "Skip GitHub Issue Creation")
	flag.BoolVar(&configuration.SkipExistingApplications, "skipExistingApplications", false, "Skip Audit and Evaluation against existing applications")
	flag.BoolVar(&configuration.SkipIQEvaluations, "skipIQEvaluations", false, "Skip IQ Evaluations against latest Release or Package assets")

	flag.Usage = func() {
		_, _ = fmt.Fprint(os.Stdout, "Usage: \niq-scm-audit [options]\n")
		flag.PrintDefaults()
	}

	err := flag.CommandLine.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(err.Error())
	}

	for _, requiredFlag := range requiredFlags {
		if len(*requiredFlag.Field) == 0 {
			*requiredFlag.Field = os.Getenv(requiredFlag.EnvironmentalVariable)
		}
		if len(*requiredFlag.Field) == 0 {
			_, _ = fmt.Fprint(os.Stdout, "\nMissing required argument: "+requiredFlag.Usage+". Supply via command line ("+requiredFlag.Name+") or environmental variable ("+requiredFlag.EnvironmentalVariable+").\n")
			flag.Usage()

			os.Exit(1)
		}
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	audit(configuration)
}

func appendFlag(flags []RequiredFlag, field *string, name string, usage string, environmentalVariable string) []RequiredFlag {
	flag.StringVar(field, name, "", usage + " (" + environmentalVariable + ")")
	requiredFlag := new(RequiredFlag)
	requiredFlag.Field = field
	requiredFlag.Name = name
	requiredFlag.Usage = usage
	requiredFlag.EnvironmentalVariable = environmentalVariable
	return append(flags, *requiredFlag)
}

func audit(configuration *AuditConfiguration) {
	log.Println("Getting IQ Applications")
	var iqClient = iq.NewIqClient(*configuration.IqServerUrl, *configuration.IqUsername, *configuration.IqPassword)
	var applications = iqClient.GetApplications()
	log.Println("Getting or Creating IQ Organization - " + *configuration.IqOrganization)
	var scmOrganization = iqClient.GetOrCreateOrganization(*configuration.IqOrganization)
	iqClient.SetOrganizationScm(scmOrganization.Id, *configuration.GitHubToken)

	log.Println("Getting GitHub Repositories")
	var gitHubClient = github.NewGitHubClient(*configuration.GitHubToken)
	var repositories = gitHubClient.GetRepositories(*configuration.GitHubQuery)

	issueTemplate, templateError := template.ParseFiles("github-issue.md")
	if templateError != nil {
		log.Fatal(templateError)
	}
	var issuesData []IssueData

	makeLocalDirectory("work")
	defer removeLocalDirectory("work")

	for _, repository := range repositories {
		var existingConfiguredApplication = false
		if configuration.SkipExistingApplications == true {
			for _, application := range applications.Applications {
				if len(application.RepositoryUrl) > 0 &&
					(application.RepositoryUrl == repository.RepositoryFragment.Url ||
						application.RepositoryUrl == strings.Replace(repository.RepositoryFragment.Url, "https", "http", 1) ||
						application.RepositoryUrl == repository.RepositoryFragment.SshUrl) {
					log.Println("Existing Application Configured, Skipping - " + application.Name + ":" + application.PublicId + ":" + application.RepositoryUrl)
					existingConfiguredApplication = true
					break
				}
			}
			if existingConfiguredApplication == true {
				continue
			}
		}

		log.Println("Creating IQ Application - " + repository.RepositoryFragment.Name)
		var application = iqClient.GetOrCreateApplication(scmOrganization.Id, repository.RepositoryFragment.Name, repository.RepositoryFragment.Name)
		iqClient.SetApplicationScm(application.Id, repository.RepositoryFragment.Url)

		var dependencies []github.Dependency
		for _, dependencyGraph := range repository.RepositoryFragment.DependencyGraphManifests.Nodes {
			dependencies = append(dependencies, dependencyGraph.Dependencies.Nodes...)
		}

		issueData := new(IssueData)

		issueData.IqServerUrl = *configuration.IqServerUrl
		issueData.Repository = application.PublicId
		issueData.Contact = *configuration.IqContact
		issueData.NameWithOwner = repository.RepositoryFragment.NameWithOwner

		if dependencies != nil && len(dependencies) > 0 {
			bom := sbom.NewSbom(dependencies)
			sbomScanTicket := iqClient.ScanSbom(application.Id, *bom)

			sbomScanResult := iqClient.GetSbomScanResult(sbomScanTicket.StatusUrl)
			issueData.AuditReportUrl = sbomScanResult.ReportHtmlUrl
		}

		if !configuration.SkipIQEvaluations {
			if len(repository.RepositoryFragment.Releases.Nodes) > 0 {
				release := repository.RepositoryFragment.Releases.Nodes[0]
				assetDownloadPath := filepath.Join("work", repository.RepositoryFragment.NameWithOwner, "latest-release")
				makeLocalDirectory(assetDownloadPath)
				for _, asset := range release.ReleaseAssets.Nodes {
					assetDownloadLocation := filepath.Join(assetDownloadPath, asset.Name)
					log.Println("Downloading - " + asset.Name)
					downloadRelease(*gitHubClient, assetDownloadLocation, asset.Url)
				}
				log.Println("Evaluating latest release")
				evaluationResult := iqClient.Evaluate(assetDownloadPath, application.PublicId, "stage-release")
				issueData.ReleaseReportUrl = evaluationResult.ReportHtmlUrl
			}

			if len(repository.RepositoryFragment.Packages.Nodes) > 0 {
				pkg := repository.RepositoryFragment.Packages.Nodes[0]
				fileDownloadPath := filepath.Join("work", repository.RepositoryFragment.NameWithOwner, "latest-package")
				makeLocalDirectory(fileDownloadPath)
				for _, file := range pkg.LatestVersion.Files.Nodes {
					fileDownloadLocation := filepath.Join(fileDownloadPath, file.Name)
					log.Println("Downloading - " + file.Name)
					downloadRelease(*gitHubClient, fileDownloadLocation, file.Url)
				}
				log.Println("Evaluating latest package")
				evaluationResult := iqClient.Evaluate(fileDownloadPath, application.PublicId, "release")
				issueData.PackageReportUrl = evaluationResult.ReportHtmlUrl
			}
		}

		issuesData = append(issuesData, *issueData)
	}

	if !configuration.SkipIssueCreation {
		for _, issueData := range issuesData {
			var templateBytes bytes.Buffer
			templateError := issueTemplate.Execute(&templateBytes, issueData)

			if templateError != nil {
				log.Fatal(templateError)
			}

			gitHubClient.CreateIssue(issueData.NameWithOwner, "Configure Nexus IQ", templateBytes.String())
		}
	}
}

func makeLocalDirectory(directory string) {
	// https://github.com/golang/go/issues/22323
	errorMakeDir := os.MkdirAll("." + string(filepath.Separator) + directory, 0700)
	if errorMakeDir != nil {
		log.Fatal(errorMakeDir)
	}
}

func removeLocalDirectory(directory string) {
	errorRemoveDir := os.RemoveAll("." + string(filepath.Separator) + directory)
	if errorRemoveDir != nil {
		log.Fatal(errorRemoveDir)
	}
}

func downloadRelease(gitHubClient github.GitHubClient, path string, url string) {
	releaseBytes := gitHubClient.DownloadRelease(url)
	downloadFile, createError := os.Create(path)
	if createError != nil {
		log.Fatal(createError)
	}
	_, copyError := io.Copy(downloadFile, bytes.NewBuffer(releaseBytes))
	if copyError != nil {
		log.Fatal(copyError)
	}
	closeError := downloadFile.Close()
	if closeError != nil {
		log.Fatal(closeError)
	}
}