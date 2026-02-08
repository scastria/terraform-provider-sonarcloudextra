terraform {
  required_providers {
    sonarcloudextra = {
      source  = "scastria/sonarcloudextra"
    }
  }
}

provider "sonarcloudextra" {}

resource "sonarcloudextra_project" "sonar-test" {
  organization      = "greenstreetadvisors"
  name              = "sonar-test"
  installation_keys = "{0be8dc28-2b65-4de0-bd14-ebdc298403a6}"
  use_existing      = true
}