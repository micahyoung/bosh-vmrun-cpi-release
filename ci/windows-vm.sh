#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

cd $(dirname $0)
RELEASE_DIR=../
: ${STATE_DIR:="$PWD/state"}
if ! [ -f $STATE_DIR/windows-vm-env.sh ]; then
  echo "no $STATE_DIR/windows-vm-env.sh file. Create and fill with required fields"
  exit 1
fi

source $STATE_DIR/windows-vm-env.sh

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

if [ -n ${RECREATE_RELEASE:-""} -o ! -f $STATE_DIR/cpi.tgz ] ; then
  echo "-----> `date`: Create dev release"
  $bosh_bin create-release --sha2 --force --dir $RELEASE_DIR --tarball $STATE_DIR/cpi.tgz
fi

echo "-----> `date`: Deploy Start"

cpi_url=file://$STATE_DIR/cpi.tgz
cpi_sha1=$(shasum -a1 < $STATE_DIR/cpi.tgz | awk '{print $1}')

$bosh_bin ${BOSH_COMMAND:-"create-env"} windows-vm.yml \
  --vars-env CI \
  --vars-store $STATE_DIR/windows-vm-creds.yml \
  --state $STATE_DIR/windows-vm-state.json \
  -v cpi_url=$cpi_url \
  -v cpi_sha1=$cpi_sha1 \
  ${RECREATE_VM:+"--recreate"} \
  ;

echo "-----> `date`: Deploy Done"
