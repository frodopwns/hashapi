package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	// set up port flag
	portPtr := flag.String("port", "8080", "port to bind the api server to")
	hostPtr := flag.String("host", "", "hostname to serve")
	certPtr := flag.String("cert", "", "path to ssl crt file")
	keyPtr := flag.String("key", "", "path to ssl key file")
	flag.Parse()

	// this env entity is used to pass shared data to handlers
	env := &Env{
		HashMap: NewHashMap(),
		Stats:   NewStats(),
		wg:      &sync.WaitGroup{},
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
		fmt.Print("Waiting for processing to finish...")
		env.wg.Wait()
		fmt.Println("done")
		os.Exit(0)
	}()

	info := "Running server at %s://localhost:%s\n"
	var err error
	if *certPtr == "" || *keyPtr == "" {
		fmt.Printf(info, "http", *portPtr)
		err = http.ListenAndServe(fmt.Sprintf("%s:%s", *hostPtr, *portPtr), nil)
	} else if *certPtr != "" && *keyPtr != "" {
		fmt.Printf(info, "https", "443")
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:443", *hostPtr), *certPtr, *keyPtr, nil)
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
