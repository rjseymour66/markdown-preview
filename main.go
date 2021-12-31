package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	header = `<!DOCTYPE html>
	<html>
		<head>
			<meta http-equiv="content-type" content="text/html; charset=utf-8">
			<title>Markdown Preview Tool</title>
		</head>
	</html>
	<body>
	`
	footer = `
	</body>
	</html>
	`
)

func main() {
	// Parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	flag.Parse()

	// If user did not provide input file, show usage
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filename string) error {
	// Read all the data from the input file into []byte and check for errors
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	// filepath.Base is x-platform and returns the last element of the path (usually the filename)
	outFName := fmt.Sprintf("%s.html", filepath.Base(filename))
	fmt.Println(outFName)

	return saveHTML(outFName, htmlData)
}

func parseContent(input []byte) []byte {
	// Parse the md file through blackfriday and bluemonday
	// to generate a valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	// combine the header, body, and footer constants with buffer
	// create buffer of bytes to write to the file
	var buffer bytes.Buffer

	// Write HTML to bytes buffer
	buffer.WriteString(header)
	buffer.Write(body)
	buffer.WriteString(footer)

	// returns a slice of length b.Len() of unread bytes of buffer
	return buffer.Bytes()
}

// Wrapper around ioutil.WriteFile() func
func saveHTML(outFname string, data []byte) error {
	// Write the bytes to the file with perms that are
	// rw for owner but r for everyone else
	return ioutil.WriteFile(outFname, data, 0644)
}
