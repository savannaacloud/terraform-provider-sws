---
page_title: "sws_pipeline Resource - terraform-provider-sws"
description: |-
  A managed data pipeline.
---

# sws_pipeline

A managed data pipeline.

## Example

```hcl
resource "sws_pipeline" "etl" { name = "etl" }
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
terraform import sws_pipeline.<local_name> <id>
```
