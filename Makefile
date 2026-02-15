build:
	go build -o randomizer .

run: build
	./randomizer
