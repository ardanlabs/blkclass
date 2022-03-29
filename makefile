SHELL := /bin/bash

# curl -il -X GET http://localhost:8080/v1/genesis
# curl -il -X GET http://localhost:8080/v1/accounts/list
# curl -il -X GET http://localhost:8080/v1/tx/uncommitted/list

# curl -il -X POST http://localhost:8080/v1/tx/submit -d '{"nonce": 1, "from": "bill", "to": "maddie", "value": 300, "tip": 10}'
# curl -il -X POST http://localhost:8080/v1/tx/submit -d '{"nonce": 2, "from": "bill", "to": "betty", "value": 200, "tip": 20}'
# curl -il -X POST http://localhost:8080/v1/tx/submit -d '{"nonce": 3, "from": "bill", "to": "carlos", "value": 150, "tip": 30}'
# curl -il -X POST http://localhost:8080/v1/tx/submit -d '{"nonce": 4, "from": "bill", "to": "ed", "value": 400, "tip": 40}'


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
