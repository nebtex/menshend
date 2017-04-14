#!/usr/bin/env bash
go get github.com/mitchellh/gox
cd cmd/menshend
mkdir -p dist dist/bin
curl -fsL https://github.com/gliderlabs/sigil/releases/download/v0.4.0/sigil_0.4.0_Linux_x86_64.tgz  | tar -zxC dist/bin
./dist/bin/sigil -p -f version.tmpl MENSHEND_RELEASE=$MENSHEND_RELEASE > version.go
gox -output dist/menshend_{{.OS}}_{{.Arch}}
cd dist
dist=$(pwd)
mkdir -p prebuild release
for file in *
do
    if [[ -f $file ]]; then
        arch_dir="prebuild/${file%%.*}"
        if [ "${file#*.}" == "exe" ]; then
          binary="menshend.${file#*.}"
        else
          binary="menshend"
        fi
        mkdir -p "${arch_dir}"
        cp $file "${arch_dir}/${binary}"
        cp ../../../README.md "${arch_dir}/README.md"
        cp ../../../LICENSE "${arch_dir}/LICENSE"
        cd "${arch_dir}"
        zip ${file%%.*}.zip README.md LICENSE ${binary}
        cp ${file%%.*}.zip ../../release
        tar -cvzf ${file%%.*}.tar.gz README.md LICENSE ${binary}
        cp ${file%%.*}.tar.gz ../../release
        cd $dist
    fi
done
go get -u github.com/tcnksm/ghr
ghr -u nebtex -replace $MENSHEND_RELEASE release

cd ..
docker docker login -u $DOCKER_HUB_USER -p $DOCKER_HUB_PASSWORD

docker build -t nebtex/menshend:$MENSHEND_RELEASE .
# upload to dockerhub
docker push nebtex/menshend:$MENSHEND_RELEASE
