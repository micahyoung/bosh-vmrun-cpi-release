#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../../

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

bosh_cli_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64"
if ! [ -f bin/bosh ]; then
  curl -L $bosh_cli_url > bin/bosh
  chmod +x bin/bosh
fi

echo "-----> `date`: Downloading ESXi stemcell"
stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3541.5"
if ! [ -f state/stemcell.tgz ]; then
  curl -L $stemcell_url > state/stemcell.tgz
fi

source state/env.sh
: ${VCENTER_HOST:?"!"}
: ${VCENTER_USER:?"!"}
: ${VCENTER_PASSWORD:?"!"}
: ${VCENTER_DATACENTER:?"!"}
: ${VCENTER_DATASTORE:?"!"}
: ${DIRECTOR_IP?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_DNS:?"!"}
: ${VCENTER_NETWORK_NAME:?"!"}

if ! [ -f state/bosh.pem ]; then
  ssh-keygen -f state/bosh.pem -P ''
fi

if [ -n ${RECREATE_RELEASE:=""} ]; then
  echo "-----> `date`: Create dev release"
  bin/bosh create-release --sha2 --force --dir $RELEASE_DIR --tarball ./state/cpi.tgz
fi

echo "-----> `date`: Create env"

bin/bosh interpolate ~/workspace/bosh-deployment/bosh.yml \
  -v internal_ip="$DIRECTOR_IP" \
  --vars-store ./state/creds.yml \
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
bin/bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  -o ~/workspace/bosh-deployment/vsphere/cpi.yml \
  -o govmomi-vsphere-cpi-opsfile.yml \
  --vars-store ./state/creds.yml \
  --state ./state/state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v director_name=bosh-1 \
  -v internal_ip="$DIRECTOR_IP"  \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v internal_dns="$NETWORK_DNS"  \
  -v network_name="$VCENTER_NETWORK_NAME" \
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
