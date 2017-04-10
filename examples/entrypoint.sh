#!/usr/bin/env bash

/bin/sleep 10
export VAULT_TOKEN=myroot
./menshend admin apply -api http://localhost:8787/v1 -f frontend-branch.yml
./menshend admin apply -api http://localhost:8787/v1 -f frontend-branch-2.yml
./menshend admin apply -api http://localhost:8787/v1 -f terminal.yml
