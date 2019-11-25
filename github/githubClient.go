package github

import (
	"context"
	"fmt"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	auditHttp "iq-scm-audit/http"
	"log"
	"net/http"
	"os"
)

const cloudApiUrl = "https://api.github.com"
const graphQlEndpoint = "/graphql"
const issueEndpoint = "/repos/%v/issues"

type GitHubClient struct {
	Token string
}

type (
	RepositorySearch struct {
		RepositoryCount int
		Nodes[] Repository
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	}
)

type Repository struct {
	RepositoryFragment struct {
		Name string
		NameWithOwner string
		Url string
		SshUrl string
		DependencyGraphManifests DependencyGraphManifests
		Packages Packages `graphql:"packages(last: 1)"`
		Releases Releases `graphql:"releases(last: 1)"`
	} `graphql:"... on Repository"`
}

type (
	DependencyGraphManifests struct {
		TotalCount int
		Nodes[] struct {
			Filename     string
			Dependencies struct {
				Nodes[] Dependency
			}
		}
	}
)

type (
	Releases struct {
		Nodes[] struct {
			Url string
			ReleaseAssets struct {
				Nodes[] struct {
					Name string
					Url string
				}
			} `graphql:"releaseAssets(last: 10)"`
		}
	}
)

type (
	Packages struct {
		TotalCount int
		Nodes[] struct {
			Name     string
			LatestVersion struct {
				Files struct {
					Nodes[] struct {
						Name string
						Url string
					}
				} `graphql:"files(last: 10)"`
			}
		}
	}
)

type Dependency struct {
	PackageManager string
	PackageName    string
	Requirements   string
}

type transport struct {}

func NewGitHubClient(token string) *GitHubClient {
	var gitHubClient = new(GitHubClient)
	gitHubClient.Token = token
	return gitHubClient
}

func(client *GitHubClient) GetRepositories(query string) []Repository {
	httpClient := newGraphQlClient(client.Token)
	variables := map[string] interface {} {
		"queryString": githubv4.String(query + " fork:true"),
		"repositoryCursor":  (*githubv4.String)(nil),
	}
	var allRepositories []Repository
	var errors []string
	for {
		var query struct {
			Search RepositorySearch `graphql:"search(query: $queryString, type: REPOSITORY, first: 5, after: $repositoryCursor)"`
		}
		err := httpClient.Query(context.Background(), &query, variables)
		if err != nil {
			errors = append(errors, err.Error())
			if len(errors) > 9 {
				log.Print("GraphQL Query to GitHub failed.")
				for _, queuedError := range errors {
					log.Print(queuedError)
				}
				os.Exit(1)
			}
		}
		allRepositories = append(allRepositories, query.Search.Nodes...)
		if !query.Search.PageInfo.HasNextPage {
			break
		}
		variables["repositoryCursor"] = githubv4.NewString(query.Search.PageInfo.EndCursor)
	}
	return allRepositories
}

func (client *GitHubClient) CreateIssue(repositoryNameWithOwner string, title string, markdown string) {
	httpClient := newHttpClient(client.Token)
	httpClient.HttpPost(cloudApiUrl + fmt.Sprintf(issueEndpoint, repositoryNameWithOwner), map[string] string {
		"title": title,
		"body": markdown,
	})
}

func (client *GitHubClient) DownloadRelease(url string) []byte {
	httpClient := new(auditHttp.HttpClient)
	return httpClient.HttpGet(url)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/vnd.github.hawkgirl-preview+json")
	req.Header.Add("Accept", "application/vnd.github.packages-preview+json")
	return http.DefaultTransport.RoundTrip(req)
}

func newGraphQlClient(token string) *githubv4.Client {
	httpClient := &http.Client{Transport: &transport{}}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	oauthClient := oauth2.NewClient(ctx, src)
	return githubv4.NewEnterpriseClient(cloudApiUrl+graphQlEndpoint, oauthClient)
}

func newHttpClient(token string) *auditHttp.HttpClient {
	client := new(auditHttp.HttpClient)
	client.Token = token
	return client
}