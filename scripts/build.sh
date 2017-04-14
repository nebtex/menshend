#!/usr/bin/env bash
# set version
if [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  echo "building binaries and docker images"
else
  echo "Skipping build"
  exit 0
fi

if [ "$CI" == "true" ]; then

else
  MENSHEND_RELEASE=latest
fi

go get github.com/mitchellh/gox
cd cmd/menshend
mkdir -p dist
curl -fsL https://github.com/gliderlabs/sigil/releases/download/v0.4.0/sigil_0.4.0_Linux_x86_64.tgz  | tar -zx
mv sigil dist
./dist/sigil -p -f version.tmpl MENSHEND_RELEASE=$MENSHEND_RELEASE > version.go
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
ghr -u nebtex -replace latest release


## build docker image and publish

