#!/usr/bin/env bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=myroot
vault mount consul
vault write consul/config/access address=consul:8500 token=test_token
POLICY='key "kuper" { policy = "write" }'
echo $POLICY | base64 | vault write consul/roles/admin policy=-
POLICY='key "kuper/Roles/devops" { policy = "read" }\nkey "kuper/Roles/frontend" { policy = "read" }'
echo $POLICY | base64 | vault write consul/roles/devops policy=-
POLICY='key "kuper/Roles/frontend" { policy = "read" }'
echo $POLICY | base64 | vault write consul/roles/fronted policy=-
