#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)

echo "-----> `date`: Downloading ESXi stemcell"
stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3541.5"
if ! [ -f state/stemcell.tgz ]; then
  curl -L $stemcell_url > state/stemcell.tgz
fi

if ! [ -f state/bosh.pem ]; then
  ssh-keygen -f state/bosh.pem -P ''
fi

echo "-----> `date`: Create dev release"
bosh create-release --sha2 --force --dir ./../ --tarball ./state/cpi.tgz

echo "-----> `date`: Create env"

stemcell_sha1=$(shasum -a1 < state/stemcell.tgz | awk '{print $1}')
bosh create-env deployment.yml \
  -v cpi_url=file://./state/cpi.tgz \
  -v auth_url=. \
  -v default_key_name=. \
  -v default_security_groups=. \
  -v net_id=. \
  -v openstack_domain=. \
  -v openstack_password=. \
  -v openstack_project=. \
  -v openstack_tenant=. \
  -v openstack_username=. \
  -v region=. \
  -v stemcell_url=file://./state/stemcell.tgz \
  -v stemcell_sha1=$stemcell_sha1 \
  -v vm_ip=10.0.0.3 \
  --state ./state/state.json \
  ;
