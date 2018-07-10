#!/bin/bash
set -o errexit
set -o pipefail

ovftool="$OVFTOOL_BIN_PATH"
vmrun="$VMRUN_BIN_PATH"
stemcell_vmx="$(dirname $0)/test/fixtures/test.vmx"

"$vmrun" list | grep vm-store-path | while read file; do
  if ! [ -f $file ]; then
    mkdir -p "$(dirname "$file")"
    cp $stemcell_vmx "$file"
  fi
  "$vmrun" stop "$file"
  "$vmrun" deletevm "$file"
done
