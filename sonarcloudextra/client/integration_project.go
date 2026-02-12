package client

import "strings"

const (
	ProjectSearchPath        = "api/projects/search"
	ProjectsDeletePath       = "api/projects/delete"
	AlmProvisionProjectsPath = "api/alm_integration/provision_projects"
	AlmListRepositoriesPath  = "api/alm_integration/list_repositories"
)

type IntegrationProject struct {
	Organization      string
	Name              string
	BitbucketRepoUuid string
	UseExisting       bool
}

type IntegrationProjectSearchResponse struct {
	Components []IntegrationProjectComponent `json:"components,omitempty"`
}

type IntegrationProjectComponent struct {
	Organization string `json:"organization,omitempty"`
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	UseExisting  bool   `json:"-"`
}

type IntegrationAlmListRepositoriesResponse struct {
	Repositories []AlmRepository `json:"repositories,omitempty"`
}

type AlmRepository struct {
	InstallationKey string                        `json:"installationKey,omitempty"`
	LinkedProjects  []IntegrationProjectComponent `json:"linkedProjects,omitempty"`
}

func (ip *IntegrationProject) IntegrationProjectEncodeId() string {
	return ip.Organization + IdSeparator + ip.Name
}

func IntegrationProjectDecodeId(s string) (string, string) {
	tokens := strings.Split(s, IdSeparator)
	return tokens[0], tokens[1]
}

func IntegrationProjectEncodeSonarId(s1 string, s2 string) string {
	return s1 + SonarSeparator + s2
}
