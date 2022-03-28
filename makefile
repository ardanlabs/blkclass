SHELL := /bin/bash

# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor
