# bosh-vmrun-cpi-release

[BOSH CPI](https://bosh.io/docs/cpi-api-v1/) for VMWare Workstation/Fusion using `vmrun` and related binaries

## Releases

The software is under very active development and there there is currently no published releases on [bosh.io](bosh.io/releases).  See the usage section for instructions to build a dev release using the bosh-cli.

## Pre-requisites

* Linux or MacOS host
* VMware Fusion or Workstation installed (tested against on Fusion 8 and Workstation 14)
* Linux or Windows Stemcell for vsphere
    * Linux stemcells are at [bosh.io/stemcells](https://bosh.io/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent)
    * Windows stemcells must be manually built using [bosh-windows-stemcell-builder](https://github.com/cloudfoundry-incubator/bosh-windows-stemcell-builder) due to Microsoft licensing restrictions.

## Limitations

* CPI can only be used for stand-alone VMs created with `bosh create-env`, not for full bosh directors nor deployments
  * `vmrun` can only communicate with the hypervisor when run on the host machine

## Usage

### Fusion/Workstation setup

* Network configured for NAT
    * Fusion
      * VMware Fusion Menu -> Preferences -> Network
      * Create/choose a network with these settings
        * [x] Allow virtual machines on this network to connect to external networks (using NAT)
        * [x] Connect this host Mac to this network
        * [x] Provide addresses on this network via DHCP
        * Choose a specific subnet range (ex: 10.0.0.0/255.255.255.0)
    * Workstation
       * Edit -> Virtual Network Editor
       * Create/choose a network with these settings
         * [x] NAT (share host's IP address with VMs)
         * [x] Use local DHCP service to distribute IP addresses to VMs
         * [x] Connect a host virtual adapter ([your vm network name]) to this network
         * Choose a specific subnet range (ex: 10.0.0.0/255.255.255.0)
* Find the paths for these binaries: `vmrun`, `ovftool`, and `vmware-vdiskmanager`
  * Workstation typically has them on the `PATH` already
  * Fusion includes all under:
    * `/Applications/VMware Fusion.app/Contents/Library/`
    * `/Applications/VMware Fusion.app/Contents/Library/VMware OVF Tool/`

### Example deployment

```
# vendor blobs for golang-release
git clone https://github.com/bosh-packages/golang-release ./state/golang-release
bosh vendor-package --dir ./ golang-1.9-linux ./state/golang-release
bosh vendor-package --dir ./ golang-1.9-darwin ./state/golang-release

# create dev release
bosh create-release --sha2 --force --dir ./ --tarball ./state/cpi.tgz

# download a stemcell
curl -L "https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-xenial-go_agent?v=97.10" > ./state/stemcell.tgz

# create the vm <example manifest below>
bosh create-env my-vm.yml \
  --vars-store ./state/vm-creds.yml \
  --state ./state/vm_state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v stemcell_url=file://$PWD/state/stemcell.tgz \
  -v stemcell_sha1=$(shasum -a1 < state/stemcell.tgz | awk '{print $1}') \
  -v vmrun_bin_path="$(which vmrun)" \
  -v ovftool_bin_path="$(which ovftool)" \
  -v vdiskmanager_bin_path="$(which vmware-vdiskmanager)" \
  -v vm_store_path="/tmp/vm-store-path" \
  -v internal_ip=10.0.0.5  \
  -v internal_cidr=255.255.255.0 \
  -v internal_gw=10.0.0.2 \
  -v network_name=vmnet3 \
  ;
```

#### Example manifest

```
---
name: my-vm

releases:
- name: bosh-vmrun-cpi
  url: ((cpi_url))

resource_pools:
- name: vms
  network: default
  stemcell:
    url: ((stemcell_url))
    sha1: ((stemcell_sha1))
  cloud_properties:
    cpu: 2
    ram: 4_096
    disk: 40_000

    # optional bootstrap script, runs before bosh-agent starts
    bootstrap:
      script_content: |
        # add your own bootstrap actions here
      script_path: '/tmp/bootstrap.sh'
      interpreter_path: '/bin/bash'
      username: 'vcap'
      password: 'c1oudc0w' #same as stemcell
  env:
    bosh:
      mbus:
        cert: ((mbus_bootstrap_ssl))

disk_pools:
- name: disks
  disk_size: 65_536

networks:
- name: default
  type: manual
  subnets:
  - range: ((internal_cidr))
    gateway: ((internal_gw))
    static: [((internal_ip))]
    dns: [8.8.8.8]
    cloud_properties:
      name: ((network_name))

instance_groups:
- name: my-vm
  instances: 1
  jobs: []
  resource_pool: vms
  persistent_disk_pool: disks
  networks:
  - name: default
    static_ips: [((internal_ip))]
  properties:
    agent:
      env:
        bosh: {}
    ntp: &ntp
    - time1.google.com
    - time2.google.com
    - time3.google.com
    - time4.google.com
cloud_provider:
  mbus: https://mbus:((mbus_bootstrap_password))@((internal_ip)):6868
  cert: ((mbus_bootstrap_ssl))
  properties:
    blobstore:
      provider: local
      path: '/var/vcap/micro_bosh/data/cache' #depends on agents internal location
    agent: {mbus: "https://mbus:((mbus_bootstrap_password))@0.0.0.0:6868"}
    ntp: *ntp
    vmrun:
      vmrun_bin_path: "((vmrun_bin_path))"
      ovftool_bin_path: "((ovftool_bin_path))"
      vdiskmanager_bin_path: "((vdiskmanager_bin_path))"
      vm_store_path: "((vm_store_path))"
  template:
    name: vmrun_cpi
    release: bosh-vmrun-cpi

variables:
- name: mbus_bootstrap_password
  type: password

- name: default_ca
  type: certificate
  options:
    is_ca: true
    common_name: ca

- name: mbus_bootstrap_ssl
  type: certificate
  options:
    ca: default_ca
    common_name: ((internal_ip))
    alternative_names: [((internal_ip))]
```

