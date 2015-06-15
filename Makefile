SHELL := /bin/bash

options:
	echo "please decide to clean build-debug, build or run"
clean:
	rm -f bin/ssiege
	rm -f bin/ssiege-debug
	rm -f benchmark/bindata.go

build-debug:
	make clean
	go-bindata -debug=true -pkg="benchmark" -o benchmark/bindata.go benchmark/frontend/*
	go build -o bin/ssiege-debug ssiege.go
build:
	make clean
	go-bindata -debug=false -pkg="benchmark" -o benchmark/bindata.go benchmark/frontend/*
	go build -o bin/ssiege ssiege.go
run-debug:
	make build-debug
	./bin/ssiege-debug examples/github.json examples/bitbucket.json
run:
	make build
	./bin/ssiege examples/github.json examples/bitbucket.json
