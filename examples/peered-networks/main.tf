###############################################################################
#  Two networks peered together.
#
#  The simplest VPC-peering shape: net-a ↔ net-b, with one Ubuntu instance
#  in each side so you can quickly verify connectivity end-to-end.
###############################################################################

terraform {
  required_providers {
    sws = {
      source  = "savannaacloud/sws"
      version = "~> 0.4"
    }
  }
}

provider "sws" {}

variable "ssh_public_key" {
  description = "Public key to inject for SSH into the test VMs."
  type        = string
  default     = ""
}

# ── Networks ──────────────────────────────────────────────────────────────
resource "sws_network" "a" {
  name = "net-a"
  cidr = "10.30.1.0/24"
}

resource "sws_network" "b" {
  name = "net-b"
  cidr = "10.30.2.0/24"
}

# ── Peering ───────────────────────────────────────────────────────────────
resource "sws_vpc_peering" "a_to_b" {
  name = "peering-a-to-b"
  config = jsonencode({
    local_network_id = sws_network.a.id
    peer_network_id  = sws_network.b.id
  })
}

# ── Test instances (optional) ─────────────────────────────────────────────
data "sws_image" "ubuntu" { name = "Ubuntu 22.04 LTS" }
data "sws_plan"  "small"  { name = "m1.small" }

resource "sws_keypair" "test" {
  count      = var.ssh_public_key == "" ? 0 : 1
  name       = "peering-test-key"
  public_key = var.ssh_public_key
}

resource "sws_security_group" "test" {
  name        = "sg-peering-test"
  description = "Allow ping + SSH between the two networks"
}

resource "sws_security_group_rule" "icmp" {
  security_group_id = sws_security_group.test.id
  direction         = "ingress"
  protocol          = "icmp"
  remote_ip_prefix  = "10.30.0.0/16"   # both nets
  description       = "Allow ICMP between net-a and net-b"
}

resource "sws_security_group_rule" "ssh" {
  security_group_id = sws_security_group.test.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  description       = "SSH management"
}

resource "sws_instance" "vm_a" {
  name       = "vm-a"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.a.id
  keypair    = length(sws_keypair.test) > 0 ? sws_keypair.test[0].name : null
  public_ip  = true
}

resource "sws_instance" "vm_b" {
  name       = "vm-b"
  plan       = data.sws_plan.small.name
  image      = data.sws_image.ubuntu.id
  network_id = sws_network.b.id
  keypair    = length(sws_keypair.test) > 0 ? sws_keypair.test[0].name : null
  public_ip  = true
}

# ── Outputs ───────────────────────────────────────────────────────────────
output "network_a_id" { value = sws_network.a.id }
output "network_b_id" { value = sws_network.b.id }
output "peering_id"   { value = sws_vpc_peering.a_to_b.id }
output "vm_a_ip"      { value = sws_instance.vm_a.ip_address }
output "vm_b_ip"      { value = sws_instance.vm_b.ip_address }
