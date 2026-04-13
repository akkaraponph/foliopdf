package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/presspdf"
)

func main() {
	pdfPath := "../pdf/14_markdown.pdf"
	outputDir := "output"

	paths, err := presspdf.ConvertToImages(pdfPath, outputDir,
		presspdf.WithFormat(presspdf.JPEG),
		presspdf.WithDPI(300),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converted %d page(s) to JPEG:\n", len(paths))
	for _, p := range paths {
		fmt.Println(" ", p)
	}
}
