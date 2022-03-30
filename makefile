SHELL := /bin/bash

# Wallets
# Kennedy: 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32
# Pavel: 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4
# Ceasar: 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76
# Baba: 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9
# Ed: 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0
# Miner1: 0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8
# Miner2: 0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61

# curl -il -X GET http://localhost:8080/v1/genesis
# curl -il -X GET http://localhost:8080/v1/accounts/list
# curl -il -X GET http://localhost:8080/v1/tx/uncommitted/list

# go run app/wallet/cli/main.go -t "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9" -n 1 -v 100 -p 15
# go run app/wallet/cli/main.go -t "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76" -n 2 -v 200 -p 15
# go run app/wallet/cli/main.go -t "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0" -n 3 -v 450 -p 15
# go run app/wallet/cli/main.go -t "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76" -n 4 -v 230 -p 15

# ==============================================================================
# Local support

up:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7181 --web-public-host 0.0.0.0:8180 --web-private-host 0.0.0.0:9180 --node-miner-name=miner2 --node-db-path zblock/blocks2.db | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep "main -race" | grep -v grep | sed -n 1,1p | cut -c1-5)

clear-db:
	cat /dev/null > zblock/blocks.db

admin:
	go run app/wallet/cli/main.go	

# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor
