---
page_title: "sws_lb_member Resource - terraform-provider-sws"
description: |-
  A backend member in an LB pool.
---

# sws_lb_member

A backend member in an LB pool.

## Example

```hcl
resource "sws_lb_member" "web1" {
  pool_id       = sws_lb_pool.web.id
  address       = sws_instance.web.ip_address
  protocol_port = 80
  subnet_id     = sws_network.app.subnet_id
}
```

## Argument Reference

### Required

- `pool_id` (string) — Parent pool UUID.
- `address` (string) — Member IP (typically an instance fixed IP).
- `protocol_port` (int) — Port on the member.
- `subnet_id` (string) — Subnet of the member IP.

### Optional

- `weight` (int) — Member weight. Default 1.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Composite `<pool_id>:<member_id>`.


## Import

```
terraform import sws_lb_member.<local_name> <id>
```

For composite-ID resources, the import id is shown in the Read-only `id` attribute (e.g. `<router_id>:<subnet_id>`).
