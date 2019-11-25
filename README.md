### IQ SCM Audit

#### Overview

This tool will take a GitHub graphql repository search query to fetch a list of GitHub
source code repositories and for each repository:

- Create an IQ Application
- Configure IQ Application against source control
- Scan GitHub reported dependencies against IQ Application using third party data API
- Download and evaluate policy against latest GitHub release assets
- Download and evaluate policy against latest GitHub Packages assets
- Create GitHub Issue in repository with results and hints on how to configure CI tools

#### Usage

```
Usage:
iq-scm-audit [options]
  -gitHubQuery string
    	Query String for GitHub graphql repository search (GITHUB_QUERY)
  -gitHubToken string
    	GitHub Token (GITHUB_TOKEN)
  -iqOrganization string
    	Organization to create new applications (IQ_ORGANIZATION)
  -iqPassword string
    	Nexus IQ Password (IQ_PASSWORD)
  -iqServerUrl string
    	Nexus IQ Server Url (IQ_SERVER_URL)
  -iqUsername string
    	Nexus IQ Username (IQ_USERNAME)
  -iqcontact string
    	Email of person to contact for access to Nexus IQ (IQ_CONTACT)
  -skipExistingApplications
    	Skip Audit and Evaluation against existing applications
  -skipIQEvaluations
    	Skip IQ Evaluations against latest Release or Package assets
  -skipIssueCreation
    	Skip GitHub Issue Creation
```

#### Example Queries

Queries can be formed to search for organizations:

```
org:whyjustin
```

or particular repositories:

```
whyjustin/spring-hello-webmvc
```