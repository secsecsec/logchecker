#!/bin/bash

# Copyright (c) 2015, Alexander Zaytsev. All rights reserved.
# Use of this source code is governed by a MIT-style
# license that can be found in the LICENSE file.
#

program="logchecker"
gobin="`which go`"
repo="github.com/z0rr0/logchecker"
buildDir=""

if [ -z "$GOPATH" ]; then
    echo "ERROR: set GOPATH env"
    exit 1
fi
if [ ! -x "$gobin" ]; then
    echo "ERROR: can't find 'go' binary"
    exit 2
fi

if [ -n "$TRAVIS_BUILD_DIR" ]; then
	buildDir="$TRAVIS_BUILD_DIR"
else
	buildDir="${GOPATH}/src/${repo}"
fi

cd ${buildDir}/logchecker
go test -v -cover -coverprofile=coverage.out || exit 1

echo "all tests done"
exit 0
