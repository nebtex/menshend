
#==============================================================================
#											 					Linters
#==============================================================================

lint:
	gometalinter --disable-all --enable=dupl --enable=errcheck --enable=goconst \
	--enable=golint --enable=gosimple --enable=ineffassign --enable=interfacer \
	--enable=misspell --enable=staticcheck --enable=structcheck --enable=gocyclo \
	--enable=unused --enable=vet --enable=vetshadow --enable=lll \
	--line-length=80 --deadline=60s --vendor --dupl-threshold=100 ./...

install_external_libraries:
	bash scripts/install_tools.sh
	trash -k

update_trash:
	trash -u

deps: install_external_libraries

update_deps: update_trash install_external_libraries

create_test_services:
	docker run -d -e 'CONSUL_LOCAL_CONFIG={"datacenter":"test", "acl_datacenter": "test", "acl_master_token": "test_token", "acl_default_policy": "deny"}' -p 8500:8500 --name consul consul:0.7.2
	docker run -d -e 'SKIP_SETCAP=yes' -e 'VAULT_DEV_ROOT_TOKEN_ID=myroot' -e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200' -p 8200:8200 vault

configure_vault:
	bash scripts/configure_vault.sh


remove_test_services:
ifndef CI
	$ docker stop $(POSTGRES_NAME) && docker rm  $(POSTGRES_NAME)>/dev/null; true
endif

test_env: create_test_services  create_test_dbs

remove_test_env: remove_test_services

.PHONY: deps run_test_services create_test_dbs remove_test_services test_env remove_test_env
