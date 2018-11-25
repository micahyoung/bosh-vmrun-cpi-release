#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../
: ${STATE_DIR:="$PWD/state"}
if ! [ -f $STATE_DIR/bosh-deployment-env.sh ]; then
  echo "no $STATE_DIR/bosh-deployment-env.sh file. Create and fill with required fields"
  exit 1
fi

source $STATE_DIR/bosh-deployment-env.sh

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

bosh_deployment_url="https://github.com/cloudfoundry/bosh-deployment.git"
if ! [ -d $STATE_DIR/bosh-deployment ]; then
  git clone $bosh_deployment_url $STATE_DIR/bosh-deployment
  pushd $STATE_DIR/bosh-deployment
    git checkout v1.1.0
  popd
fi

if [ -n ${RECREATE_RELEASE:-""} -o ! -f $STATE_DIR/cpi.tgz ] ; then
  echo "-----> `date`: Create dev release"
  $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball $STATE_DIR/cpi.tgz
fi

echo "-----> `date`: Deploy Start" 
cpi_url=file://$STATE_DIR/cpi.tgz
cpi_sha1=$(shasum -a1 < $STATE_DIR/cpi.tgz | awk '{print $1}')

$bosh_bin ${BOSH_COMMAND:-"create-env"} $STATE_DIR/bosh-deployment/bosh.yml \
  -o ./vmrun-cpi-opsfile.yml \
  -o $STATE_DIR/bosh-deployment/jumpbox-user.yml \
  --vars-store $STATE_DIR/bosh-deployment-creds.yml \
  --state $STATE_DIR/bosh-deployment-state.json \
  --vars-env CI \
  -v cpi_url=$cpi_url \
  -v cpi_sha1=$cpi_sha1 \
  ${RECREATE_VM:+"--recreate"} \
;

$bosh_bin -e $CI_internal_ip alias-env ci \
  --ca-cert=<($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path=/default_ca/certificate) \
;

$bosh_bin -e ci login \
  --client=admin \
  --client-secret=$($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path=/admin_password) \
  --ca-cert=<($bosh_bin int $STATE_DIR/bosh-deployment-creds.yml --path=/default_ca/certificate) \
;

$bosh_bin -e ci update-cloud-config -n \
  vmrun-cloud-config.yml \
  --vars-env CI \
;