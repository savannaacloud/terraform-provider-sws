---
page_title: "sws_volume_snapshot Resource - terraform-provider-sws"
description: |-
  A point-in-time snapshot of an sws_volume.
---

# sws_volume_snapshot

A point-in-time snapshot of an sws_volume.

## Example

```hcl
resource "sws_volume_snapshot" "nightly" {
  name      = "nightly-2026-05-11"
  volume_id = sws_volume.data.id
}
```

## Argument Reference

### Required

- `name` (string) — Snapshot name.
- `volume_id` (string) — Source volume UUID.

### Optional

- `description` (string) — Free-form description.
- `force` (bool) — Snapshot a volume that is `in-use`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Snapshot UUID.
- `status` (string) — `creating` / `available` / `error` / etc.


## Import

```
terraform import sws_volume_snapshot.<local_name> <id>
```
