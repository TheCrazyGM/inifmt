package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestAlignSection(t *testing.T) {
	tests := []struct {
		name            string
		lines           []string
		includeComments bool
	}{
		{
			name:            "basic alignment",
			lines:           []string{"key1=val1", "longerkey=val2", "k=v"},
			includeComments: false,
		},
		{
			name:            "with comments and blanks, exclude comments",
			lines:           []string{"; comment", "key1=val1", "", "longerkey=val2"},
			includeComments: false,
		},
		{
			name:            "with comments and blanks, include comments",
			lines:           []string{"; comment", "key1=val1", "", "longerkey=val2"},
			includeComments: true,
		},
		{
			name:            "no equals sign",
			lines:           []string{"key1", "key2"},
			includeComments: false,
		},
		{
			name:            "empty input",
			lines:           []string{},
			includeComments: false,
		},
		{
			name:            "mixed with and without equals",
			lines:           []string{"key1=val1", "noequals", "key2=val2"},
			includeComments: false,
		},
		{
			name:            "pre-aligned",
			lines:           []string{"key1 = val1", "key2 = val2"},
			includeComments: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alignSection(tt.lines, tt.includeComments)
			assertAligned(t, got)
		})
	}
}

func TestSingleSpaceFormat(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "basic conversion",
			lines: []string{"key1=val1", "key2  =  val2", "key3= val3"},
		},
		{
			name:  "no equals sign",
			lines: []string{"key1", "key2"},
		},
		{
			name:  "already correct",
			lines: []string{"key1 = val1"},
		},
		{
			name:  "empty input",
			lines: []string{},
		},
		{
			name:  "comments and blanks",
			lines: []string{"; comment", "key1=val1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// singleSpaceFormat expects a scanner
			input := strings.Join(tt.lines, "\n")
			if len(tt.lines) > 0 {
				input += "\n"
			}
			scanner := bufio.NewScanner(strings.NewReader(input))
			got, err := singleSpaceFormat(scanner)
			if err != nil {
				t.Fatalf("singleSpaceFormat() unexpected error: %v", err)
			}
			assertSingleSpace(t, got)
		})
	}
}

func TestAlignIni(t *testing.T) {
	tests := []struct {
		name       string
		cfg        formatConfig
		inputLines []string
	}{
		{
			name: "global alignment",
			cfg:  formatConfig{perSection: false, includeComments: false},
			inputLines: []string{
				"key1=val1",
				"[section1]",
				"longkey=val2",
				"k=v",
			},
		},
		{
			name: "per-section alignment",
			cfg:  formatConfig{perSection: true, includeComments: false},
			inputLines: []string{
				"global_key=global_value",
				"[section1]",
				"s1key1=val1",
				"s1longerkey=val2",
				"[section2]",
				"s2key=val3",
				"s2lk=val4",
				"; comment in section 2",
			},
		},
		{
			name: "per-section with comments included",
			cfg:  formatConfig{perSection: true, includeComments: true},
			inputLines: []string{
				"[section1]",
				"; comment1",
				"key1=val1",
				"[section2]",
				"longkey=val2",
				"; comment2",
			},
		},
		{
			name:       "empty input",
			cfg:        formatConfig{perSection: false, includeComments: false},
			inputLines: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.Join(tt.inputLines, "\n")
			if len(tt.inputLines) > 0 {
				input += "\n"
			}
			scanner := bufio.NewScanner(strings.NewReader(input))
			got, err := alignIni(scanner, tt.cfg)
			if err != nil {
				t.Fatalf("alignIni() unexpected error: %v", err)
			}
			if tt.cfg.perSection {
				var block []string
				for _, l := range got {
					trimmed := strings.TrimSpace(l)
					if strings.HasPrefix(trimmed, "[") {
						if len(block) > 0 {
							assertAligned(t, block)
							block = nil
						}
						continue
					}
					block = append(block, l)
				}
				if len(block) > 0 {
					assertAligned(t, block)
				}
			} else {
				assertAligned(t, got)
			}
		})
	}
}

func assertAligned(t *testing.T, lines []string) {
	t.Helper()
	eqMin, eqMax := -1, -1
	for i, l := range lines {
		if !strings.Contains(l, "=") {
			// skip non key/value lines
			continue
		}
		if !strings.Contains(l, " = ") {
			t.Fatalf("line %d not normalized around '=': %q", i, l)
		}
		if strings.HasSuffix(l, " ") {
			t.Fatalf("line %d has trailing spaces: %q", i, l)
		}
		pos := strings.Index(l, "=")
		leading := len(l) - len(strings.TrimLeft(l, " \t"))
		col := pos - leading
		if eqMin == -1 {
			eqMin, eqMax = col, col
		} else {
			if col < eqMin {
				eqMin = col
			}
			if col > eqMax {
				eqMax = col
			}
		}
	}
	if eqMax-eqMin > 1 {
		t.Fatalf("alignment columns vary more than 1 space: min %d max %d", eqMin, eqMax)
	}
}

// assertSingleSpace verifies each line containing '=' has exactly one space on each side and no trailing whitespace.
func assertSingleSpace(t *testing.T, lines []string) {
	t.Helper()
	for i, l := range lines {
		if !strings.Contains(l, "=") {
			continue
		}
		if !strings.Contains(l, " = ") {
			t.Fatalf("line %d does not contain ' = ' delimiter: %q", i, l)
		}
		if strings.Count(l, " = ") != 1 {
			t.Fatalf("line %d contains multiple ' = ' sequences: %q", i, l)
		}
		if strings.HasSuffix(l, " ") {
			t.Fatalf("line %d has trailing space: %q", i, l)
		}
	}
}
