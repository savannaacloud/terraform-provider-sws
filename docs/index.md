---
page_title: "Savannaa Provider"
description: |-
  Terraform provider for Savannaa Cloud — manage compute, networks, security groups, keypairs, and more via the Savannaa API.
---

# Savannaa Provider

Provision and manage [Savannaa Cloud](https://savannaa.com) infrastructure with Terraform / OpenTofu.

## Example

```hcl
terraform {
  required_providers {
    sws = {
      source  = "savannaacloud/sws"
      version = "~> 0.1"
    }
  }
}

provider "sws" {
  # api_url / api_key / project_name / region read from env by default
}
```

## Schema

### Optional

- `api_url` (String) — Savannaa API base URL. Default `https://api.savannaa.com/v3` or env `SWS_API_URL`.
- `api_key` (String, Sensitive) — API key from **Account → API Keys** (starts with `ctk_`). Default env `SWS_API_KEY`.
- `project_name` (String) — Project name. Default env `SWS_PROJECT_NAME`.
- `region` (String) — `ng-abuja-1` or `ng-lagos-1`. Default `ng-abuja-1` or env `SWS_REGION`.
