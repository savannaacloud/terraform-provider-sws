---
page_title: "sws_router Resource - terraform-provider-sws"
description: |-
  A Neutron router with auto-discovered external gateway.
---

# sws_router

A Neutron router with auto-discovered external gateway.

## Example

```hcl
resource "sws_router" "main" {
  name = "main-router"
}
```

## Argument Reference

### Required

- `name` (string) — Display name.

### Optional

- `description` (string) — Free-form description.
- `external_network_id` (string) — External network UUID. If omitted, the project's default external network is auto-discovered.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Router UUID.


## Import

```
terraform import sws_router.<local_name> <id>
```
