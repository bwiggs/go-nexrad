package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/svg"
)

func minifySvg(largeInputFile string, outputFile string) {
	// read the input file
	content, err := ioutil.ReadFile(largeInputFile)
	if err != nil {
		log.Fatal(err)
	}
	text := string(content)
	//textSmaller := text[0:100]
	//fmt.Println(textSmaller)

	// minify the large file
	m := minify.New()
	m.Add("image/svg+xml", &svg.Minifier{
		Precision: 10,
	})
	//in := "Because my coffee was too cold, I heated it in the microwave."
	out, err := m.String("image/svg+xml", text)
	if err != nil {
		log.Fatal(err)
	}

	// writes minified content to another file
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(out)
	if err2 != nil {
		log.Fatal(err2)
	}

	// gets the filesize of each file
	// large file
	largeinfo, err := os.Stat(largeInputFile)
	if err != nil {
		log.Fatal(err)
	}
	largeFileSizeInt := largeinfo.Size()
	largeFileSize := HumanFileSize(float64(largeFileSizeInt))
	//largeFileSize := strconv.Itoa(int(largeFileSizeInt))
	// output file
	outputinfo, err := os.Stat(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	outputFileSizeInt := outputinfo.Size()
	outputFileSize := HumanFileSize(float64(outputFileSizeInt))
	//outputFileSize := strconv.Itoa(int(outputFileSizeInt))

	fmt.Println("Minified", largeInputFile, "(" + largeFileSize + ")", "->", outputFile, "(" + outputFileSize + ")")
}

/*
* Convert bytes to human readable format
* https://hakk.dev/docs/golang-convert-file-size-human-readable/
*/
var (
	suffixes [5]string
)
func Round(val float64, roundOn float64, places int ) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
func HumanFileSize(size float64) string {
	//fmt.Println(size)
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"
	
	base := math.Log(size)/math.Log(1024)
	getSize := Round(math.Pow(1024, base - math.Floor(base)), .5, 2)
	//fmt.Println(int(math.Floor(base)))
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64)+" "+string(getSuffix)
}