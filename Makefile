BASENAME = $(shell basename $(shell pwd))

compile:
	docker run --rm -v "$(shell pwd)":/code --mount type=volume,source="$(BASENAME)_cache",target=/code/target --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry cosmwasm/workspace-optimizer:0.12.11

clippy:
	cargo clippy --fix
	cargo fmt

ictest-basic:	
# compile it here then do:
	cd test && go test -race -v -run TestContract .

ictest-whitelist:	
# compile it here then do:
	cd test && go test -race -v -run TestWhitelist .