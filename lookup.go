package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/a-h/gemini"
)

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
