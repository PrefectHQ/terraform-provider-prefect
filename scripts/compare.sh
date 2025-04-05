#!/usr/bin/env bash

# Compares two versions Prefect OpenAPI schema:
# - one from a pre-existing file: openapi.json (stored in GitHub actions, for
#   example)
# - one from a retrieved file: openapi-current.json
#

# Prevent GitHub actions from exiting on the first error
set +e 

# Get the current schema
wget https://api.prefect.cloud/api/openapi.json \
  -O openapi-current.json \


# Compare with the previous schema from GitHub artifact
diff openapi.json openapi-current.json > /dev/null
exitcode="$?"

# Remove existing diff file, if any
rm -f diff.md

if [[ "${exitcode}" -eq 1 ]]; then
  echo 'difference found, capturing changes'

  # Provide context and open a diff snippet.
  printf 'API change detected. Create issue(s) if any code changes are required.' >> diff.md
  printf '\n\n```diff' >> diff.md

  # Insert the diff.
  dyff between openapi.json openapi-current.json \
    --set-exit-code \
    --omit-header \
    --output=github \
    >> diff.md

  # Close the diff snippet.
  printf '```' >> diff.md
else
  echo 'no differences found'
fi

# Turn exit-on-error back on.
set -e
