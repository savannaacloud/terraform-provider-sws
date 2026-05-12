---
page_title: "sws_serverless_container Resource - terraform-provider-sws"
description: |-
  A Zun (serverless) container.
---

# sws_serverless_container

A Zun (serverless) container.

## Example

```hcl
resource "sws_serverless_container" "api" {
  name  = "api"
  image = "myorg/api:latest"
}
```

## Argument Reference

### Required

- `name` (string) — Container name.
- `image` (string) — Docker image reference.

### Optional

- `network_id` (string) — Network to attach to.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Container UUID.
- `status` (string) — `Creating` / `Running` / `Stopped`.
- `address` (string) — Container IP.


## Import

```
terraform import sws_serverless_container.<local_name> <id>
```
