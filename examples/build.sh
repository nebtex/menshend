#!/usr/bin/env bash

current_dir=$PWD

# build binaries
cd ../cmd/menshend && go build && cp menshend $current_dir

cd $current_dir

# run docker-build
docker build -t nebtex/menshend:development .

# upload to dockerhub
docker push nebtex/menshend:development
