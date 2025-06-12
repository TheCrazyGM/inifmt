package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestAlignSection(t *testing.T) {
	tests := []struct {
		name            string
		lines           []string
		includeComments bool
		want            []string
	}{
		{
			name:            "basic alignment",
			lines:           []string{"key1=val1", "longerkey=val2", "k=v"},
			includeComments: false,
			want:            []string{"key1     =val1", "longerkey=val2", "k        =v"},
		},
		{
			name:            "with comments and blanks, exclude comments",
			lines:           []string{"; comment", "key1=val1", "", "longerkey=val2"},
			includeComments: false,
			want:            []string{"; comment", "key1     =val1", "", "longerkey=val2"},
		},
		{
			name:            "with comments and blanks, include comments",
			lines:           []string{"; comment", "key1=val1", "", "longerkey=val2"},
			includeComments: true,
			want:            []string{"; comment  ", "key1     =val1", "           ", "longerkey=val2"}, // Note: behavior for aligning pure comment/blank lines might be specific
		},
		{
			name:            "no equals sign",
			lines:           []string{"key1", "key2"},
			includeComments: false,
			want:            []string{"key1", "key2"},
		},
		{
			name:            "empty input",
			lines:           []string{},
			includeComments: false,
			want:            []string{},
		},
		{
			name:            "mixed with and without equals",
			lines:           []string{"key1=val1", "noequals", "key2=val2"},
			includeComments: false,
			want:            []string{"key1=val1", "noequals", "key2=val2"},
		},
		{
			name:            "pre-aligned",
			lines:           []string{"key1 = val1", "key2 = val2"},
			includeComments: false,
			want:            []string{"key1 = val1", "key2 = val2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alignSection(tt.lines, tt.includeComments)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("alignSection() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSingleSpaceFormat(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "basic conversion",
			lines: []string{"key1=val1", "key2  =  val2", "key3= val3"},
			want:  []string{"key1 = val1", "key2 = val2", "key3 = val3"},
		},
		{
			name:  "no equals sign",
			lines: []string{"key1", "key2"},
			want:  []string{"key1", "key2"},
		},
		{
			name:  "already correct",
			lines: []string{"key1 = val1"},
			want:  []string{"key1 = val1"},
		},
		{
			name:  "empty input",
			lines: []string{},
			want:  []string{},
		},
		{
			name:  "comments and blanks",
			lines: []string{"; comment", "key1=val1", ""},
			want:  []string{"; comment", "key1 = val1", ""},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("singleSpaceFormat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlignIni(t *testing.T) {
	tests := []struct {
		name       string
		cfg        formatConfig
		inputLines []string
		wantLines  []string
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
			wantLines: []string{
				"key1   =val1",
				"[section1]",
				"longkey=val2",
				"k      =v",
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
			wantLines: []string{
				"global_key=global_value", // Aligned with itself as it's before any section
				"[section1]",
				"s1key1     =val1",
				"s1longerkey=val2",
				"[section2]",
				"s2key=val3",
				"s2lk =val4",
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
			wantLines: []string{
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
			wantLines:  []string{},
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
			if !reflect.DeepEqual(got, tt.wantLines) {
				t.Errorf("alignIni() got = %v, want %v", got, tt.wantLines)
			}
		})
	}
}
