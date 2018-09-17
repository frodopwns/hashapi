package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// create constants for error messages
const (
	NoHashIdErr         = "must include a hash id after /hash/"
	BadHadhIdErr        = "id value after /hash/ must be an integer"
	NoPayloadPresentErr = "a form field 'password' must be included withthe POST"
	HashIdNotFound      = "no hash with that id was found"
)

// Error represents a handler error. It makes it easier and cleaner to return errors from handlers.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (s StatusError) Error() string {
	return fmt.Sprintf("%d: %s", s.Code, s.Err.Error())
}

// Returns our HTTP status code.
func (s StatusError) Status() int {
	return s.Code
}

// wrap the usual go handler so we can add variables
type Handler struct {
	env *Env
	H   func(env *Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// start tracking time for the request
	start := time.Now()

	// if we are in a terminating state, stop user here
	if h.env.Terminating {
		http.Error(w, "Shutting down...", http.StatusServiceUnavailable)
		return
	}

	// call the handler with added vars
	err := h.H(h.env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
		return
	}

	// if this is a post then we need to update stats
	if r.Method == "POST" {
		runtime := time.Now().Sub(start).Seconds() * 1e6
		h.env.Stats.Update(runtime)
	}
}

// handle the /hash endpoint
func hashHandler(env *Env, w http.ResponseWriter, req *http.Request) error {

	// handle POSTS here
	if req.Method == "POST" {

		// get password value from post
		pw := req.PostFormValue("password")
		if pw == "" {
			return StatusError{http.StatusBadRequest, fmt.Errorf(NoPayloadPresentErr)}
		}

		// reserve hash index key
		id := env.HashMap.Save("pending")

		// spin off a go routine to hash the password in 5 seconds
		go func(id int, pw string) {
			time.Sleep(5 * time.Second)
			env.HashMap.Update(id, HashMe(pw))
		}(id, pw)

		fmt.Fprintln(w, id)
	}

	// handle GETS here
	if req.Method == "GET" {
		// this is the mini length a request path must have before it matters
		pathlen := len("/hash/")
		// if the url is too short to contain a hash id, bail
		if len(req.URL.Path) <= pathlen {
			// return a 400
			return StatusError{http.StatusBadRequest, fmt.Errorf(NoHashIdErr)}
		}
		// parse id string from url path
		idStr := strings.TrimSuffix(req.URL.Path[pathlen:], "/")
		// convert to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return StatusError{http.StatusBadRequest, fmt.Errorf(BadHadhIdErr)}

		}

		// get the hash from storage
		value, ok := env.HashMap.Get(id)
		if !ok {
			return StatusError{http.StatusNotFound, fmt.Errorf(HashIdNotFound)}
		}

		// send hash to the user
		fmt.Fprintln(w, value)

	}
	return nil
}

// statsHandler returns the json repr of the current app stats
func statsHandler(env *Env, w http.ResponseWriter, req *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, env.Stats.JSON())
	return nil
}

// shutdownHandler handler calls to /shutdown
func shutdownHandler(env *Env, w http.ResponseWriter, req *http.Request) error {
	// stop handling new requests
	env.Terminating = true
	// be polite
	fmt.Fprintln(w, "goodbye!")
	// trigger our main shutdown code via signal
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	return nil
}

// HashMe returns the base64 encoded SHA512 of th einput string
func HashMe(raw string) string {
	h512 := sha512.New()
	fmt.Fprint(h512, raw)
	return base64.URLEncoding.EncodeToString(h512.Sum(nil))
}
