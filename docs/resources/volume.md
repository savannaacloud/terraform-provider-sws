---
page_title: "sws_volume Resource - terraform-provider-sws"
description: |-
  A Cinder block-storage volume (Ceph RBD-backed on Savannaa).
---

# sws_volume

A Cinder block-storage volume (Ceph RBD-backed on Savannaa).

## Example

```hcl
resource "sws_volume" "data" {
  name = "data"
  size = 100
}
```

## Argument Reference

### Required

- `name` (string) — Display name.
- `size` (int) — Size in GiB.

### Optional

- `description` (string) — Free-form description.
- `volume_type` (string) — Volume type. Defaults to the project's default.
- `availability_zone` (string) — AZ to place the volume in.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Volume UUID.
- `status` (string) — `creating` / `available` / `in-use` / `error` / etc.


## Import

```
terraform import sws_volume.<local_name> <id>
```
