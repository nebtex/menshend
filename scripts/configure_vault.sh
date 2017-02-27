#!/usr/bin/env bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=myroot
vault auth-enable userpass
vault auth-enable github
vault write auth/userpass/users/kuper password=test ttl=72h
