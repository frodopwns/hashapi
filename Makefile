all: test build run 

build:
	go build

test:
	GOCACHE=off go test -v pkg/hashapi/*.go

run:
	./hashapi

docker-build:
	docker build -t hashapi .

docker-run:
	docker run -it -p 8080:8080 hashapi