package main

import (
	"bufio"
	"log"
	"os"

	"github.com/a-h/gemini"
)

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
