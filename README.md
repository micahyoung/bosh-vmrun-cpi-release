# bosh-vmrun-cpi-release

[BOSH CPI](https://bosh.io/docs/cpi-api-v1/) for VMWare Workstation/Fusion using `vmrun` and related binaries

## Releases

The software is under very active development and there there is currently no published releases on [bosh.io](bosh.io/releases).  See the usage section for instructions to build a dev release using the bosh-cli.

You can find bre-built tarballs on the [releases](https://github.com/micahyoung/bosh-vmrun-cpi-release/releases) page.

## Pre-requisites

* VMware Fusion or Workstation installed on Linux, MacOS or Windows 10 (tested against on Fusion 8 and Workstation 14)
* Linux or Windows Stemcell for vsphere
    * Linux stemcells are at [bosh.io/stemcells](https://bosh.io/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent)
    * Windows stemcells must be manually built using [bosh-windows-stemcell-builder](https://github.com/cloudfoundry-incubator/bosh-windows-stemcell-builder) due to Microsoft licensing restrictions.

## Deployment scenarios

* Stand-alone VMs created with `bosh create-env`
  * `bosh` CLI communicates directly with the hypervisor on the host machine
* BOSH director via `bosh create-env` and deployments via `bosh deploy` over SSH tunnel
  * SSH server must be running on `vmrun` hypervisor with public key enabled
  * OS-specific CPI binary will be installed on the hypervisor
  * CPI must run as user with privileges to execute `vmrun`

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
# create the vm <example manifest below>
bosh create-env my-vm.yml \
  --vars-store ./state/vm-creds.yml \
  --state ./state/vm_state.json \
  -v cpi_url="https://github.com/micahyoung/bosh-vmrun-cpi-release/releases/download/v1.0.2/bosh-vmrun-cpi-release-1.0.2.tgz" \
  -v cpi_sha1="893f13f2f8084838092f7e095634c09c1b959096" \
  -v stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-xenial-go_agent?v=97.10" \
  -v stemcell_sha1="9e832921a4a1279b8029b72f962abcb2f981b32c" \
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
  sha1: ((cpi_sha1))

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
      script_path: '/tmp/bootstrap.sh'  # full path where to write temporary script on VM
      interpreter_path: '/bin/bash'     # full path to script interpreter
      username: 'vcap'                  # VM username with permissions to run the script
      password: 'c1oudc0w'              # hard-coded stemcell password (eventually changed by bosh-agent)
      ready_process_name: 'bosh-agent'  # process name that the script should wait for before running
      min_wait_seconds: 180             # time to sleep before polling for ready_process_name (defaults to 0)
      max_wait_seconds: 300             # max time to wait for ready_process_name - fails if exceeded  env:
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
      vm_start_max_wait_seconds: 600
      vm_soft_shutdown_max_wait_seconds: 30
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

