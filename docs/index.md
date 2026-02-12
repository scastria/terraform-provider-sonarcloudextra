# SonarCloud Extra Provider
The SonarCloud Extra provider extends the official SonarCloud provider with `use_existing` flags to make it easier
to import existing resources without having to run `terraform import` and using 'api/alm_integration/' endpoint.
## Example Usage
```hcl
terraform {
  required_providers {
    sonarcloudextra = {
      source  = "scastria/sonarcloudextra"
      version = "~> 0.1.0"
    }
  }
}

# Configure the SonarCloud Extra Provider
provider "sonarcloudextra" {
  token = "XXXX"
  num_retries = 3
  retry_delay = 30
}
```
## Argument Reference
* `token` - **(Required, String)** The token that will invoke all SonarCloud API commands. Token Authentication. Can be specified via env variable `SONARCLOUD_TOKEN`.
* `num_retries` - **(Optional, Integer)** Number of retries for each SonarCloud API call in case of 429-Too Many Requests or any 5XX status code. Can be specified via env variable `SONARCLOUD_NUM_RETRIES`. Default: 3.
* `retry_delay` - **(Optional, Integer)** How long to wait (in seconds) in between retries. Can be specified via env variable `SONARCLOUD_RETRY_DELAY`. Default: 30.
