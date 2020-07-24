#!/usr/bin/env bash

CMD=( go test -coverpkg=$(go list ./... | grep -v /test | paste -sd "," -) -coverprofile=coverage.txt -covermode=atomic $(go list ./...) -json )

if [ "$#" -ne 2 ]; then
	"${CMD[@]}"
else
	if [ "$RERECORD_FAILED_TESTS" == "true" ]; then
		RECORD=true "${CMD[@]}" $1 $2
	else
		"${CMD[@]}" $1 $2
	fi
fi
