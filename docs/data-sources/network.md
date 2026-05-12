---
page_title: "sws_network Data Source - terraform-provider-sws"
description: |-
  Look up an existing network by name.
---

# sws_network (Data Source)

Look up an existing network by name.

## Example

```hcl
data "sws_network" "default" {
  name = "default"
}
```

## Argument Reference

### Required

- `name` (string) — Network name (exact match).


## Attribute Reference

### Read-only

- `id` (string) — Network UUID.

