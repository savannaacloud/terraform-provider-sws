---
page_title: "sws_lb_health_monitor Resource - terraform-provider-sws"
description: |-
  An Octavia health monitor for a pool. One monitor per pool.
---

# sws_lb_health_monitor

An Octavia health monitor for a pool. One monitor per pool.

## Example

```hcl
resource "sws_lb_health_monitor" "web" {
  pool_id     = sws_lb_pool.web.id
  type        = "HTTP"
  delay       = 5
  timeout     = 3
  max_retries = 3
  url_path    = "/healthz"
}
```

## Argument Reference

### Required

- `pool_id` (string) — Parent pool UUID.
- `type` (string) — `TCP`, `HTTP`, `HTTPS`, `PING`.
- `delay` (int) — Seconds between probes.
- `timeout` (int) — Probe timeout in seconds.
- `max_retries` (int) — Failures before marking member down.

### Optional

- `url_path` (string) — HTTP probe path. Default `/`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Health monitor UUID.


## Import

```
terraform import sws_lb_health_monitor.<local_name> <id>
```
