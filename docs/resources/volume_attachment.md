---
page_title: "sws_volume_attachment Resource - terraform-provider-sws"
description: |-
  Attaches an sws_volume to an sws_instance.
---

# sws_volume_attachment

Attaches an sws_volume to an sws_instance.

## Example

```hcl
resource "sws_volume_attachment" "data" {
  instance_id = sws_instance.web.id
  volume_id   = sws_volume.data.id
}
```

## Argument Reference

### Required

- `instance_id` (string) — Instance UUID.
- `volume_id` (string) — Volume UUID.

### Optional

- `device` (string) — Guest device name, e.g. `/dev/vdb`. Auto-assigned if omitted.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Composite `<instance_id>:<volume_id>`.


## Import

```
terraform import sws_volume_attachment.<local_name> <id>
```

For composite-ID resources, the import id is shown in the Read-only `id` attribute (e.g. `<router_id>:<subnet_id>`).
