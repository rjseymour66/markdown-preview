package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

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
<body>
`
	footer = `
</body>
</html>`
)

func main() {
	// Parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	flag.Parse()

	// If user did not provide input file, show usage
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Read the md file into []byte and turn it into HTML
func run(filename string, out io.Writer, skipPreview bool) error {
	// Read all the data from the input file into []byte and check for errors
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	// filepath.Base is x-platform and returns the last element of the path (usually the filename)
	// outName := fmt.Sprintf("%s.html", filepath.Base(filename))

	// create temporary file and check for errors
	// args: dir where to create file, filename pattern generator
	temp, err := ioutil.TempFile("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	outName := temp.Name()

	fmt.Fprintln(out, outName)

	// check that HTML is generated and saved correctly
	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	// Return nothing if skipPreview is true. HTML is already saved ^^^
	if skipPreview {
		return nil
	}

	// Remove the temp file after the function returns
	defer os.Remove(outName)

	// preview the md file
	return preview(outName)
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

// Uses default OS browser to preview md file fname
func preview(fname string) error {
	// command name and params to execute default browser
	cName := ""
	cParams := []string{}

	// Define executable based on OS
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	// Append filename to params slice
	cParams = append(cParams, fname)

	// Locate executable in PATH
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	// Open the file using the default program
	err = exec.Command(cPath, cParams...).Run()

	// Give the browser some time to open the file before deleting it
	time.Sleep(2 * time.Second)
	return err
}
