deps:
	glide install

update_deps: update_trash install_external_libraries

create_test_services:
	docker run -d -e 'SKIP_SETCAP=yes' -e 'VAULT_DEV_ROOT_TOKEN_ID=myroot' -e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200' -p 8200:8200 vault

configure_vault:
	bash scripts/configure_vault.sh

test:
	bash scripts/test.sh

build:
	bash scripts/build.sh

bundle_react:
	bash scripts/bundle_react.sh

test_env: create_test_services  create_test_dbs

remove_test_env: remove_test_services

.PHONY: deps run_test_services create_test_dbs remove_test_services test_env remove_test_env
