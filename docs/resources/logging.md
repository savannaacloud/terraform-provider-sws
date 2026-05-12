---
page_title: "sws_logging Resource - terraform-provider-sws"
description: |-
  A managed log aggregation pipeline.
---

# sws_logging

A managed log aggregation pipeline.

## Example

```hcl
resource "sws_logging" "app" { name = "app-logs" }
```

## Argument Reference

### Required

- `name` (string) — Pipeline name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Pipeline UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_logging.<local_name> <id>
```
