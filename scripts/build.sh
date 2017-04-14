#!/usr/bin/env bash
go get github.com/mitchellh/gox
cd cmd/menshend
gox
go get -u github.com/tcnksm/ghr

