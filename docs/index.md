---
page_title: "Savannaa Provider"
description: |-
  Terraform provider for Savannaa Cloud — manage compute, networking, storage,
  load balancers, DNS, Kubernetes, managed databases, object storage, and
  long-tail managed services via the Savannaa API.
---

# Savannaa Provider

Provision and manage [Savannaa Cloud](https://savannaa.com) infrastructure
with Terraform / OpenTofu. The provider talks to the Savannaa public API
(`https://savannaa.com/api/...`) using your account's API key.

## Quick start

```hcl
terraform {
  required_providers {
    sws = {
      source  = "savannaacloud/sws"
      version = "~> 0.4"
    }
  }
}

provider "sws" {
  # Reads SWS_API_URL / SWS_API_KEY / SWS_PROJECT_NAME / SWS_REGION
  # from the environment by default. Or set them explicitly:
  api_url      = "https://savannaa.com"
  api_key      = var.sws_api_key
  project_name = "user-yourname"
  region       = "ng-abuja-1"
}

data "sws_image" "ubuntu" { name = "Ubuntu 22.04 LTS" }
data "sws_plan"  "small"  { name = "m1.small" }

resource "sws_keypair" "admin" {
  name       = "admin"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "sws_network" "app" {
  name = "app-net"
  # An inline /24 subnet is created automatically. Override with `cidr`,
  # or set `cidr = ""` to skip and add subnets explicitly via sws_subnet.
}

resource "sws_instance" "web" {
  name       = "web-01"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.app.id
  keypair    = sws_keypair.admin.name
  public_ip  = true   # allocates + associates a floating IP after ACTIVE
}

output "ip" { value = sws_instance.web.ip_address }
```

## Authentication

Generate an API key in the console at **Account → API Keys** (`ctk_…`).
Then either set environment variables:

```bash
export SWS_API_URL="https://savannaa.com"
export SWS_API_KEY="ctk_…"            # full key, not just prefix
export SWS_PROJECT_NAME="user-yourname"
export SWS_REGION="ng-abuja-1"        # or ng-lagos-1
```

Or pass them in the `provider` block (never commit the key to git).

## Schema

### Optional

- `api_url` (String) — Savannaa API base URL. Default `https://savannaa.com` or env `SWS_API_URL`.
- `api_key` (String, Sensitive) — API key from **Account → API Keys**. Default env `SWS_API_KEY`.
- `project_name` (String) — Project name. Default env `SWS_PROJECT_NAME`.
- `region` (String) — `ng-abuja-1` or `ng-lagos-1`. Default `ng-abuja-1` or env `SWS_REGION`.

## What's in v0.4

37 resources + 4 data sources covering most of the Savannaa platform.

**Core IaaS:** `sws_instance`, `sws_keypair`, `sws_network`, `sws_subnet`,
`sws_router`, `sws_router_interface`, `sws_floating_ip`, `sws_security_group`,
`sws_security_group_rule`, `sws_volume`, `sws_volume_attachment`,
`sws_volume_snapshot`.

**Load balancing (Octavia):** `sws_load_balancer`, `sws_lb_listener`,
`sws_lb_pool`, `sws_lb_member`, `sws_lb_health_monitor`.

**DNS (Designate):** `sws_dns_zone`, `sws_dns_record`, `sws_private_dns_zone`.

**Managed services:** `sws_object_bucket`, `sws_kubernetes_template`,
`sws_kubernetes_cluster`, `sws_managed_database`, `sws_serverless_container`,
`sws_file_storage`, `sws_cache`, `sws_queue`, `sws_kafka`, `sws_logging`,
`sws_cdn`, `sws_notification`, `sws_pipeline`, `sws_registry`,
`sws_backup_policy`, `sws_bastion`, `sws_vpc_peering`, `sws_vault_secret`,
`sws_alarm`, `sws_tag`.

**Data sources (lookup by name):** `sws_image`, `sws_plan`, `sws_network`,
`sws_security_group`.

See the sidebar for each resource's full schema and examples.
