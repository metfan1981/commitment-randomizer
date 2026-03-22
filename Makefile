build:
	go build -o randomizer .

test-sim:
	go test -run TestSimulate -v -count=1

run: build
	./randomizer
