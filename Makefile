
TIME=1s
args=
path=./...

GOBIN=$(shell go env GOPATH)/bin

FILEPATH=fasthttp-routing/v2/example/main.go
FILEPATH=fiber/v2/example/main.go
run:
	go run $(FILEPATH)

lint: setup
	@$(GOBIN)/staticcheck $(path) $(args)
	@go vet $(path) $(args)
	@echo "StaticCheck & Go Vet found no problems on your code!"

test: setup
	$(GOBIN)/richgo test $(path) $(args)

bench:
	go test -bench=. -benchtime=$(TIME)

request:
	curl -XPOST localhost:8765/adapted/42?qparam=barbar \
		-H 'Content-Type: application/json' \
		-H 'brand: FakeBrand' \
		-d '{"id":32, "name":"John"}'
	@echo

plain-request:
	curl -XPOST localhost:8765/not-adapted/42?qparam=barbar \
		-H 'Content-Type: application/json' \
		-H 'brand: Dito' \
		-d '{"id":32, "name":"John"}'
	@echo

setup: $(GOBIN)/richgo $(GOBIN)/staticcheck
$(GOBIN)/richgo:
	GO111MODULE=off go get -u github.com/kyoh86/richgo

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
