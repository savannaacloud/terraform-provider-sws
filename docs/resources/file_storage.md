---
page_title: "sws_file_storage Resource - terraform-provider-sws"
description: |-
  A managed file storage share (NFS-style).
---

# sws_file_storage

A managed file storage share (NFS-style).

## Example

```hcl
resource "sws_file_storage" "shared" {
  name = "shared"
}
```

## Argument Reference

### Required

- `name` (string) — Share name.

### Optional

- `config` (string) — JSON-encoded service-specific config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Share UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_file_storage.<local_name> <id>
```
