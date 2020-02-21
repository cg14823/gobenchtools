package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cg14823/gobenchtools"
)

func main() {
	var fileIn, fileOut string
	flag.StringVar(&fileIn, "bench-file", "", "A file containing the output of benchmarks")
	flag.StringVar(&fileOut, "out", "", "The file to write the parsed values to")
	flag.Parse()

	if fileIn == "" {
		fmt.Println("-bench-file is required")
		os.Exit(1)
	}

	out, err := gobenchtools.ParseFile(fileIn)
	if err != nil {
		fmt.Println("Could not parse input: ", err.Error())
		os.Exit(1)
	}

	err = outputParsed(fileOut, out)
	if err != nil {
		fmt.Println("Could not write output: ", err.Error())
		os.Exit(1)
	}
}

func outputParsed(fileOut string, parsed gobenchtools.ParsedBench) error {
	out, err := json.Marshal(&parsed)
	if err != nil {
		return nil
	}

	// output to stdout
	if fileOut == "" {
		fmt.Println(string(out))
		return nil
	}

	return ioutil.WriteFile(fileOut, out, gobenchtools.DefaultFileMode)
}
