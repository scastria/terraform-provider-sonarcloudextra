# Resource: sonarcloudextra_integration_project
Represents a integration_project
## Example usage
```hcl
resource "sonarcloudextra_integration_project" "sonar_test" {
  organization        = "your_sonar_organization"
  name                = bitbucket_repository.repo.name
  bitbucket_repo_uuid = bitbucket_repository.repo.id
  use_existing        = true
}
```
## Argument Reference
* `name` - **(Required, String, ForceNew)** Project name (used to build the Sonar project key in this provider).
* `organization` - **(Required, String, ForceNew)** Sonar organization key.
* `bitbucket_repo_uuid` - **(Required, String, ForceNew)** Bitbucket repository UUID/installation key used by Sonar ALM provisioning.
* `use_existing` - **(Optional, Boolean, ForceNew)** Create-only. When true, the provider will first search SonarCloud for an existing project matching the derived `project key`. Prevents the need for an import. Default: false.
## Attribute Reference
* `id` - **(String)** The internal Terraform ID in the format: <organization>:<name>.
## Import
Integration projects can be imported using a proper value of `id` as described above
