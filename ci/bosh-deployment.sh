#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../
: ${STATE_DIR:="$PWD/state"}
if ! [ -f $STATE_DIR/env.sh ]; then
  echo "no $STATE_DIR/env.sh file. Create and fill with required fields"
  exit 1
fi

source $STATE_DIR/env.sh
: ${VMRUN_BIN_PATH?"!"}
: ${OVFTOOL_BIN_PATH?"!"}
: ${VDISKMANAGER_BIN_PATH?"!"}
: ${VMRUN_NETWORK:?"!"}
: ${BOSH_DIRECTOR_IP?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_DNS:?"!"}
: ${NETWORK_RESERVED_RANGE:?"!"}
: ${WINDOWS_STEMCELL:?"!"}
: ${LINUX_STEMCELL:?"!"}
: ${SSH_TUNNEL_HOST:?"!"}
: ${SSH_TUNNEL_USERNAME:?"!"}
: ${SSH_TUNNEL_PLATFORM:?"!"}
: ${SSH_TUNNEL_PRIVATE_KEY:?"!"}
: ${STEMCELL_STORE_PATH:?"!"}

if [ -n ${RESET:-""} ]; then
  RECREATE_RELEASE="y"
  RECREATE_VM="y"
fi

export HOME="$STATE_DIR/bosh_home"

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-darwin-amd64"
bosh_bin="bin/bosh-$OSTYPE"
if ! [ -f $bosh_bin ]; then
  curl -L $bosh_cli_linux_url > bin/bosh-linux-gnu
  curl -L $bosh_cli_darwin_url > bin/bosh-darwin17
  chmod +x bin/bosh*
fi

bosh_deployment_url="https://github.com/cloudfoundry/bosh-deployment.git"
if ! [ -d $STATE_DIR/bosh-deployment ]; then
  git clone $bosh_deployment_url $STATE_DIR/bosh-deployment
  pushd $STATE_DIR/bosh-deployment
    git checkout v1.1.0
  popd
fi

if ! [ -f $LINUX_STEMCELL ]; then
  echo "Error: linux stemcell is required. Downlaod manually"
fi

if [ -n ${RECREATE_RELEASE:-""} -o ! -f $STATE_DIR/cpi.tgz ] ; then
  echo "-----> `date`: Create dev release"
  $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball $STATE_DIR/cpi.tgz
fi

echo "-----> `date`: Deploy Start"

vm_store_path=$STATE_DIR/vm-store-path
if ! [ -d $vm_store_path ]; then
  mkdir -p $vm_store_path
fi

linux_stemcell_sha1=$(shasum -a1 < $LINUX_STEMCELL | awk '{print $1}')

cpi_url=file://$STATE_DIR/cpi.tgz
cpi_sha1=$(shasum -a1 < $STATE_DIR/cpi.tgz | awk '{print $1}')

$bosh_bin ${BOSH_COMMAND:-"create-env"} $STATE_DIR/bosh-deployment/bosh.yml \
  -o ./vmrun-cpi-opsfile.yml \
  -o $STATE_DIR/bosh-deployment/jumpbox-user.yml \
  --vars-store $STATE_DIR/bosh-deployment-creds.yml \
  --state $STATE_DIR/bosh-deployment-state.json \
  -v director_name="vmrun" \
  -v blobstore_agent_password="foo" \
  -v nats_password="bar" \
  -v cpi_url=$cpi_url \
  -v cpi_sha1=$cpi_sha1 \
  -v internal_ip="$BOSH_DIRECTOR_IP" \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v stemcell_url=file://$LINUX_STEMCELL \
  -v stemcell_sha1=$linux_stemcell_sha1 \
  -v stemcell_store_path="$STEMCELL_STORE_PATH" \
  -v network_name="$VMRUN_NETWORK" \
  -v vm_store_path="$vm_store_path" \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  -v ssh_tunnel_host="$SSH_TUNNEL_HOST" \
  -v ssh_tunnel_username="$SSH_TUNNEL_USERNAME" \
  -v ssh_tunnel_platform="$SSH_TUNNEL_PLATFORM" \
  --var-file ssh_tunnel_private_key=<(echo "$SSH_TUNNEL_PRIVATE_KEY") \
  ${RECREATE_VM:+"--recreate"} \
  ;

$bosh_bin -e $BOSH_DIRECTOR_IP alias-env bosh \
  --ca-cert=<($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path /default_ca/certificate) \
;

$bosh_bin -e bosh login \
  --client=admin \
  --client-secret=$($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path /admin_password) \
  --ca-cert=<($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path /default_ca/certificate) \
;

$bosh_bin -e bosh update-cloud-config -n \
  vmrun-cloud-config.yml \
  -v internal_reserved_range="$NETWORK_RESERVED_RANGE" \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v network_name="$VMRUN_NETWORK" \
;
