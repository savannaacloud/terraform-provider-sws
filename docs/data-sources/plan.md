---
page_title: "sws_plan Data Source - terraform-provider-sws"
description: |-
  Look up a compute plan (flavor) by name.
---

# sws_plan (Data Source)

Look up a compute plan (flavor) by name.

## Example

```hcl
data "sws_plan" "small" {
  name = "m1.small"
}
```

## Argument Reference

### Required

- `name` (string) — Plan name, e.g. `m1.small`.


## Attribute Reference

### Read-only

- `id` (string) — Plan UUID.
- `vcpus` (int) — vCPU count.
- `ram_mb` (int) — Memory in MiB.
- `disk_gb` (int) — Root disk in GiB.

