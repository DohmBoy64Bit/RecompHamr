// Command docscheck verifies required durable project documentation and contract terms.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const contractPath = "docs/documentation-contract.json"

type documentationContract struct {
	Required        []string            `json:"required"`
	RequiredContent map[string][]string `json:"required_content"`
}

var exitProcess = os.Exit
var readDocument = os.ReadFile
var statPath = os.Stat

func main() {
	exitProcess(checkContractFile(contractPath, readDocument, statPath))
}

func checkContractFile(path string, read func(string) ([]byte, error), stat func(string) (os.FileInfo, error)) int {
	body, err := read(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "docscheck: cannot read contract %s: %v\n", path, err)
		return 1
	}

	var contract documentationContract
	if err := json.Unmarshal(body, &contract); err != nil {
		fmt.Fprintf(os.Stderr, "docscheck: invalid contract %s: %v\n", path, err)
		return 1
	}
	if len(contract.Required) == 0 {
		fmt.Fprintf(os.Stderr, "docscheck: contract %s contains no required documents\n", path)
		return 1
	}

	code := run(contract.Required, stat)
	if code == 0 {
		code = checkContent(contract.RequiredContent, read)
	}
	return code
}

func checkContent(requirements map[string][]string, read func(string) ([]byte, error)) int {
	missing := false
	for path, terms := range requirements {
		body, err := read(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "docscheck: cannot read %s: %v\n", path, err)
			missing = true
			continue
		}
		for _, term := range terms {
			if !bytes.Contains(body, []byte(term)) {
				fmt.Fprintf(os.Stderr, "docscheck: %s does not document %q\n", path, term)
				missing = true
			}
		}
	}
	if missing {
		return 1
	}
	fmt.Println("docscheck: required documentation contract is satisfied")
	return 0
}

func run(paths []string, stat func(string) (os.FileInfo, error)) int {
	missing := false
	for _, path := range paths {
		info, err := stat(path)
		if err != nil || info.IsDir() || info.Size() == 0 {
			fmt.Fprintf(os.Stderr, "docscheck: missing or empty %s\n", path)
			missing = true
		}
	}
	if missing {
		return 1
	}
	fmt.Println("docscheck: required durable documentation is present")
	return 0
}
