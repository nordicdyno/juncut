package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	pretty = flag.Bool("pretty", false, "enable pretty output")
)

func main() {
	flag.Parse()

	var input bytes.Buffer
	_, err := io.Copy(&input, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	result, err := parseAndFix(input.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if !*pretty {
		fmt.Println(string(result))
		return
	}

	prettyResult, err := formatJSON(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(prettyResult))
}

func formatJSON(data []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ") // indent with 2 spaces
	if err != nil {
		return nil, err
	}
	return prettyJSON.Bytes(), nil
}
