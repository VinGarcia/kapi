
TIME=1s
args=
path=./...

GOPATH=$(shell go env GOPATH)

run:
	go run cmd/main.go

test: setup
	$(GOPATH)/bin/richgo test $(path) $(args)

bench:
	go test -bench=. -benchtime=$(TIME)

request:
	curl -XPOST localhost:8765/adapted/42?qparam=barbar \
		-H 'Content-Type: application/json' \
		-H 'brand: Dito' \
		-d '{"id":32, "name":"John"}'
	@echo

plain-request:
	curl -XPOST localhost:8765/not-adapted/42?qparam=barbar \
		-H 'Content-Type: application/json' \
		-H 'brand: Dito' \
		-d '{"id":32, "name":"John"}'
	@echo

setup: .make.setup
.make.setup:
	go get github.com/kyoh86/richgo
	touch .make.setup

