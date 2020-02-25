all:
	mkdir -p build
	go build -o ./build/gobench2json ./cmd/gobench2json
	go build -o ./build/gobench2xml ./cmd/gobench2xml
	go build -o ./build/gobenchchronos ./cmd/gobenchchronos

install:
	go install ./cmd/gobench2json
	go install ./cmd/gobench2xml
	go install ./cmd/gobenchchronos

clean:
	rm -r ./build
