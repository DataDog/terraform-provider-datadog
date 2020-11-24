#!/usr/bin/env bash

./extract-code-blocks.sh

for f in examples/v*/*/*.go ; do
    df=$(dirname $f)/$(basename $f .go)
    mkdir -p $df
    cp $f $df/main.go
done

ls examples/v*/*/*/main.go | xargs -P $(($(nproc)*2)) -n 1 go build -o /dev/null
if [ $? -ne 0 ]; then
    echo -e "Failed to build examples"
    exit 1
fi
