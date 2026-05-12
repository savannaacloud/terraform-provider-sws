---
page_title: "sws_instance Resource - terraform-provider-sws"
description: |-
  A compute instance (VM).
---

# sws_instance

A compute instance (VM).

## Example

```hcl
resource "sws_instance" "web" {
  name       = "web-01"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.app.id
  keypair    = sws_keypair.admin.name
  public_ip  = true
}
```

## Argument Reference

### Required

- `name` (string) — Display name.
- `plan` (string) — Flavor name (e.g. `m1.small`) or UUID. The provider resolves names to UUIDs automatically.
- `image` (string) — Image UUID. Use `data.sws_image.<name>.id`.

### Optional

- `network_id` (string) — Network UUID. Defaults to the project's default network.
- `keypair` (string) — Keypair name to inject for SSH access.
- `public_ip` (bool) — Allocate + associate a public IP after ACTIVE. Default `false`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Instance UUID.
- `ip_address` (string) — Public IP if allocated, else fixed IP.
- `status` (string) — BUILD / ACTIVE / ERROR / etc.


## Import

```
terraform import sws_instance.<local_name> <id>
```
