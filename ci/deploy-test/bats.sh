#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

bosh_cli_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-darwin-amd64"
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
: ${DIRECTOR_IP?"!"}
: ${FIRST_IP:?"!"}
: ${SECOND_IP:?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_RESERVED_RANGE:?"!"}
: ${NETWORK_STATIC_RANGE:?"!"}
: ${VCENTER_NETWORK_NAME:?"!"}
DIRECTOR_ADMIN_PASSWORD=$(bosh int $PWD/state/bosh-deployment-creds.yml --path /admin_password)
DIRECTOR_CA_CERT=$(bosh int $PWD/state/bosh-deployment-creds.yml --path /default_ca/certificate)
ENVIRONMENT=bats
PRIVATE_KEY="$(bin/bosh int $PWD/state/bosh-deployment-creds.yml --path /jumpbox_ssh/private_key)"
echo "$PRIVATE_KEY" > $PWD/state/bosh.pem

export BAT_STEMCELL=$PWD/state/stemcell.tgz
export BAT_DEPLOYMENT_SPEC=$PWD/state/bats.yml
export BAT_BOSH_CLI=$PWD/bin/bosh
export BAT_DNS_HOST=$DIRECTOR_IP
export BAT_INFRASTRUCTURE=vsphere
export BAT_NETWORKING=manual
export BAT_PRIVATE_KEY="$PRIVATE_KEY"

export BOSH_ENVIRONMENT=$ENVIRONMENT
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET="$DIRECTOR_ADMIN_PASSWORD"
export BOSH_CA_CERT="$DIRECTOR_CA_CERT"

cat > $PWD/state/bats.yml <<EOF
---
cpi: vsphere
properties:
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: latest
  pool_size: 1
  instances: 1
  second_static_ip: "$SECOND_IP" # Secondary (private) IP assigned to the bat-release job vm, used for testing network reconfiguration, must be in the primary network & different from static_ip
  networks:
  - name: static
    type: manual
    static_ip: "$FIRST_IP" # Primary (private) IP assigned to the bat-release job vm, must be in the static range
    cidr: "$NETWORK_CIDR"
    reserved: ['$NETWORK_RESERVED_RANGE', '$DIRECTOR_IP'] # multiple reserved ranges are allowed but optional
    static: ['$NETWORK_STATIC_RANGE']
    gateway: "$NETWORK_GW"
    vlan: "$VCENTER_NETWORK_NAME" # vSphere network name
EOF

bosh alias-env $ENVIRONMENT \
  -e https://$DIRECTOR_IP:25555 \
  --ca-cert="$BOSH_CA_CERT" \
;

#export BAT_DEBUG_MODE=true
pushd ~/workspace/bosh-acceptance-tests
  bundle
  bundle exec rspec spec --tag ~vip_networking --tag ~dynamic_networking --tag ~root_partition --tag ~raw_ephemeral_storage
popd
