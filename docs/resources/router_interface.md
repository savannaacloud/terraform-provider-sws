---
page_title: "sws_router_interface Resource - terraform-provider-sws"
description: |-
  Attaches a subnet to a router (AWS subnet → route-table equivalent).
---

# sws_router_interface

Attaches a subnet to a router (AWS subnet → route-table equivalent).

## Example

```hcl
resource "sws_router_interface" "app" {
  router_id = sws_router.main.id
  subnet_id = sws_network.app.subnet_id
}
```

## Argument Reference

### Required

- `router_id` (string) — Router UUID.
- `subnet_id` (string) — Subnet UUID.

### Optional

_None._


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Composite `<router_id>:<subnet_id>`.


## Import

```
terraform import sws_router_interface.<local_name> <id>
```

For composite-ID resources, the import id is shown in the Read-only `id` attribute (e.g. `<router_id>:<subnet_id>`).
