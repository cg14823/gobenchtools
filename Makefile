all:
	mkdir -p build
	go build -o ./build/gobench2json ./cmd/gobench2json
	go build -o ./build/gobenchchronos ./cmd/gobenchchronos

clean:
	rm -r ./build