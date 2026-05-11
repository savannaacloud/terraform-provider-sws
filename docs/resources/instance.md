---
page_title: "sws_instance Resource - terraform-provider-sws"
description: |-
  Manages a Savannaa compute instance.
---

# sws_instance

A Savannaa compute instance (virtual machine).

## Example

```hcl
data "sws_image" "ubuntu" { name = "Ubuntu 22.04" }
data "sws_plan"  "small"  { name = "m1.small" }

resource "sws_instance" "web" {
  name      = "web-01"
  plan      = data.sws_plan.small.name
  image     = data.sws_image.ubuntu.id
  keypair   = sws_keypair.admin.name
  public_ip = true
}
```

## Schema

### Required

- `name` — display name
- `plan` — plan/flavor name (use `sws_plan` data source)
- `image` — image UUID (use `sws_image` data source)

### Optional

- `network_id` — network UUID, defaults to the project's default network
- `keypair` — `sws_keypair` name to inject for SSH
- `public_ip` — allocate a public IP at create. Default `true`

### Read-only

- `id` — server UUID
- `ip_address` — public IP if allocated, else primary fixed IP
- `status` — `BUILD` / `ACTIVE` / `ERROR` / etc.

## Import

```
terraform import sws_instance.web <uuid>
```
