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
  RECREATE_RELEASE="y"
  RECREATE_VM="y"
fi

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-darwin-amd64"
bosh_bin="bin/bosh-$OSTYPE"
if ! [ -f $bosh_bin ]; then
  curl -L $bosh_cli_linux_url > bin/bosh-linux-gnu
  curl -L $bosh_cli_darwin_url > bin/bosh-darwin17
  chmod +x bin/bosh*
fi

concourse_bosh_deployment_url="https://github.com/concourse/concourse-bosh-deployment.git"
if ! [ -d state/concourse-bosh-deployment ]; then
  git clone $concourse_bosh_deployment_url state/concourse-bosh-deployment
  pushd state/concourse-bosh-deployment
    git checkout 14323c4cc2320107dfb2dc622bca0d6861c517d4
  popd
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

linux_stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-xenial-go_agent?v=97.10"
if ! [ -f state/linux-stemcell.tgz ]; then
  echo "-----> `date`: Downloading stemcell"
  curl -L $linux_stemcell_url > state/linux-stemcell.tgz
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

linux_stemcell_sha1=$(shasum -a1 < state/linux-stemcell.tgz | awk '{print $1}')

#export BOSH_LOG_LEVEL=debug
HOME=$PWD/state/bosh_home \
$bosh_bin ${BOSH_COMMAND:-"create-env"} state/concourse-bosh-deployment/lite/concourse.yml \
  -l state/concourse-bosh-deployment/versions.yml \
  -o concourse-vmrun-opsfile.yml \
  --vars-store ./state/concourse-creds.yml \
  --state ./state/concourse_state.json \
  -v deployment_name="concourse" \
  -v vm_type="small" \
  -v public_ip="$DIRECTOR_IP" \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v internal_ip="$DIRECTOR_IP"  \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v dns_recursor_ip="$NETWORK_DNS"  \
  -v linux_stemcell_url=file://$PWD/state/linux-stemcell.tgz \
  -v linux_stemcell_sha1=$linux_stemcell_sha1 \
  -v vm_store_path="$vm_store_path" \
  -v network_name="$VMRUN_NETWORK" \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  -v vcap_mkpasswd=$VCAP_MKPASSWD \
  ${RECREATE_VM:+"--recreate"} \
  ;

echo "-----> `date`: Deploy Done"
