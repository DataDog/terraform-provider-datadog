#!/usr/bin/env bash

EXIT_CODE=0

# Check gofmt
echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(gofmt -d -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` to reformat code."
    EXIT_CODE=1
fi

# Check the example terraform files pass terraform fmt
echo "==> Checking that examples pass with terraform fmt requirements"
if [[ -n $(terraform fmt -recursive -check -diff examples)  ]]; then
    echo "Files in the \`example\` folder aren't terraform formatted"
    echo "You can use the command \`make fmt\` to reformat the following:"
    echo "$(terraform fmt -recursive -diff examples)"
    EXIT_CODE=2
fi

exit $EXIT_CODE
