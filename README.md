# terraform-provider-sws

Terraform provider for [Savannaa Cloud](https://savannaa.com).

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
  # Reads SWS_API_URL, SWS_API_KEY, SWS_PROJECT_NAME, SWS_REGION
  # from the environment by default.
}

data "sws_image" "ubuntu" {
  name = "Ubuntu 22.04"
}

data "sws_plan" "small" {
  name = "m1.small"
}

resource "sws_keypair" "admin" {
  name       = "admin"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "sws_network" "app" {
  name = "app-net"
}

resource "sws_instance" "web" {
  name       = "web-01"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.app.id
  keypair    = sws_keypair.admin.name
  public_ip  = true
}

output "ip" {
  value = sws_instance.web.ip_address
}
```

## Auth

Generate an API key in the Savannaa console (**Account → API Keys**).
Then either set env vars:

```bash
export SWS_API_URL="https://api.savannaa.com/v3"
export SWS_API_KEY="ctk_..."          # full key, not just the prefix
export SWS_PROJECT_NAME="user-yourname"
export SWS_REGION="ng-abuja-1"
```

Or put them in the `provider` block (do NOT commit `api_key` to git):

```hcl
provider "sws" {
  api_url      = "https://api.savannaa.com/v3"
  api_key      = var.sws_api_key
  project_name = "user-yourname"
  region       = "ng-abuja-1"
}
```

## Resources

| Resource | Status |
|----------|--------|
| `sws_instance` | ✅ |
| `sws_network` | ✅ |
| `sws_security_group` | ✅ |
| `sws_keypair` | ✅ |
| `sws_subnet` | 🚧 v0.2 |
| `sws_floating_ip` | 🚧 v0.2 |
| `sws_volume` | 🚧 v0.2 |
| `sws_security_group_rule` | 🚧 v0.2 |

## Data sources

| Data source | Status |
|-------------|--------|
| `sws_image` | ✅ |
| `sws_plan` | ✅ |

## Building locally

```bash
go build -o terraform-provider-sws
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/savannaacloud/sws/0.1.0/$(go env GOOS)_$(go env GOARCH)
mv terraform-provider-sws ~/.terraform.d/plugins/registry.terraform.io/savannaacloud/sws/0.1.0/$(go env GOOS)_$(go env GOARCH)/
```

## Releasing

1. Tag: `git tag v0.1.0 && git push origin v0.1.0`
2. GitHub Actions runs goreleaser, builds platforms, signs with GPG, uploads to a GitHub Release.
3. Terraform Registry detects the release within ~5 min (once the provider is published).

## License

MPL-2.0
