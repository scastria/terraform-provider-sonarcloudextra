package client

const (
	ProjectSearchPath        = "api/projects/search%s"
	ProjectsDeletePath       = "api/projects/delete%s"
	AlmProvisionProjectsPath = "api/alm_integration/provision_projects%s"
	AlmListRepositoriesPath  = "api/alm_integration/list_repositories%s"
)

type Project struct {
	Organization     string `json:"-"`
	Name             string `json:"-"`
	ProjectKey       string `json:"-"`
	InstallationKeys string `json:"-"`
	UseExisting      bool   `json:"-"`
}

type ProjectSearchResponse struct {
	Components []ProjectComponent `json:"components,omitempty"`
}

type ProjectComponent struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

type AlmListRepositoriesResponse struct {
	Repositories []AlmRepository `json:"repositories,omitempty"`
}

type AlmRepository struct {
	Label           string             `json:"label,omitempty"`
	InstallationKey string             `json:"installationKey,omitempty"`
	LinkedProjects  []ProjectComponent `json:"linkedProjects,omitempty"`
	Private         bool               `json:"private,omitempty"`
}
