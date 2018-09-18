all: test 

build:
	go build

test:
	GOCACHE=off go test -v
