#!/usr/bin/env bash

/bin/sleep 10
export VAULT_TOKEN=myroot
/bin/menshend admin apply -api http://menshend:8787/v1 -f /services/frontend-branch.yml
/bin/menshend admin apply -api http://menshend:8787/v1 -f /services/frontend-branch-2.yml
/bin/menshend admin apply -api http://menshend:8787/v1 -f /services/terminal.yml
