#!/bin/bash
set -o errexit
set -o pipefail

stemcell_vmx="$(dirname $0)/test/fixtures/test.vmx"
if which vmrun; then
  vmrun=$(which vmrun)
else
  vmrun="/Applications/VMware Fusion.app/Contents/Library/vmrun"
fi

"$vmrun" list | grep vm-store-path | while read file; do
  if ! [ -f $file ]; then
    mkdir -p "$(dirname "$file")"
    cp $stemcell_vmx "$file"
  fi
  "$vmrun" stop "$file"
  "$vmrun" deletevm "$file"
done
