---
page_title: "sws_cdn Resource - terraform-provider-sws"
description: |-
  A managed CDN distribution.
---

# sws_cdn

A managed CDN distribution.

## Example

```hcl
resource "sws_cdn" "site" { name = "site-cdn" }
```

## Argument Reference

### Required

- `name` (string) — Distribution name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Distribution UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_cdn.<local_name> <id>
```
