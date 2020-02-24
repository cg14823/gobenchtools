package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cg14823/gobenchtools"
)

func main() {
	var fileIn, fileOut, packageStrip, suiteName string
	flag.StringVar(&fileIn, "bench-file", "", "A file containing the output of benchmarks")
	flag.StringVar(&fileOut, "out", "", "The file to write the parsed values to")
	flag.StringVar(&packageStrip, "pkg-strip", "github.com/", "Remove prefix of package name")
	flag.StringVar(&suiteName, "suite-name", "", "The name to give the test-suite")
	flag.Parse()

	if fileIn == "" {
		fmt.Println("-bench-file is required")
		os.Exit(1)
	}

	out, err := gobenchtools.ParseToXML(fileIn, suiteName, packageStrip)
	if err != nil {
		fmt.Println("Could not parse bench output:", err)
		os.Exit(1)
	}

	xmlOut, err := xml.MarshalIndent(out, "", "    ")
	if err != nil {
		fmt.Println("Could not produce xml bench output:", err)
		os.Exit(1)
	}

	if fileOut == "" {
		fmt.Println(string(xmlOut))
		os.Exit(0)
	}

	err = ioutil.WriteFile(fileOut, xmlOut, gobenchtools.DefaultFileMode)
	if err != nil {
		fmt.Println("Could not produce xml output file:", err)
		os.Exit(1)
	}
}
