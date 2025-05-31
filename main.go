package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func alignIni(input *bufio.Scanner, perSection bool, includeComments bool) []string {
	var lines []string
	for input.Scan() {
		lines = append(lines, input.Text())
	}

	if !perSection {
		return alignSection(lines, includeComments)
	}

	// Per-section alignment
	var output []string
	var sectionLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// Flush previous section
			if len(sectionLines) > 0 {
				output = append(output, alignSection(sectionLines, includeComments)...)
				sectionLines = nil
			}
			output = append(output, line)
		} else {
			sectionLines = append(sectionLines, line)
		}
	}
	if len(sectionLines) > 0 {
		output = append(output, alignSection(sectionLines, includeComments)...)
	}
	return output
}

func alignSection(lines []string, includeComments bool) []string {
	maxPos := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !includeComments && (trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#")) {
			continue
		}
		pos := strings.Index(line, "=")
		if pos > maxPos {
			maxPos = pos
		}
	}
	var output []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !includeComments && (trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#")) {
			output = append(output, line)
			continue
		}
		pos := strings.Index(line, "=")
		if pos > 0 {
			spacesNeeded := maxPos - pos
			left := line[:pos]
			right := line[pos+1:]
			alignedLine := fmt.Sprintf("%s%*s=%s", left, spacesNeeded, "", right)
			output = append(output, alignedLine)
		} else {
			output = append(output, line)
		}
	}
	return output
}

func singleSpaceFormat(scanner *bufio.Scanner) []string {
	var output []string
	for scanner.Scan() {
		line := scanner.Text()
		pos := strings.Index(line, "=")
		if pos > 0 {
			left := strings.TrimSpace(line[:pos])
			right := strings.TrimSpace(line[pos+1:])
			output = append(output, left+" = "+right)
		} else {
			output = append(output, line)
		}
	}
	return output
}


var writeFlag bool
var perSectionFlag bool
var singleSpaceFlag bool
var includeCommentsFlag bool

var rootCmd = &cobra.Command{
	Use:   "inifmt [file]",
	Short: "Aligns '=' signs in INI-style files for readability.",
	Long: `inifmt is a tool to neatly align '=' signs in INI-style configuration files.

If a file is provided as an argument, it will be read and formatted.
If no file is provided, input will be read from stdin (e.g., pipe or redirect).

By default, comments and blank lines are not included in alignment (output as-is). Use --include-comments/-C to include them in alignment.
By default, alignment is global (across the whole file). Use --per-section/-s to align within each section independently.
Use --single-space/-u to remove formatting and ensure only a single space around '='.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var scanner *bufio.Scanner
		var fromFile bool
		var filename string
		if len(args) == 1 {
			filename = args[0]
			file, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()
			scanner = bufio.NewScanner(file)
			fromFile = true
		} else {
			scanner = bufio.NewScanner(os.Stdin)
			fromFile = false
		}
		var output []string
		if singleSpaceFlag {
			output = singleSpaceFormat(scanner)
		} else {
			output = alignIni(scanner, perSectionFlag, includeCommentsFlag)
		}

		if writeFlag {
			if fromFile {
				f, err := os.Create(filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()
				for _, line := range output {
					_, err := f.WriteString(line + "\n")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
						os.Exit(1)
					}
				}
			} else {
				fmt.Fprintln(os.Stderr, "[Warning] --write ignored when reading from stdin.")
				for _, line := range output {
					fmt.Println(line)
				}
			}
		} else {
			for _, line := range output {
				fmt.Println(line)
			}
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&writeFlag, "write", "w", false, "Write changes back to the file (if file argument is given)")
	rootCmd.Flags().BoolVarP(&perSectionFlag, "per-section", "s", false, "Align '=' within each INI section independently")
	rootCmd.Flags().BoolVarP(&singleSpaceFlag, "single-space", "u", false, "Remove formatting and ensure only a single space around '='")
	rootCmd.Flags().BoolVarP(&includeCommentsFlag, "include-comments", "C", false, "Include comments and blank lines in alignment (default: false)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
