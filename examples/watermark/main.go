package main

import (
	"fmt"
	"log"

	"github.com/akkaraponph/foliopdf"
)

func main() {
	input := "../pdf/14_markdown.pdf"

	// 1. Template watermark — centered "DRAFT".
	err := foliopdf.WatermarkPDF(input, "output/draft.pdf",
		foliopdf.WatermarkTemplate("draft"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/draft.pdf")

	// 2. Confidential with custom opacity.
	err = foliopdf.WatermarkPDF(input, "output/confidential.pdf",
		foliopdf.WatermarkTemplate("confidential"),
		foliopdf.WatermarkOpacity(0.15),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/confidential.pdf")

	// 3. Repeating pattern watermark.
	err = foliopdf.WatermarkPDF(input, "output/pattern.pdf",
		foliopdf.WatermarkText("COPY"),
		foliopdf.WatermarkPattern(180, 180),
		foliopdf.WatermarkFontSize(36),
		foliopdf.WatermarkOpacity(0.1),
		foliopdf.WatermarkColor(200, 0, 0),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/pattern.pdf")

	// 4. Custom position, no rotation.
	err = foliopdf.WatermarkPDF(input, "output/custom.pdf",
		foliopdf.WatermarkText("Company Inc."),
		foliopdf.WatermarkPosition(400, 30),
		foliopdf.WatermarkRotation(0),
		foliopdf.WatermarkFontSize(14),
		foliopdf.WatermarkOpacity(0.4),
		foliopdf.WatermarkColor(100, 100, 100),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created output/custom.pdf")
}
