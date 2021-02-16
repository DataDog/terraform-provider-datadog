#!/usr/bin/env bash
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
    printf "Some dependencies were not found in LICENSE-3rdparty.csv, please add: $DEPS_NOT_FOUND\n\n"
    exit 1
else
    echo "LICENSE-3rdparty.csv is up to date"
fi
