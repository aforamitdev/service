SHELL:=/bin/bash

VERSION := 1.0


tidy:
	go mod tidy
	go mod vendor
 
run:
	go run app/sales-api/main.go