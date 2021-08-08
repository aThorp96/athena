package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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

func handleLookup(w gemini.ResponseWriter, r *gemini.Request) {
	const URLTemplate string = "https://api.dictionaryapi.dev/api/v2/entries/en_US/%s"
	// TODO: Sanitize input
	word := r.URL.RawQuery
	log.Printf("Looking up \"%s\"\n", word)

	resp, err := http.Get(fmt.Sprintf(URLTemplate, word))
	if err != nil {
		log.Fatalln(err)
	}
	// Why is the body the stream?
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	w.Write(definitionToPage(body))
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

func main() {
	cfg := configure()
	lookupHandler := gemini.RequireInputHandler(gemini.HandlerFunc(handleLookup), "Enter seach word")
	rootHandler := gemini.HandlerFunc(handleRoot)

	router := mux.NewMux()
	router.AddRoute("/lookup", lookupHandler)
	router.AddRoute("/", rootHandler)

	ctx := context.Background()
	domainHandler, err := gemini.NewDomainHandlerFromFiles(cfg.domain, cfg.certPath, cfg.keyPath, router)
	if err != nil {
		log.Fatal("Error creating domain handler:", err)
	}

	if err != nil {
		log.Fatal("error creating domain handler B:", err)
	}

	// Start the server for two domains (a.gemini / b.gemini).
	addr := fmt.Sprintf(":%d", cfg.port)
	err = gemini.ListenAndServe(ctx, addr, domainHandler)
	if err != nil {
		log.Fatal("error:", err)
	}
}
