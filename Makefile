SHELL:=/bin/bash

VERSION := 1.0

all:sales-api

sales-api:
	docker build \
	-f zarf/docker/dockerfile.sales-api \
	-t sales-api-amd64:1.0 \
	--build-arg VCS_REF=`git rev-parse HEAD` \
	--build-arg  BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%M:%SZ"` \
	.


tidy:
	go mod tidy
	go mod vendor
 
run:
	go run app/sales-api/main.go