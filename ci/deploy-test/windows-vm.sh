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

source state/env.sh
: ${VMRUN_BIN_PATH?"!"}
: ${OVFTOOL_BIN_PATH?"!"}
: ${VDISKMANAGER_BIN_PATH?"!"}
: ${VMRUN_NETWORK:?"!"}
: ${DIRECTOR_IP?"!"}
: ${NETWORK_CIDR:?"!"}
: ${NETWORK_GW:?"!"}
: ${NETWORK_DNS:?"!"}

if [ -n ${RESET:-""} ]; then
  FORGET_DISKS="y"
  RECREATE_VM="y"
  RECREATE_RELEASE="y"
fi

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-darwin-amd64"
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

#STEMCELL=/Users/micah/workspace/stemcells/bosh-stemcell-1200.13-vsphere-esxi-windows2012R2-go_agent.tgz
STEMCELL=~/workspace/stemcells/bosh-stemcell-1709.8-vsphere-esxi-windows2016-go_agent.tgz
if ! [ -f $STEMCELL ]; then
	echo "missing stemcell: $STEMCELL"
	exit 1
fi

if [ -n ${RECREATE_RELEASE:-""} ]; then
  echo "-----> `date`: Create dev release"
  HOME=$PWD/state/bosh_home \
    $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball $PWD/state/cpi.tgz
fi

echo "-----> `date`: Create env"

if [ -n ${FORGET_DISKS:-""} ]; then
  jq -r ' .disks = [] | .current_disk_id = ""' state/bosh_state.json > state/new_bosh_state.json
  mv state/new_bosh_state.json state/bosh_state.json
fi

vm_store_path=$PWD/state/vm-store-path
if ! [ -d $vm_store_path ]; then
  mkdir -p $vm_store_path
fi

HOME=$PWD/state/bosh_home \
$bosh_bin interpolate windows-vm.yml \
  -v internal_ip="$DIRECTOR_IP" \
  --vars-store ./state/windows-vm-creds.yml \
;

stemcell_sha1=$(shasum -a1 < $STEMCELL | awk '{print $1}')

export BOSH_LOG_LEVEL=debug
HOME=$PWD/state/bosh_home \
$bosh_bin ${BOSH_COMMAND:-create-env} windows-vm.yml \
  --vars-file ./state/windows-vm-creds.yml \
  --state ./state/bosh_state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v director_name=bosh-1 \
  -v internal_ip="$DIRECTOR_IP"  \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v network_name="$VMRUN_NETWORK" \
  -v dns_recursor_ip="$NETWORK_DNS"  \
  -v stemcell_url=file://$STEMCELL \
  -v stemcell_sha1=$stemcell_sha1 \
  -v vm_store_path="$vm_store_path" \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  ${RECREATE_VM:+"--recreate"} \
  ;
