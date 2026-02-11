terraform {
  required_providers {
    bitbucket = {
      source  = "github.com/scastria/bitbucket"
    }
    sonarcloudextra = {
      source  = "github.com/scastria/sonarcloudextra"
    }
  }
}

provider "bitbucket" {
}

provider "sonarcloudextra" {}

data "bitbucket_project" "Project" {
  key = "ARCH"
}

resource "bitbucket_repository" "repo" {
  project_id = data.bitbucket_project.Project.id
  name = "shawn-sonar-test"
  use_existing = true
}

resource "sonarcloudextra_project" "sonar-test" {
  organization      = "greenstreetadvisors"
  name              = bitbucket_repository.repo.name
  bitbucket_repo_uuid = bitbucket_repository.repo.id
  use_existing = true
}
