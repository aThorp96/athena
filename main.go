package main

import (
	"context"
	"fmt"
	"log"

	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
	"github.com/spf13/pflag"
)

// Configuration struct.
type athenaConfig struct {
	domain   string
	port     uint
	certPath string
	keyPath  string
}

func configure() athenaConfig {
	const domainHelp string = "Domain to serve. Should match the certificate."
	const portHelp string = "Port to listen on. If not fronted by a proxy server, should be 1965."
	const certPathHelp string = "Path to certificate file."
	const keyPathHelp string = "Path to key file."

	config := athenaConfig{}

	pflag.StringVarP(&config.domain, "domain", "d", "localhost", domainHelp)
	pflag.UintVarP(&config.port, "port", "p", 1965, portHelp)
	pflag.StringVarP(&config.certPath, "certPath", "c", "server.crt", certPathHelp)
	pflag.StringVarP(&config.keyPath, "keyPath", "k", "server.key", keyPathHelp)

	pflag.Parse()

	return config
}

func handleRoot(w gemini.ResponseWriter, r *gemini.Request) {
	page := newAthenaDocument()
	page.AddLine("Search for a word to continue")
	rawPage, err := page.Build()
	if err != nil {
		log.Fatal(err)
	}
	w.Write(rawPage)
}

// routes is a wrapper to map the route strings to their respective functions.
func routes() map[string]gemini.HandlerFunc {
	lookupHandler := gemini.RequireInputHandler(gemini.HandlerFunc(handleLookup), "Enter seach word")
	return map[string]gemini.HandlerFunc{
		// Add new routes here
		"/":       gemini.HandlerFunc(handleRoot),
		"/lookup": lookupHandler.(gemini.HandlerFunc),
	}
}

func main() {
	// Configure app
	cfg := configure()

	// Initialize routes
	router := mux.NewMux()
	for route, handler := range routes() {
		router.AddRoute(route, handler)
	}

	ctx := context.Background()
	domainHandler, err := gemini.NewDomainHandlerFromFiles(cfg.domain, cfg.certPath, cfg.keyPath, router)
	if err != nil {
		log.Fatal("Error creating domain handler:", err)
	}

	addr := fmt.Sprintf(":%d", cfg.port)
	err = gemini.ListenAndServe(ctx, addr, domainHandler)
	if err != nil {
		log.Fatal("error:", err)
	}
}
