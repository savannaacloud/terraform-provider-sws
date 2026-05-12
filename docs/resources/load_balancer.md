---
page_title: "sws_load_balancer Resource - terraform-provider-sws"
description: |-
  An Octavia load balancer.
---

# sws_load_balancer

An Octavia load balancer.

## Example

```hcl
resource "sws_load_balancer" "web" {
  name          = "web-lb"
  vip_subnet_id = sws_network.app.subnet_id
}
```

## Argument Reference

### Required

- `name` (string) — Display name.
- `vip_subnet_id` (string) — Subnet to allocate the VIP on.

### Optional

- `description` (string) — Free-form description.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — LB UUID.
- `vip_address` (string) — Allocated VIP.
- `status` (string) — ACTIVE / PENDING_CREATE / ERROR / etc.


## Import

```
terraform import sws_load_balancer.<local_name> <id>
```
