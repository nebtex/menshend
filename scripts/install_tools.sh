#!/usr/bin/env bash

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

if command_exists goimports;then
    echo "GoImports is already installed"
else
    go get golang.org/x/tools/cmd/goimports
fi

if command_exists gometalinter;then
    echo "GoMetaLinter is already installed"
else
    go get -u github.com/alecthomas/gometalinter
    gometalinter --install
fi
