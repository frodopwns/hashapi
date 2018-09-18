package hashapi

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Route wraps a path and its handler
type Route struct {
	path    string
	handler func(env *Env, w http.ResponseWriter, r *http.Request) error
}

// HashApi contains everything needed to make the magic happen
type HashApi struct {
	Env      *Env
	Port     string
	Host     string
	CertPath string
	KeyPath  string
}

// NewHashApi initializes a new Hash Api pbject
func NewHashApi(port, host, certPath, keyPath string) *HashApi {
	return &HashApi{
		Env: &Env{
			HashMap: NewHashMap(),
			Stats:   NewStats(),
			wg:      &sync.WaitGroup{},
		},
		Port:     port,
		Host:     host,
		CertPath: certPath,
		KeyPath:  keyPath,
	}
}

// Routes registers new routes paths with their handlers
func (h *HashApi) Routes(routes []Route) *HashApi {
	// map routers to handlers
	for _, r := range routes {
		http.Handle(r.path, Handler{h.Env, r.handler})
	}
	return h
}

// IsSSL returns true if we have both a key and crt path
func (h *HashApi) IsSSL() bool {
	return h.KeyPath != "" && h.CertPath != ""
}

// Start sets signal hooks for shutdown and then starts the server
func (h *HashApi) Start() error {
	// handle shutdown commands
	var stopSigChan = make(chan os.Signal)
	signal.Notify(stopSigChan, syscall.SIGTERM)
	signal.Notify(stopSigChan, syscall.SIGINT)
	go func() {
		sig := <-stopSigChan
		fmt.Printf("caught sig: %+v\n", sig)
		// tell the route handlers to stop taking new requests
		h.Env.Terminating = true
		fmt.Print("Waiting for processing to finish...")
		h.Env.wg.Wait()
		fmt.Println("done")
		os.Exit(0)
	}()

	info := "Running server at %s://localhost:%s\n"
	var err error
	if h.IsSSL() {
		fmt.Printf(info, "https", "443")
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:443", h.Host), h.CertPath, h.KeyPath, nil)
	} else {
		fmt.Printf(info, "http", h.Port)
		err = http.ListenAndServe(fmt.Sprintf("%s:%s", h.Host, h.Port), nil)
	}

	return err
}

// NewServer returns a Hash Api instance with routes preconfigured
func NewServer(port, host, certPath, keyPath string) *HashApi {
	hapi := NewHashApi(
		port,
		host,
		certPath,
		keyPath,
	).Routes([]Route{
		{"/hash", hashHandler},
		{"/hash/", hashHandler},
		{"/stats", statsHandler},
		{"/stats/", statsHandler},
		{"/shutdown", shutdownHandler},
		{"/shutdown/", shutdownHandler},
	})

	return hapi
}
