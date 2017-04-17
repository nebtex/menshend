#!/usr/bin/env bash
set -e
go get github.com/rakyll/statik
#pack static files
rm -rf statik
statik -src=ui
