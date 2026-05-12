---
page_title: "sws_lb_pool Resource - terraform-provider-sws"
description: |-
  An Octavia backend pool.
---

# sws_lb_pool

An Octavia backend pool.

## Example

```hcl
resource "sws_lb_pool" "web" {
  load_balancer_id = sws_load_balancer.web.id
  name             = "web-pool"
  protocol         = "HTTP"
  lb_algorithm     = "ROUND_ROBIN"
}
```

## Argument Reference

### Required

- `load_balancer_id` (string) — Parent LB UUID.
- `name` (string) — Display name.
- `protocol` (string) — `TCP`, `HTTP`, `HTTPS`, `PROXY`.

### Optional

- `lb_algorithm` (string) — `ROUND_ROBIN` (default), `LEAST_CONNECTIONS`, `SOURCE_IP`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Pool UUID.


## Import

```
terraform import sws_lb_pool.<local_name> <id>
```
