package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// set up port flag
	portPtr := flag.String("port", "8080", "port to bind the api server to")
	flag.Parse()

	// this env entity is used to pass shared data to handlers
	env := &Env{
		HashMap: NewHashMap(),
		Stats:   NewStats(),
	}

	// map routers to handlers
	http.Handle("/hash", Handler{env, hashHandler})
	http.Handle("/hash/", Handler{env, hashHandler})
	http.Handle("/stats", Handler{env, statsHandler})
	http.Handle("/stats/", Handler{env, statsHandler})
	http.Handle("/shutdown", Handler{env, shutdownHandler})

	// handle shutdown commands
	var stopSigChan = make(chan os.Signal)
	signal.Notify(stopSigChan, syscall.SIGTERM)
	signal.Notify(stopSigChan, syscall.SIGINT)
	go func() {
		sig := <-stopSigChan
		fmt.Printf("caught sig: %+v\n", sig)
		// tell the route handlers to stop taking new requests
		env.Terminating = true
		// wait 6 seconds because that should sllow hash processes to finish
		fmt.Println("Wait for 6 second to finish processing")
		time.Sleep(6 * time.Second)
		os.Exit(0)
	}()

	fmt.Printf("Running server at http://localhost:%s\n", *portPtr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *portPtr), nil))
}
