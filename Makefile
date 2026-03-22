build:
	go build -o randomizer .

sim:
	go test -run TestSimulate -v -count=1

run: build
	./randomizer
