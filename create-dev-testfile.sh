#!/bin/bash
set -eo pipefail

main() {
  resource=$1
  name=$2

  if [[ -z $resource ]]; then
    echo "No resource provided. Pass one in through \`make dev-new resource=<resource>\`"
    exit 1
  fi

  if [[ -z $name ]]; then
    echo ""
    echo "No resource name provided, using \`${resource}\` as name"
    name=${resource}
  fi

  resource="prefect_${resource}"

  dev_file_target="./dev/${name}"

  mkdir -p $dev_file_target

  cat <<EOF > $dev_file_target/${resource}.tf
terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

provider "prefect" {}

resource "${resource}" "${name}" {}
EOF

  echo ""
  echo "run:"
  echo "cd ${dev_file_target} && terraform plan"
}

main $@
