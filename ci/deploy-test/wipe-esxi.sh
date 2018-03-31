#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)

source state/env.sh
: ${VCENTER_HOST:?"!"}
: ${VCENTER_USER:?"!"}
: ${VCENTER_PASSWORD:?"!"}
: ${VCENTER_DATACENTER:?"!"}
: ${VCENTER_DATASTORE:?"!"}

export GOVC_URL=https://$VCENTER_USER:$VCENTER_PASSWORD@$VCENTER_HOST
export GOVC_INSECURE=true

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi


RUNNING_VMS=$(govc find . -type m -runtime.powerState poweredOn | cut -d'/' -f4-)
for vm in $RUNNING_VMS; do
  govc vm.power -force -off $vm 
done

INSTANCE_VMS=$(govc ls /ha-datacenter/vm/vm-* | cut -d'/' -f4-)
for vm in $INSTANCE_VMS; do
  govc vm.destroy $vm
done

if govc datastore.ls env/env-*.iso 2>/dev/null >/dev/null; then
  ENV_ISOS=$(govc datastore.ls env | grep -e 'env-.*\.iso' | grep -v -- -flat)
  for iso in $ENV_ISOS; do
    govc datastore.rm -f env/$iso
  done
fi

if govc datastore.ls disk-*.vmdk 2>/dev/null >/dev/null; then
  DISKS=$(govc datastore.ls disk-*.vmdk | grep -v -- -flat)
  for disk in $DISKS; do
    govc datastore.rm -f $disk
  done
fi

STEMCELL_VMS=$(govc ls /ha-datacenter/vm/cs-* | cut -d'/' -f4-)
for vm in $STEMCELL_VMS; do
  govc vm.destroy $vm
done

