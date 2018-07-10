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
  FORGET_STEMCELLS="y"
  FORGET_DISKS="y"
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

bosh_deployment_url="https://github.com/cloudfoundry/bosh-deployment"
if ! [ -d state/bosh-deployment ]; then
  git clone $bosh_deployment_url state/bosh-deployment
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

echo "-----> `date`: Downloading ESXi stemcell"
stemcell_url="https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent?v=3541.5"
if ! [ -f state/stemcell.tgz ]; then
  curl -L $stemcell_url > state/stemcell.tgz
fi

if [ -n ${RECREATE_RELEASE:-""} ]; then
  echo "-----> `date`: Create dev release"
  HOME=$PWD/state/bosh_home \
    $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball ./state/cpi.tgz
fi

echo "-----> `date`: Create env"

$bosh_bin interpolate state/bosh-deployment/bosh.yml \
  -o state/bosh-deployment/jumpbox-user.yml \
  -o state/bosh-deployment/misc/powerdns.yml \
  -v internal_ip="$DIRECTOR_IP" \
  --vars-store ./state/bosh-deployment-creds.yml \
;

DIRECTOR_CA_CERT=$($bosh_bin int state/bosh-deployment-creds.yml --path /default_ca/certificate)

if [ -n ${FORGET_STEMCELLS:-""} ]; then
  jq -r '.stemcells = [] | .current_stemcell_id = ""' state/bosh_state.json > state/new_bosh_state.json
  mv state/new_bosh_state.json state/bosh_state.json
fi

if [ -n ${FORGET_DISKS:-""} ]; then
  jq -r ' .disks = [] | .current_disk_id = ""' state/bosh_state.json > state/new_bosh_state.json
  mv state/new_bosh_state.json state/bosh_state.json
fi

vm_store_path=$PWD/state/vmrun/vms
if ! [ -d $vm_store_path ]; then
  mkdir -p $vm_store_path
fi

stemcell_sha1=$(shasum -a1 < state/stemcell.tgz | awk '{print $1}')

#export BOSH_LOG_LEVEL=debug
HOME=$PWD/state/bosh_home \
$bosh_bin create-env state/bosh-deployment/bosh.yml \
  -o state/bosh-deployment/jumpbox-user.yml \
  -o state/bosh-deployment/misc/powerdns.yml \
  -o state/bosh-deployment/vsphere/cpi.yml \
  -o vmrun-vsphere-cpi-opsfile.yml \
  --vars-file ./state/bosh-deployment-creds.yml \
  --state ./state/bosh_state.json \
  -v cpi_url=file://$PWD/state/cpi.tgz \
  -v director_name=bosh-1 \
  -v internal_ip="$DIRECTOR_IP"  \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
  -v dns_recursor_ip="$NETWORK_DNS"  \
  -v stemcell_url=file://$PWD/state/stemcell.tgz \
  -v stemcell_sha1=$stemcell_sha1 \
  -v vm_store_path="$vm_store_path" \
  -v network_name="$VMRUN_NETWORK" \
  -v vmrun_bin_path="$VMRUN_BIN_PATH" \
  -v ovftool_bin_path="$OVFTOOL_BIN_PATH" \
  -v vdiskmanager_bin_path="$VDISKMANAGER_BIN_PATH" \
  -v vcap_mkpasswd=$VCAP_MKPASSWD \
  -v vcenter_dc="deleteme" \
  -v vcenter_ds="deleteme" \
  -v vcenter_ip="deleteme" \
  -v vcenter_user="deleteme" \
  -v vcenter_password="deleteme" \
  -v vcenter_templates=bosh-1-templates \
  -v vcenter_vms=bosh-1-vms \
  -v vcenter_disks=bosh-1-disks \
  -v vcenter_cluster=cluster1 \
  ${RECREATE_VM:+"--recreate"} \
  ;

exit
cat > state/cloud-config-opsfile.yml <<EOF
- type: replace
  path: /networks/name=default/subnets/0/reserved
  value: [$NETWORK_RESERVED_RANGE]

- type: replace
  path: /compilation/workers
  value: 1
EOF

HOME=$PWD/state/bosh_home \
$bosh_bin update-cloud-config state/bosh-deployment/vsphere/cloud-config.yml \
  --non-interactive \
  --client admin \
  --client-secret $($bosh_bin int $PWD/state/bosh-deployment-creds.yml --path /admin_password) \
  --ca-cert "$($bosh_bin int $PWD/state/bosh-deployment-creds.yml --path /default_ca/certificate)" \
  -e $DIRECTOR_IP \
  -o state/cloud-config-opsfile.yml \
  -v network_name="$VMRUN_NETWORK" \
  -v internal_cidr="$NETWORK_CIDR" \
  -v internal_gw="$NETWORK_GW" \
;
