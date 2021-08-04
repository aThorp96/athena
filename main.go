package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

type Definition struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example"`
	Synonyms   []string `json:"synonyms"`
}
type Meanings struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
}
type Word struct {
	Word      string              `json:"word"`
	Phonetics []map[string]string `json:"phonetics"`
	Meanings  []Meanings          `json:"meanings"`
}

func getFileContent(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	content := ""
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return content
}

func newAthenaDocument() gemini.DocumentBuilder {
	page := gemini.NewDocumentBuilder()
	page.SetHeader(getFileContent("./public/header.gmi"))
	page.AddLine("---")
	page.SetFooter(getFileContent("./public/footer.gmi"))
	return page
}

func definitionToPage(data []byte) []byte {
	page := newAthenaDocument()

	var payload []Word
	err := json.Unmarshal(data, &payload)
	if err != nil {
		page.AddLine("An unknown error has occored...")
		log.Println("Error unmarshalling response: %s", err)
	}

	for i := 0; i < len(payload); i++ {
		word := payload[i]

		page.AddLine("")
		page.AddH1Header(strings.Title(word.Word))

		for j := 0; j < len(word.Phonetics); j++ {
			phonetic := word.Phonetics[j]
			page.AddLine(fmt.Sprintf("\t(%s)", phonetic["text"]))
		}

		page.AddLine("")

		for j := 0; j < len(word.Meanings); j++ {
			meaning := word.Meanings[j]
			page.AddLine(meaning.PartOfSpeech)

			for k := 0; k < len(meaning.Definitions); k++ {
				definition := meaning.Definitions[k]
				page.AddH2Header("Definition:")
				page.AddLine(definition.Definition)
				if len(definition.Example) > 0 {
					page.AddH3Header("Example:")
					page.AddLine(definition.Example)
				}
				page.AddLine("")
			}
		}
	}
	return page.Build()
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
	w.Write(page.Build())
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
