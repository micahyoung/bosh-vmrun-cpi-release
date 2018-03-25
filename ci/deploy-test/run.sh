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

if [ -f state/state.json ]; then
  jq -r '.stemcells = [] | .current_stemcell_id = ""' state/state.json > state/new_state.json
  mv state/new_state.json state/state.json
fi

echo "-----> `date`: Create dev release"
bosh create-release --sha2 --force --dir $RELEASE_DIR --tarball ./state/cpi.tgz

echo "-----> `date`: Create env"

export BOSH_LOG_LEVEL=debug
stemcell_sha1=$(shasum -a1 < state/stemcell.tgz | awk '{print $1}')
bosh create-env deployment.yml \
  -v cpi_url=file://./state/cpi.tgz \
  -v auth_url=. \
  -v default_key_name=. \
  -v default_security_groups=. \
  -v net_id=. \
  -v vcenter_address=$VCENTER_HOST \
  -v vcenter_user=$VCENTER_USER \
  -v vcenter_password=$VCENTER_PASSWORD \
  -v vcenter_datacenter=$VCENTER_DATACENTER \
  -v vcenter_datastore=$VCENTER_DATASTORE \
  -v region=. \
  -v stemcell_url=file://./state/stemcell.tgz \
  -v stemcell_sha1=$stemcell_sha1 \
  -v vm_ip="172.16.125.10" \
  -v net_dns="[172.16.125.2]" \
  -v net_gateway="172.16.125.2" \
  -v net_range="172.16.125.0/24" \
  -v private_key_path=./state/bosh.pem \
  --state ./state/state.json \
  ;
