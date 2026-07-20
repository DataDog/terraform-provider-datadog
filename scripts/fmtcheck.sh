#!/usr/bin/env bash

EXIT_CODE=0

set -o pipefail

# By default fmtcheck validates the whole repository. When FMTCHECK_SCOPE is
# set (a newline-separated list of files), only those files are checked so a
# per-artifact caller isn't failed by unrelated formatting drift elsewhere.
go_files=()
tf_files=()
if [[ -n ${FMTCHECK_SCOPE:-} ]]; then
    while IFS= read -r scoped_file; do
        [[ -z ${scoped_file} ]] && continue
        case "${scoped_file}" in
            *.go) go_files+=("${scoped_file}") ;;
            *.tf | *.tfvars | *.tftest.hcl) tf_files+=("${scoped_file}") ;;
        esac
    done <<< "${FMTCHECK_SCOPE}"
else
    while IFS= read -r go_file; do
        go_files+=("${go_file}")
    done < <(find . -name '*.go' | grep -v vendor)
    tf_files=(examples)
fi

# Check goimports
if [[ ${#go_files[@]} -gt 0 ]]; then
    echo "==> Checking that code complies with goimports requirements..."
    goimports_files=$(goimports -format-only -d -l "${go_files[@]}")
    if [[ -n ${goimports_files} ]]; then
        echo 'gofmt needs running on the following files:'
        echo "${goimports_files}"
        echo "You can use the command: \`make fmt\` to reformat code."
        EXIT_CODE=1
    fi
fi

# Check the example terraform files pass terraform fmt
if [[ ${#tf_files[@]} -gt 0 ]]; then
    echo "==> Checking that examples pass with terraform fmt requirements"
    if ! terraform_fmt=$(terraform fmt -recursive -check -diff "${tf_files[@]}" 2>&1); then
        echo "Files in the \`example\` folder aren't terraform formatted"
        echo "You can use the command \`make fmt\` to reformat the following:"
        echo "${terraform_fmt}"
        EXIT_CODE=2
    fi
fi

exit $EXIT_CODE
