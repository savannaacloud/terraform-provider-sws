---
page_title: "sws_managed_database Resource - terraform-provider-sws"
description: |-
  A managed database (Trove). Polls ACTIVE (3-6 min).
---

# sws_managed_database

A managed database (Trove). Polls ACTIVE (3-6 min).

## Example

```hcl
resource "sws_managed_database" "app" {
  name       = "app-db"
  datastore  = "mysql"
  version    = "8.0"
  flavor_id  = data.sws_plan.small.id
  size       = 50
  network_id = sws_network.app.id
}
```

## Argument Reference

### Required

- `name` (string) — Instance name.
- `datastore` (string) — `mysql`, `postgres`, `mariadb`.
- `version` (string) — Engine version, e.g. `8.0`.
- `flavor_id` (string) — Flavor UUID.
- `size` (int) — Volume size in GiB.
- `network_id` (string) — Network UUID.

### Optional

- `root_enabled` (bool) — Enable root password after creation.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Instance UUID.
- `status` (string) — `BUILD` / `ACTIVE` / `ERROR`.
- `address` (string) — Instance IP.


## Import

```
terraform import sws_managed_database.<local_name> <id>
```
