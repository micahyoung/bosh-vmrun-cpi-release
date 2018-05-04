#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

govc_cli_linux_url="https://github.com/vmware/govmomi/releases/download/v0.17.1/govc_linux_amd64.gz"
govc_cli_darwin_url="https://github.com/vmware/govmomi/releases/download/v0.17.1/govc_darwin_amd64.gz"
govc_bin="bin/govc-$OSTYPE"
if ! [ -f $govc_bin ]; then
  curl -L $govc_cli_linux_url | gzip -d > bin/govc-linux-gnu
  curl -L $govc_cli_darwin_url | gzip -d > bin/govc-darwin17
  chmod +x bin/govc*
fi


source state/env.sh
: ${VCENTER_HOST:?"!"}
: ${VCENTER_USER:?"!"}
: ${VCENTER_PASSWORD:?"!"}
: ${VCENTER_DATACENTER:?"!"}
: ${VCENTER_DATASTORE:?"!"}

export GOVC_URL=https://$VCENTER_USER:$VCENTER_PASSWORD@$VCENTER_HOST
export GOVC_INSECURE=true


RUNNING_VMS=$($govc_bin find . -type m -runtime.powerState poweredOn | cut -d'/' -f4-)
for vm in $RUNNING_VMS; do
  if [ "$vm" == "jumpbox" ]; then continue; fi

  $govc_bin vm.power -force -off $vm 
done

INSTANCE_VMS=$($govc_bin ls /ha-datacenter/vm/vm-* | cut -d'/' -f4-)
for vm in $INSTANCE_VMS; do
  if [ "$vm" == "jumpbox" ]; then continue; fi

  $govc_bin vm.destroy $vm
done

if $govc_bin datastore.ls env/env-*.iso 2>/dev/null >/dev/null; then
  ENV_ISOS=$($govc_bin datastore.ls env | grep -e 'env-.*\.iso' | grep -v -- -flat)
  for iso in $ENV_ISOS; do
    $govc_bin datastore.rm -f env/$iso
  done
fi

if $govc_bin datastore.ls disk-*.vmdk 2>/dev/null >/dev/null; then
  DISKS=$($govc_bin datastore.ls disk-*.vmdk | grep -v -- -flat)
  for disk in $DISKS; do
    $govc_bin datastore.rm -f $disk
  done
fi

STEMCELL_VMS=$($govc_bin ls /ha-datacenter/vm/cs-* | cut -d'/' -f4-)
for vm in $STEMCELL_VMS; do
  $govc_bin vm.destroy $vm
done

