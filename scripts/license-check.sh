#!/usr/bin/env bash
echo "Ensuring all dependencies are present in LICENSE-3rdparty.csv ..."
# Tidy every module so its go.sum reflects the current dependency set, then
# check them all. The generator (.generator-v2) is a separate Go module, so its
# dependencies must be checked too; otherwise generator dependency changes
# silently bypass this gate.
GO_SUMS="go.sum"
go mod tidy
if [ -f ".generator-v2/go.sum" ]; then
    (cd .generator-v2 && go mod tidy)
    GO_SUMS="$GO_SUMS .generator-v2/go.sum"
fi
ALL_DEPS=`cat $GO_SUMS | awk '{print $1}' | sort -u | sed "s|^\(.*\)|go.sum,\1,|"`
DEPS_NOT_FOUND=""
for one_dep in `echo $ALL_DEPS`; do
    cat LICENSE-3rdparty.csv | grep "$one_dep" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        DEPS_NOT_FOUND="${DEPS_NOT_FOUND}\n${one_dep}<LICENSE>,<COPYRIGHT>"
    fi
done
if [ -n "$DEPS_NOT_FOUND" ]; then
    printf "Some dependencies were not found in LICENSE-3rdparty.csv, please add: $DEPS_NOT_FOUND\n\n"
    exit 1
else
    echo "LICENSE-3rdparty.csv is up to date"
fi
