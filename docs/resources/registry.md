---
page_title: "sws_registry Resource - terraform-provider-sws"
description: |-
  A managed container registry.
---

# sws_registry

A managed container registry.

## Example

```hcl
resource "sws_registry" "private" { name = "private" }
```

## Argument Reference

### Required

- `name` (string) — Registry name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Registry UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_registry.<local_name> <id>
```
