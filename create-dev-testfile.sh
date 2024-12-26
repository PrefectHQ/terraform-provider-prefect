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

  dev_file_target="${PWD}/dev/${name}"

  mkdir -p $dev_file_target

  cat <<EOF > $dev_file_target/${resource}.tf
terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

provider "prefect" {
$([ -n "$PREFECT_API_URL" ] && echo '  endpoint = "'$PREFECT_API_URL'"')
$([ -n "$PREFECT_API_KEY" ] && echo '  api_key = "'$PREFECT_API_KEY'"')
$([ -n "$PREFECT_CLOUD_ACCOUNT_ID" ] && echo '  account_id = "'$PREFECT_CLOUD_ACCOUNT_ID'"')
}

resource "${resource}" "${name}" {}
EOF

  cmd="cd ${dev_file_target} && terraform plan"
  echo ""
  echo "run:"
  echo "${cmd}"
  echo "(copied to clipboard)"
  printf "${cmd}" | pbcopy
}

main $@
