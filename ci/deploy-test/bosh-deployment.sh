#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../../

echo "-----> `date`: Downloading ESXi stemcell"
stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3541.5"
if ! [ -f state/stemcell.tgz ]; then
  curl -L $stemcell_url > state/stemcell.tgz
fi

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

source state/env.sh
: ${VCENTER_HOST:?"!"}
: ${VCENTER_USER:?"!"}
: ${VCENTER_PASSWORD:?"!"}
: ${VCENTER_DATACENTER:?"!"}
: ${VCENTER_DATASTORE:?"!"}

if ! [ -f state/bosh.pem ]; then
  ssh-keygen -f state/bosh.pem -P ''
fi

if [ -n ${RECREATE_RELEASE:=""} ]; then
  echo "-----> `date`: Create dev release"
  bosh create-release --sha2 --force --dir $RELEASE_DIR --tarball ./state/cpi.tgz
fi

echo "-----> `date`: Create env"

bosh interpolate ~/workspace/bosh-deployment/bosh.yml \
  -v internal_ip="172.16.125.5" \
  --vars-store ./state/certs.yml \
;

if [ -n ${RECREATE_STEMCELLS:=""} ]; then
  jq -r '.stemcells = [] | .current_stemcell_id = ""' state/state.json > state/new_state.json
  mv state/new_state.json state/state.json
fi

if [ -n ${RECREATE_DISKS:=""} ]; then
  jq -r ' .disks = [] | .current_disk_id = ""' state/state.json > state/new_state.json
  mv state/new_state.json state/state.json
fi

export BOSH_LOG_LEVEL=debug
stemcell_sha1=$(shasum -a1 < state/stemcell.tgz | awk '{print $1}')
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  -o ~/workspace/bosh-deployment/vsphere/cpi.yml \
  -o govmomi-vsphere-cpi-opsfile.yml \
  --vars-file ./state/certs.yml \
  --vars-store ./state/creds.yml \
  --state ./state/state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v director_name=bosh-1 \
  -v internal_cidr="172.16.125.0/24" \
  -v internal_gw="172.16.125.2" \
  -v internal_ip="172.16.125.5"  \
  -v internal_dns="172.16.125.2"  \
  -v network_name="VM Network" \
  -v vcenter_dc=$VCENTER_DATACENTER \
  -v vcenter_ds=$VCENTER_DATASTORE \
  -v vcenter_ip=$VCENTER_HOST \
  -v vcenter_user=$VCENTER_USER \
  -v vcenter_password=$VCENTER_PASSWORD \
  -v vcenter_templates=bosh-1-templates \
  -v vcenter_vms=bosh-1-vms \
  -v vcenter_disks=bosh-1-disks \
  -v vcenter_cluster=cluster1 \
  -v stemcell_url=file://$PWD/state/stemcell.tgz \
  -v stemcell_sha1=$stemcell_sha1 \
  ;
