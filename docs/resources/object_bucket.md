---
page_title: "sws_object_bucket Resource - terraform-provider-sws"
description: |-
  An object-storage bucket (Ceph RGW / Swift container).
---

# sws_object_bucket

An object-storage bucket (Ceph RGW / Swift container).

## Example

```hcl
resource "sws_object_bucket" "logs" {
  name = "my-app-logs"
}
```

## Argument Reference

### Required

- `name` (string) — Bucket name. Used as the ID.

### Optional

_None._


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Same as name.


## Import

```
terraform import sws_object_bucket.<local_name> <id>
```
