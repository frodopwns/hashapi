package main

import (
	"flag"
	"log"

	"github.com/frodopwns/hashapi/pkg/hashapi"
)

func main() {

	// set up port flag
	portPtr := flag.String("port", "8080", "port to bind the api server to")
	hostPtr := flag.String("host", "", "hostname to serve")
	certPtr := flag.String("cert", "", "path to ssl crt file")
	keyPtr := flag.String("key", "", "path to ssl key file")
	flag.Parse()

	api := hashapi.NewServer(
		*portPtr,
		*hostPtr,
		*certPtr,
		*keyPtr,
	)
	log.Fatal(api.Start())

}
