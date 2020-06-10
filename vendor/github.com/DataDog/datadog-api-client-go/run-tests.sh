#!/usr/bin/env sh
echo "Ensuring all dependencies are present in LICENSE-3rdparty.csv ..."
go mod tidy
ALL_DEPS=`cat go.sum | awk '{print $1}' | uniq | sort | sed "s|^\(.*\)|go.sum,\1,|"`
DEPS_NOT_FOUND=""
for one_dep in `echo $ALL_DEPS`; do
    cat LICENSE-3rdparty.csv | grep "$one_dep" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        DEPS_NOT_FOUND="${DEPS_NOT_FOUND}\n${one_dep}<LICENSE>,<COPYRIGHT>"
    fi
done
if [ -n "$DEPS_NOT_FOUND" ]; then
    echo "Some dependencies were not found in LICENSE-3rdparty.csv, please add: $DEPS_NOT_FOUND"
    exit 1
else
    echo "LICENSE-3rdparty.csv is up to date"
fi

# make sure the below installed dependencies don't get added to go.mod/go.sum
# unfortunately there's no better way to fix this than change directory
# this might get solved in Go 1.14: https://github.com/golang/go/issues/30515
cd `mktemp -d`
GO111MODULE=on go get -u golang.org/x/lint/golint
GO111MODULE=on go get -u gotest.tools/gotestsum@v0.4.1
cd -

golint ./...
declare -i RESULT=0
gotestsum --jsonfile gotestsum.json --format testname -- -coverpkg=$(go list ./... | grep -v /test | paste -sd "," -) -coverprofile=coverage.txt -covermode=atomic -v $(go list ./...)
RESULT+=$?
if [ "$CI" == "true" -a "$RESULT" -ne 0 ]; then
    RESULT=0
    echo "\n============= Rerunning failed tests =============\n"
    # NOTE: since `go test` (and `gotestsum`) don't allow specifying multiple different test cases
    # from different test modules with `-run`, we run them one by one in form of:
    # gotestsum <arguments> github.com/DataDog/datadog-api-client-go/tests/api/v<version>/datadog -run ^TestCaseName$
    while read -r i ; do
        gotestsum --format testname -- -v $i
        RESULT+=$?
    done <<EOF
        `cat gotestsum.json | jq -s -r -c '.[] | select(.Action=="fail") | select (.Test!=null) | "\(.Package) -run ^\(.Test)$"'`
EOF
fi
go mod tidy
exit $RESULT
