package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/folio"
)

func main() {
	pdfPath := "../pdf/14_markdown.pdf"
	outputDir := "output"

	paths, err := folio.ConvertToImages(pdfPath, outputDir,
		folio.WithFormat(folio.JPEG),
		folio.WithDPI(300),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converted %d page(s) to JPEG:\n", len(paths))
	for _, p := range paths {
		fmt.Println(" ", p)
	}
}
