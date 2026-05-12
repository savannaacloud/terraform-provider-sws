---
page_title: "sws_tag Resource - terraform-provider-sws"
description: |-
  A free-form tag attached to project resources.
---

# sws_tag

A free-form tag attached to project resources.

## Example

```hcl
resource "sws_tag" "env" { name = "env=prod" }
```

## Argument Reference

### Required

- `name` (string) — Tag value.

### Optional

- `config` (string) — JSON-encoded scope.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Tag UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_tag.<local_name> <id>
```
