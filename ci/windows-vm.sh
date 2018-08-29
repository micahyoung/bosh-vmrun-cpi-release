#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../
if ! [ -f state/env.sh ]; then
  echo "no state/env.sh file. Create and fill with required fields"
  exit 1
fi

source state/env.sh
: ${VMRUN_BIN_PATH?"!"}
: ${OVFTOOL_BIN_PATH?"!"}
: ${VDISKMANAGER_BIN_PATH?"!"}
: ${VMRUN_NETWORK:?"!"}
: ${VM_IP?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_DNS:?"!"}
: ${WINDOWS_STEMCELL:?"!"}

if [ -n ${RESET:-""} ]; then
  RECREATE_VM="y"
  RECREATE_RELEASE="y"
fi

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-darwin-amd64"
bosh_bin="bin/bosh-$OSTYPE"
if ! [ -f $bosh_bin ]; then
  curl -L $bosh_cli_linux_url > bin/bosh-linux-gnu
  curl -L $bosh_cli_darwin_url > bin/bosh-darwin17
  chmod +x bin/bosh*
fi

golang_release_url="https://github.com/bosh-packages/golang-release"
golang_release_dir="state/golang-release"
if ! [ -d "$golang_release_dir" ]; then
  git clone $golang_release_url $golang_release_dir
  HOME=$PWD/state/bosh_home \
    $bosh_bin vendor-package --dir $RELEASE_DIR golang-1.9-linux $golang_release_dir
  HOME=$PWD/state/bosh_home \
    $bosh_bin vendor-package --dir $RELEASE_DIR golang-1.9-darwin $golang_release_dir
fi

if ! [ -f $WINDOWS_STEMCELL ]; then
	echo "missing stemcell: $WINDOWS_STEMCELL"
	exit 1
fi

if [ -n ${RECREATE_RELEASE:-""} ] || ! [ -d $RELEASE_DIR/dev_releases ] ; then
  echo "-----> `date`: Create dev release"
  HOME=$PWD/state/bosh_home \
    $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball $PWD/state/cpi.tgz
fi

echo "-----> `date`: Deploy Start"

vm_store_path=$PWD/state/vm-store-path
if ! [ -d $vm_store_path ]; then
  mkdir -p $vm_store_path
fi

stemcell_sha1=$(shasum -a1 < $WINDOWS_STEMCELL | awk '{print $1}')

HOME=$PWD/state/bosh_home \
$bosh_bin ${BOSH_COMMAND:-"create-env"} windows-vm.yml \
  --vars-store ./state/windows-vm-creds.yml \
  --state ./state/windows_vm_state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v internal_ip="$VM_IP"  \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v network_name="$VMRUN_NETWORK" \
  -v stemcell_url=file://$WINDOWS_STEMCELL \
  -v stemcell_sha1=$stemcell_sha1 \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  -v vm_store_path="$vm_store_path" \
  ${RECREATE_VM:+"--recreate"} \
  ;

echo "-----> `date`: Deploy Done"
