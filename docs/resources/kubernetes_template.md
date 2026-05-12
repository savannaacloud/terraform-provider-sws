---
page_title: "sws_kubernetes_template Resource - terraform-provider-sws"
description: |-
  A Magnum cluster template (reusable blueprint).
---

# sws_kubernetes_template

A Magnum cluster template (reusable blueprint).

## Example

```hcl
resource "sws_kubernetes_template" "k8s28" {
  name                = "k8s-1.28"
  image               = data.sws_image.fcos.id
  keypair_id          = sws_keypair.admin.name
  external_network_id = "<ext-net-uuid>"
  master_flavor_id    = data.sws_plan.medium.id
  flavor_id           = data.sws_plan.small.id
}
```

## Argument Reference

### Required

- `name` (string) — Template name.
- `image` (string) — Fedora CoreOS image UUID.
- `keypair_id` (string) — Keypair for cluster nodes.
- `external_network_id` (string) — External network UUID for the cluster's router.
- `master_flavor_id` (string) — Master node flavor.
- `flavor_id` (string) — Worker node flavor.

### Optional

- `dns_nameserver` (string) — Default `8.8.8.8`.
- `coe_name` (string) — `kubernetes` (default), `swarm`, `mesos`.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Template UUID.


## Import

```
terraform import sws_kubernetes_template.<local_name> <id>
```
