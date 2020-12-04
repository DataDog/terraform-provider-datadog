#!/usr/bin/env bash

OUTPUT=${1:-examples}

cd ${0%/*}

ls api/v1/datadog/docs/*Api.md | xargs -n1 ./extract-code-blocks.awk -v output="${OUTPUT}/v1"
ls api/v2/datadog/docs/*Api.md | xargs -n1 ./extract-code-blocks.awk -v output="${OUTPUT}/v2"
