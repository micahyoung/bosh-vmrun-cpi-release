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

bosh_cli_linux_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-linux-amd64"
bosh_cli_darwin_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.1.1-darwin-amd64"
bosh_bin="bin/bosh-$OSTYPE"
if ! [ -f $bosh_bin ]; then
  curl -L $bosh_cli_linux_url > bin/bosh-linux-gnu
  curl -L $bosh_cli_darwin_url > bin/bosh-darwin17
  chmod +x bin/bosh*
fi

zookeeper_release_url="https://github.com/cppforlife/zookeeper-release.git"
if ! [ -d $STATE_DIR/zookeeper-release ]; then
  git clone $zookeeper_release_url $STATE_DIR/zookeeper-release
  pushd $STATE_DIR/zookeeper-release
    git checkout 42b6835197bcfb051ff9e563d4f1757e4fd6d649 #0.0.9
  popd
fi

$bosh_bin -e ci upload-stemcell $CI_trusty_stemcell_url \
  --sha1 $CI_trusty_stemcell_sha1 \
;

$bosh_bin -e ci deploy $STATE_DIR/zookeeper-release/manifests/zookeeper.yml \
  -d zookeeper \
  -n \
;

$bosh_bin -e ci -d zookeeper run-errand smoke-tests

$bosh_bin -e ci -d zookeeper run-errand status