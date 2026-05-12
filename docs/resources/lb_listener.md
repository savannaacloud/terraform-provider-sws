---
page_title: "sws_lb_listener Resource - terraform-provider-sws"
description: |-
  An Octavia listener on a load balancer.
---

# sws_lb_listener

An Octavia listener on a load balancer.

## Example

```hcl
resource "sws_lb_listener" "https" {
  load_balancer_id = sws_load_balancer.web.id
  name             = "https"
  protocol         = "HTTPS"
  protocol_port    = 443
  default_pool_id  = sws_lb_pool.web.id
}
```

## Argument Reference

### Required

- `load_balancer_id` (string) — Parent LB UUID.
- `name` (string) — Display name.
- `protocol` (string) — `TCP`, `HTTP`, `HTTPS`, `TERMINATED_HTTPS`.
- `protocol_port` (int) — Listener port.

### Optional

- `default_pool_id` (string) — Default backend pool UUID.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Listener UUID.


## Import

```
terraform import sws_lb_listener.<local_name> <id>
```
