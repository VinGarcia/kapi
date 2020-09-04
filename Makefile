
TIME=1s
args=
path=./...

GOPATH=$(shell go env GOPATH)

run:
	go run main.go

test: setup
	$(GOPATH)/bin/richgo test $(path) $(args)

bench:
	go test -bench=. -benchtime=$(TIME)

adapted:
	curl -XPOST localhost:8765/adapted/42?qparam=barbar \
		-H 'Content-Type: application/json' \
		-H 'brand: Dito' \
		-d '{"id":32, "name":"John"}'
	@echo

not-adapted:
	curl localhost:8765/not-adapted/42 -H 'brand: Dito'

setup: .make.setup
.make.setup:
	go get github.com/kyoh86/richgo
	touch .make.setup

