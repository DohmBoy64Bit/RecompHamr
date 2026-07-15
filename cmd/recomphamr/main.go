// Command recomphamr starts the barebones RecompHamr terminal application.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DohmBoy64Bit/RecompHamr/internal/app/terminal"
)

var version = "dev"

var (
	startApplication     = terminal.Run
	printApplicationHelp = terminal.PrintHelp
	exitProcess          = os.Exit
)

func main() {
	handleRunResult(run(os.Args[1:], os.Stdout))
}

func handleRunResult(err error) {
	if err != nil {
		log.Printf("recomphamr: %v", err)
		exitProcess(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "-v", "--version", "version":
			_, err := fmt.Fprintln(stdout, "recomphamr", version)
			return err
		case "-h", "--help", "help":
			return printApplicationHelp(stdout)
		}
	}
	return startApplication(stdout, version)
}
