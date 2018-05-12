#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)

if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-darwin-amd64"
bosh_bin="$PWD/bin/bosh-$OSTYPE"
if ! [ -f $bosh_bin ]; then
  curl -L $bosh_cli_linux_url > bin/bosh-linux-gnu
  curl -L $bosh_cli_darwin_url > bin/bosh-darwin17
  chmod +x bin/bosh*
fi
bosh_ln=$(dirname $bosh_bin)/bosh
rm -f $bosh_ln
ln -s $bosh_bin $bosh_ln

bwats_url="https://github.com/cloudfoundry-incubator/bosh-windows-acceptance-tests"
bwats_dir="state/bosh-windows-acceptance-tests"
if ! [ -d "$bwats_dir" ]; then
  git clone $bwats_url $bwats_dir
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
: ${VCAP_MKPASSWD:?"!"}
DIRECTOR_ADMIN_PASSWORD=$($bosh_bin int $PWD/state/bosh-deployment-creds.yml --path /admin_password)
DIRECTOR_CA_CERT=$($bosh_bin int $PWD/state/bosh-deployment-creds.yml --path /default_ca/certificate --json | jq '.Blocks[0]')

#STEMCELL=$PWD/state/stemcell.tgz
STEMCELL=$PWD/state/windows-stemcell.tgz
STEMCELL_OS="windows2012R2"
BWATS_CONFIG_FILE=$PWD/state/bwats-config.json
BOSH_HOME=$PWD/state/bosh_home

cat > $BWATS_CONFIG_FILE <<EOF
{
  "bosh": {
    "ca_cert": $DIRECTOR_CA_CERT,
    "client": "admin",
    "client_secret": "$DIRECTOR_ADMIN_PASSWORD",
    "target": "$DIRECTOR_IP"
  },
  "stemcell_path": "$STEMCELL",
  "stemcell_os": "$STEMCELL_OS",
  "az": "z1",
  "vm_type": "default",
  "vm_extensions": "",
  "network": "default",
  "skip_cleanup": true
}
EOF

pushd $bwats_dir
  HOME=$BOSH_HOME \
  PATH="$PATH:$(dirname $bosh_ln)" \
  CONFIG_JSON=$BWATS_CONFIG_FILE \
    ginkgo -v .
popd
