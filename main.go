// Package main provides a tool for formatting INI-style configuration files.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// config holds the application configuration.
type config struct {
	write       bool
	perSection  bool
	singleSpace bool
}

// formatConfig holds formatting configuration.
type formatConfig struct {
	perSection bool
}

func main() {
	var cfg config
	rootCmd := &cobra.Command{
		Use:   "inifmt [file]",
		Short: "Aligns '=' signs in INI-style files for readability.",
		Long: `inifmt is a tool to neatly align '=' signs in INI-style configuration files.

If a file is provided as an argument, it will be read and formatted.
If no file is provided, input will be read from stdin (e.g., pipe or redirect).

By default, alignment is global (across the whole file).
Use --per-section/-s to align within each section independently.
Use --single-space/-u to remove formatting and ensure only a single space around '='.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cfg, args)
		},
	}

	rootCmd.Flags().BoolVarP(&cfg.write, "write", "w", false, "Write changes back to the file (if file argument is given)")
	rootCmd.Flags().BoolVarP(&cfg.perSection, "per-section", "s", false, "Align '=' within each INI section independently")
	rootCmd.Flags().BoolVarP(&cfg.singleSpace, "single-space", "u", false, "Remove formatting and ensure only a single space around '='")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// run executes the main application logic.
func run(cfg config, args []string) error {
	// Determine input source
	var input io.Reader = os.Stdin
	var filename string
	if len(args) > 0 {
		filename = args[0]
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer file.Close()
		input = file
	}

	// Process input
	scanner := bufio.NewScanner(input)
	var result []string
	var processErr error

	if cfg.singleSpace {
		result, processErr = singleSpaceFormat(scanner)
	} else {
		fc := formatConfig{
			perSection: cfg.perSection,
		}
		result, processErr = alignIni(scanner, fc)
	}

	if processErr != nil {
		return fmt.Errorf("processing input: %w", processErr)
	}

	// Handle output
	if cfg.write && filename != "" {
		if err := writeToFile(filename, result); err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}
	} else {
		if cfg.write {
			fmt.Fprintln(os.Stderr, "[Warning] --write ignored when reading from stdin")
		}
		for _, line := range result {
			fmt.Println(line)
		}
	}

	return nil
}

// writeToFile writes lines to the specified file.
func writeToFile(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := fmt.Fprintln(file, line); err != nil {
			return fmt.Errorf("writing line: %w", err)
		}
	}
	return nil
}

// alignIni aligns INI content according to the given configuration.
func alignIni(scanner *bufio.Scanner, cfg formatConfig) ([]string, error) {
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if len(lines) == 0 { // If all lines were consumed by scanner error or input was empty
		return make([]string, 0), nil
	}

	for i, line := range lines {
		raw := strings.TrimRight(line, " \t")
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") {
			if idx := strings.Index(trimmed, "]"); idx != -1 {
				header := trimmed[:idx+1]
				rest := strings.TrimSpace(trimmed[idx+1:])
				if rest != "" {
					marker := rest[:1]
					text := strings.TrimSpace(rest[1:])
					lines[i] = header + " " + marker + " " + text
				} else {
					lines[i] = header
				}
				continue
			}
		}
		lines[i] = raw
	}

	if !cfg.perSection {
		return alignSection(lines), nil
	}

	result := make([]string, 0, len(lines))
	var sectionLines []string

	flushSection := func() {
		if len(sectionLines) > 0 {
			result = append(result, alignSection(sectionLines)...)
			sectionLines = nil
		}
	}

	for _, line := range lines {
		raw := strings.TrimRight(line, " \t")
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") {
			if idx := strings.Index(trimmed, "]"); idx != -1 {
				flushSection()
				header := trimmed[:idx+1]
				comment := strings.TrimSpace(trimmed[idx+1:])
				if comment != "" {
					result = append(result, header+" "+comment)
				} else {
					result = append(result, header)
				}
				continue
			}
		}
		sectionLines = append(sectionLines, line)
	}
	flushSection()
	// Ensure non-nil return even if all lines were section headers or filtered out
	if result == nil && len(lines) > 0 {
		return make([]string, 0), nil
	}
	return result, nil
}

// alignSection aligns the equals signs in the given lines.
func alignSection(lines []string) []string {
	if len(lines) == 0 {
		return make([]string, 0)
	}

	// First pass – determine the maximum key length (excluding indentation) among lines with '='.
	maxKeyLen := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		before, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(before)
		if l := len(key); l > maxKeyLen {
			maxKeyLen = l
		}
	}

	result := make([]string, 0, len(lines))

	for _, line := range lines {
		original := strings.TrimRight(line, " \t") // drop trailing whitespace
		trimmed := strings.TrimSpace(original)

		// Handle comment / blank lines
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			result = append(result, original)
			continue
		}

		before, after, ok := strings.Cut(original, "=")
		if !ok {
			// Line without '=' – leave as-is (after trimming trailing whitespace)
			result = append(result, original)
			continue
		}

		key := strings.TrimSpace(before)
		// Normalize internal whitespace in value
		right := strings.Join(strings.Fields(after), " ")

		spacesNeeded := max(maxKeyLen-len(key), 0)
		formatted := key + strings.Repeat(" ", spacesNeeded) + " = " + right
		result = append(result, formatted)
	}

	return result
}

// singleSpaceFormat formats lines to have single spaces around '=' and trims trailing whitespace.
func singleSpaceFormat(scanner *bufio.Scanner) ([]string, error) {
	result := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t") // remove trailing spaces
		if before, after, ok := strings.Cut(line, "="); ok {
			left := strings.TrimSpace(before)
			// Normalize internal whitespace in value
			right := strings.Join(strings.Fields(after), " ")
			result = append(result, fmt.Sprintf("%s = %s", left, right))
		} else {
			result = append(result, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	return result, nil
}
