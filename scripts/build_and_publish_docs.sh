#!/usr/bin/env bash
cd docs
gitbook install
gitbook build
cd _book
git init
git commit --allow-empty -m 'Update docs'
git checkout -b gh-pages
git add .
git commit -am 'Update docs'
git push git@github.com:nebtex/menshend gh-pages --force
