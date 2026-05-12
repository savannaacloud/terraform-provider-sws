---
page_title: "sws_alarm Resource - terraform-provider-sws"
description: |-
  An Aodh alarm on a metric.
---

# sws_alarm

An Aodh alarm on a metric.

## Example

```hcl
resource "sws_alarm" "high_cpu" { name = "high-cpu" }
```

## Argument Reference

### Required

- `name` (string) — Alarm name.

### Optional

- `config` (string) — JSON-encoded threshold + action.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Alarm UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_alarm.<local_name> <id>
```
