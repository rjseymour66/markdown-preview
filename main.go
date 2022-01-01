package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
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
	defaultTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>{{ .Title }}</title>
</head>
<body>
{{ .Body }}
</body>
</html>`
)

// content type represents the HTML content to add into the template
type content struct {
	Title string
	Body  template.HTML
}

func main() {
	// Parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	tFname := flag.String("t", "", "Alternate template name")
	flag.Parse()

	// If user did not provide input file, show usage
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *tFname, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Read the md file into []byte and turn it into HTML
func run(filename string, tFname string, out io.Writer, skipPreview bool) error {
	// Read all the data from the input file into []byte and check for errors
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// handle the parseContent error
	htmlData, err := parseContent(input, tFname)
	if err != nil {
		return err
	}

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

func parseContent(input []byte, tFname string) ([]byte, error) {
	// Parse the md file through blackfriday and bluemonday
	// to generate a valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	// Parse the contents of the defaultTemplate const into a new Template
	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	// If user provided alternate template file, replace template
	if tFname != "" {
		t, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, err
		}
	}

	// Create content with predefined title and generated body
	c := content{
		Title: "Markdown Preview Tool",
		Body:  template.HTML(body),
	}

	// combine the header, body, and footer constants with buffer
	// create buffer of bytes to write to the file
	var buffer bytes.Buffer

	// Write data to the buffer by executing the template
	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	// returns a slice of length b.Len() of unread bytes of buffer
	return buffer.Bytes(), nil
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
