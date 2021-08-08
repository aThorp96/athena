package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

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
	rawPage, err := page.Build()
	if err != nil {
		log.Fatal(err)
	}
	return rawPage
}
