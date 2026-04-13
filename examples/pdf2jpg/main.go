package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/foliopdf"
)

func main() {
	pdfPath := "../pdf/14_markdown.pdf"
	outputDir := "output"

	paths, err := foliopdf.ConvertToImages(pdfPath, outputDir,
		foliopdf.WithFormat(foliopdf.JPEG),
		foliopdf.WithDPI(300),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converted %d page(s) to JPEG:\n", len(paths))
	for _, p := range paths {
		fmt.Println(" ", p)
	}
}
