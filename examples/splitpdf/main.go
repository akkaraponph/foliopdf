package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/presspdf"
)

func main() {
	pdfPath := "../pdf/05_toc.pdf"

	// Split every page into a separate PDF (pure Go, no external tools).
	outputDir := "output/pages"
	paths, err := presspdf.SplitPDF(pdfPath, outputDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Split into %d page(s):\n", len(paths))
	for _, p := range paths {
		fmt.Println(" ", p)
	}

	// Split by custom page ranges.
	rangeDir := "output/ranges"
	paths, err = presspdf.SplitPDF(pdfPath, rangeDir,
		presspdf.WithRanges(
			presspdf.PageRange{From: 1, To: 1},
			presspdf.PageRange{From: 2, To: 3},
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nSplit into %d range(s):\n", len(paths))
	for _, p := range paths {
		fmt.Println(" ", p)
	}
}
