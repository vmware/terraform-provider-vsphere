terraform {
  required_providers {
    vsphere = {
      source = "vmware/vsphere"
    }
  }
}

provider "vsphere" {
  user                 = var.vcenter_username
  password             = var.vcenter_password
  vsphere_server       = var.vcenter_server
  allow_unverified_ssl = true
}

data "vsphere_host_thumbprint" "thumbprint1" {
  address  = var.hosts[0].hostname
  insecure = true
}

data "vsphere_host_thumbprint" "thumbprint2" {
  address  = var.hosts[1].hostname
  insecure = true
}

data "vsphere_host_thumbprint" "thumbprint3" {
  address  = var.hosts[2].hostname
  insecure = true
}

resource "vsphere_datacenter" "dc" {
  name = "acc-test-dc"
}

resource "vsphere_host" "host1" {
  datacenter = vsphere_datacenter.dc.moid
  hostname   = var.hosts[0].hostname
  username   = var.hosts[0].username
  password   = var.hosts[0].password
  thumbprint = data.vsphere_host_thumbprint.thumbprint1.id
}

resource "vsphere_host" "host2" {
  datacenter = vsphere_datacenter.dc.moid
  hostname   = var.hosts[1].hostname
  username   = var.hosts[1].username
  password   = var.hosts[1].password
  thumbprint = data.vsphere_host_thumbprint.thumbprint2.id
}

resource "vsphere_host" "host3" {
  datacenter = vsphere_datacenter.dc.moid
  hostname   = var.hosts[2].hostname
  username   = var.hosts[2].username
  password   = var.hosts[2].password
  thumbprint = data.vsphere_host_thumbprint.thumbprint3.id
}

resource "vsphere_compute_cluster" "cluster" {
  datacenter_id = vsphere_datacenter.dc.moid
  name          = "acc-test-cluster"

  ha_enabled = true
  drs_enabled = true

  host_system_ids = [
    vsphere_host.host1.id,
    vsphere_host.host2.id,
  ]
}

resource "vsphere_nas_datastore" "nfs" {
  name            = "acc-test-nfs"
  host_system_ids = [vsphere_host.host1.id, vsphere_host.host2.id]
  type            = "NFS"
  remote_hosts    = [var.nfs_host]
  remote_path     = "/store"
}