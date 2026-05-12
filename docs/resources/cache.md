---
page_title: "sws_cache Resource - terraform-provider-sws"
description: |-
  A managed Redis cache.
---

# sws_cache

A managed Redis cache.

## Example

```hcl
resource "sws_cache" "session" {
  name = "session-cache"
}
```

## Argument Reference

### Required

- `name` (string) — Cache name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Cache UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_cache.<local_name> <id>
```
