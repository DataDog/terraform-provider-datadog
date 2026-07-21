#!/usr/bin/env bash

EXIT_CODE=0

set -o pipefail

# Check goimports
echo "==> Checking that code complies with goimports requirements..."
goimports_files=$(go tool goimports -format-only -d -l $(find . -name '*.go' | grep -v vendor))
if [[ -n ${goimports_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${goimports_files}"
    echo "You can use the command: \`make fmt\` to reformat code."
    EXIT_CODE=1
fi

echo "==> Checking that terraform fmt passes"
terraform_fmt=$(terraform fmt -recursive -check -diff 2>&1)
if [[ $? -ne 0 ]]; then
    echo "Files aren't terraform formatted"
    echo "You can use the command \`make fmt\` to reformat the following:"
    echo "${terraform_fmt}"
    EXIT_CODE=2
fi

exit $EXIT_CODE
