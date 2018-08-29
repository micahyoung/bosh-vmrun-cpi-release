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
: ${VMRUN_WORKER_IP?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_DNS:?"!"}
: ${WINDOWS_STEMCELL:?"!"}
: ${LINUX_STEMCELL:?"!"}

# mount -o remount,exec /tmp
# export LC_CTYPE=en_US.UTF-8
# https://download3.vmware.com/software/wkst/file/VMware-Workstation-Full-14.1.3-9474260.x86_64.bundle
# https://download4.vmware.com/software/player/file/VMware-Player-12.5.9-7535481.x86_64.bundle
# https://download3.vmware.com/software/player/file/VMware-VIX-1.15.8-5528349.x86_64.bundle

if [ -n ${RESET:-""} ]; then
  RECREATE_RELEASE="y"
  RECREATE_VM="y"
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

if ! [ -f $LINUX_STEMCELL ]; then
  echo "Error: linux stemcell is required. Downlaod manually"
fi

if [ -n ${RECREATE_RELEASE:-""} ]; then
  echo "-----> `date`: Create dev release"
  HOME=$PWD/state/bosh_home \
    $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball ./state/cpi.tgz
fi

echo "-----> `date`: Deploy Start"

vm_store_path=$PWD/state/vm-store-path
if ! [ -d $vm_store_path ]; then
  mkdir -p $vm_store_path
fi

linux_stemcell_sha1=$(shasum -a1 < $LINUX_STEMCELL | awk '{print $1}')

HOME=$PWD/state/bosh_home \
$bosh_bin ${BOSH_COMMAND:-"create-env"} vmrun-worker.yml \
  --vars-store ./state/vmrun-worker.yml \
  --state ./state/vmrun_worker_state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v internal_ip="$VMRUN_WORKER_IP" \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v stemcell_url=file://$LINUX_STEMCELL \
  -v stemcell_sha1=$linux_stemcell_sha1 \
  -v network_name="$VMRUN_NETWORK" \
  -v vm_store_path="$vm_store_path" \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  ${RECREATE_VM:+"--recreate"} \
  ;

echo "-----> `date`: Deploy Done"
