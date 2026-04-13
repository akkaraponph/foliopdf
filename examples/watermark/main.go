package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/presspdf"
)

func main() {
	input := "../pdf/14_markdown.pdf"

	// 1. Template watermark — centered "DRAFT".
	err := presspdf.WatermarkPDF(input, "output/draft.pdf",
		presspdf.WatermarkTemplate("draft"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/draft.pdf")

	// 2. Confidential with custom opacity.
	err = presspdf.WatermarkPDF(input, "output/confidential.pdf",
		presspdf.WatermarkTemplate("confidential"),
		presspdf.WatermarkOpacity(0.15),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/confidential.pdf")

	// 3. Repeating pattern watermark.
	err = presspdf.WatermarkPDF(input, "output/pattern.pdf",
		presspdf.WatermarkText("COPY"),
		presspdf.WatermarkPattern(180, 180),
		presspdf.WatermarkFontSize(36),
		presspdf.WatermarkOpacity(0.1),
		presspdf.WatermarkColor(200, 0, 0),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/pattern.pdf")

	// 4. Custom position, no rotation.
	err = presspdf.WatermarkPDF(input, "output/custom.pdf",
		presspdf.WatermarkText("Company Inc."),
		presspdf.WatermarkPosition(400, 30),
		presspdf.WatermarkRotation(0),
		presspdf.WatermarkFontSize(14),
		presspdf.WatermarkOpacity(0.4),
		presspdf.WatermarkColor(100, 100, 100),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/custom.pdf")
}
