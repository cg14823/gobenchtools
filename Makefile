all:
	mkdir -p build
	go build -o ./build/gobench2json ./cmd/gobench2json

clean:
	rm -r ./build