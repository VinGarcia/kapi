
TIME=1s

run:
	go run main.go

test:
	go test ./...

bench:
	go test -bench=. -benchtime=$(TIME)

adapted:
	curl localhost:8765/adapted/42 -H 'brand: Dito'

not-adapted:
	curl localhost:8765/not-adapted/42 -H 'brand: Dito'
