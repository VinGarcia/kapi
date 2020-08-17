
run:
	go run main.go

adapted:
	curl localhost:8765/adapted/42 -H 'brand: Dito'

not-adapted:
	curl localhost:8765/not-adapted/42 -H 'brand: Dito'
