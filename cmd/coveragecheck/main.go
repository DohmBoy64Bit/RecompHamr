// Command coveragecheck enforces complete Go statement coverage.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var exitProcess = os.Exit

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: coveragecheck <coverprofile>")
		exitProcess(2)
		return
	}
	exitProcess(check(os.Args[1]))
}

func check(path string) int {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coveragecheck: %v\n", err)
		return 1
	}
	defer f.Close()

	total, covered := 0, 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			fmt.Fprintf(os.Stderr, "coveragecheck: malformed row %q\n", line)
			return 1
		}
		statements, parseErr := strconv.Atoi(fields[1])
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "coveragecheck: malformed statement count %q\n", fields[1])
			return 1
		}
		count, parseErr := strconv.Atoi(fields[2])
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "coveragecheck: malformed execution count %q\n", fields[2])
			return 1
		}
		total += statements
		if count > 0 {
			covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "coveragecheck: read profile: %v\n", err)
		return 1
	}
	if total == 0 {
		fmt.Fprintln(os.Stderr, "coveragecheck: profile contains no statements")
		return 1
	}
	percentage := 100 * float64(covered) / float64(total)
	if covered != total {
		fmt.Fprintf(os.Stderr, "coveragecheck: %.1f%% (%d/%d statements), require 100.0%%\n", percentage, covered, total)
		return 1
	}
	fmt.Printf("coveragecheck: 100.0%% (%d/%d statements)\n", covered, total)
	return 0
}
