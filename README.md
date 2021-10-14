# bosh-vmrun-cpi-release

[BOSH CPI](https://bosh.io/docs/cpi-api-v1/) for VMWare Workstation/Fusion Pro using `vmrun` and related binaries

## Releases

You can find bre-built tarballs on the [releases](https://github.com/micahyoung/bosh-vmrun-cpi-release/releases) page.

There are currently no published releases on [bosh.io](bosh.io/releases).

## Pre-requisites

* The following VMware vmrun hypervisors are known to work:
  * VMware Fusion 8.5 for MacOS
  * VMware Fusion 11.5 Pro for MacOS
  * VMware Workstation 14 for Windows
  * VMware Player 14.1.1 [download](https://download3.vmware.com/software/player/file/VMware-Player-14.1.1-7528167.x86_64.bundle) with VMware VIX 1.17.0 [download](https://download3.vmware.com/software/player/file/VMware-VIX-1.17.0-6661328.x86_64.bundle)
* Linux or Windows Stemcell for vsphere
    * Linux stemcells are at [bosh.io/stemcells](https://bosh.io/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent)
    * Windows stemcells must be manually built using [bosh-windows-stemcell-builder](https://github.com/cloudfoundry-incubator/bosh-windows-stemcell-builder) due to Microsoft licensing restrictions.

## Deployment scenarios

* Stand-alone VMs created with `bosh create-env`
  * `bosh` CLI communicates directly with the hypervisor on the host machine
* BOSH director via `bosh create-env` and deployments via `bosh deploy` over localhost SSH tunnel
  * SSH server must be running on `vmrun` hypervisor with public key enabled
  * OS-specific CPI binary will be installed on the hypervisor
  * CPI must run as user with privileges to execute `vmrun`

## Usage
Follow the instructions for your VMware product:

### Fusion setup
* Find the paths for these binaries: `vmrun` and`ovftool`
  * Fusion defaults:
    * `/Applications/VMware Fusion.app/Contents/Library/vmrun`
    * `/Applications/VMware Fusion.app/Contents/Library/VMware OVF Tool/ovftool`
* Network configured for NAT
   * VMware Fusion Menu -> Preferences -> Network
   * Create/choose a network with these settings
     * [x] Allow virtual machines on this network to connect to external networks (using NAT)
     * [x] Connect this host Mac to this network
     * [x] Provide addresses on this network via DHCP
     * Choose a specific subnet range (ex: 10.0.0.0/255.255.255.0)
* **Note:** Do not open Fusion while CPI is active during VM creation/updating - you'll see errors about files being inaccessible. It's fine to open after VMs are all up and running.
   * **Tip:** To exit Fusion and leave VMs running, right-click the Fusion Dock icon, hold <kbd>Option</kbd> and `Force Quit`.

### Workstation for Linux/Windows setup
* Find the paths for these binaries: `vmrun` and `ovftool`
  * Windows defaults:
    * `C:\Program Files (x86)\VMware\VMware Workstation\vmrun.exe`
    * `C:\Program Files (x86)\VMware\VMware Workstation\OVFTool\ovftool.exe`
  * Linux:
    * `/usr/bin/vmrun`
    * `/usr/bin/ovftool`
* Network configured for NAT
    * Edit -> Virtual Network Editor
    * Create/choose a network with these settings
      * [x] NAT (share host's IP address with VMs)
      * [x] Use local DHCP service to distribute IP addresses to VMs
      * [x] Connect a host virtual adapter ([your vm network name]) to this network
      * Choose a specific subnet range (ex: 10.0.0.0/255.255.255.0)
* **Note:** Do not open any BOSH VMs in Workstation while CPI is active during VM creation/updating - you'll see errors about files being inaccessible. It's fine to open and view VMs are they are all up and running.

### Player for Linux setup
* Install each `.bundle` file with appropriate license.
* Find the paths for these binaries: `vmrun` and `ovftool`
   * `/usr/bin/vmrun`
   * `/usr/bin/ovftool`
* Use the existing NAT network `vmnet8`:
   * Print network info: `ip -4 addr show vmnet8`
   * Example output:
   ```
   6: vmnet8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN group default qlen 1000
       inet 10.0.0.1/24 brd 10.0.0.255 scope global vmnet8
          valid_lft forever preferred_lft forever
   ```

### Example deployment

```
# create the vm <example manifest below>
bosh create-env my-vm.yml \
  --vars-store ./state/vm-creds.yml \
  --state ./state/vm_state.json \
  -v cpi_url="https://github.com/micahyoung/bosh-vmrun-cpi-release/releases/download/v1.1.0/bosh-vmrun-cpi-release-1.1.0.tgz" \
  -v cpi_sha1="f2c971abbc4be6c97f77306fd49ad23fac238099" \
  -v stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-xenial-go_agent?v=97.10" \
  -v stemcell_sha1="9e832921a4a1279b8029b72f962abcb2f981b32c" \
  -v vmrun_bin_path="$(which vmrun)" \
  -v ovftool_bin_path="$(which ovftool)" \
  -v vm_store_path="/tmp/vm-store-path" \
  -v internal_ip=10.0.0.5  \
  -v internal_range=10.0.0.0/24 \
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

disk_pools:
- name: disks
  disk_size: 65_536

networks:
- name: default
  type: manual
  subnets:
  - range: ((internal_range))
    gateway: ((internal_gw))
    static: [((internal_ip))]
    reserved: [((internal_gw))]
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
  properties:
    blobstore:
      provider: local
      path: '/var/vcap/micro_bosh/data/cache' #depends on agents internal location
    agent: {mbus: "https://mbus:((mbus_bootstrap_password))@0.0.0.0:6868"}
    ntp: *ntp
    vmrun:
      vmrun_bin_path: "((vmrun_bin_path))"
      ovftool_bin_path: "((ovftool_bin_path))"
      vm_store_path: "((vm_store_path))"
      vm_start_max_wait_seconds: 600
      vm_soft_shutdown_max_wait_seconds: 30
  template:
    name: vmrun_cpi
    release: bosh-vmrun-cpi

variables:
- name: mbus_bootstrap_password
  type: password
```

## Troubleshooting

### Common errors

* `Error: The operation was canceled`
   * Usually indicates your host is out of memory
   
* `Error: This VM is in use.`
   * Usually indicates VMWare Fusion or Workstation is open. This prevents the vms being modified and can leave them in an invalid state.
   * Resolution: usually closing Fusion/Workstation resolves it issue. If not, you may need to manually delete the entire VM directory and use bosh to recreate.
   
* VMs not starting or failing to come up
   * Check if there are any unknown running VMs
   ```
   vmrun list
   ```
   
   * If so, shut down all running vms
   ```
   vmrun list | grep vmx | while read vmx; do [ -d $(dirname $vmx) ] || mkdir $(dirname $vmx); [ -f $vmx ] || curl -L https://github.com/micahyoung/bosh-vmrun-cpi-release/raw/v1.0.0/src/bosh-vmrun-cpi/test/fixtures/test.vmx -o $vmx; vmrun stop $vmx hard; done   
   ```
   * Reattempt operation
   
### Recovery from a hard-shutdown

If you have shutdown the physical machine running the Workstation/Fusion, your VMs are likely in a very inconsistent state. There are at least two ways to go about recovering

#### Delete all state and recreate all VMs: 

1. Stop all running VMs
1. Remove entire vm-store-path directory
1. Remove any BOSH deployment `state.json` files
1. Redeploy all VMs


## Development
### Running tests

Linux/MacOS
```bash
export SSH_HOSTNAME="localhost"
export SSH_PORT="22"
export SSH_USERNAME=<user with priveledges to execute vmrun>
export SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa_vmrun)"
export SSH_PLATFORM=<"windows" | "linux" | "darwin">

cd src/bosh-vmrun-cpi
ginkgo -r
```

Windows
```powershell
$env:SSH_HOSTNAME="localhost"
$env:SSH_PORT="22"
$env:SSH_USERNAME=<user with priveledges to execute vmrun>
$env:SSH_PRIVATE_KEY=@(cat ~\.ssh\id_rsa_vmrun | out-string)
$env:SSH_PLATFORM=<"windows" | "linux" | "darwin">

cd .\src\bosh-vmrun-cpi
ginkgo -r

```
