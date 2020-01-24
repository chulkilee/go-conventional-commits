package main

import (
	"reflect"
	"testing"
)

func TestParseHeader(t *testing.T) {
	sets := []struct {
		input string
		want  Header
	}{
		{"feat: hello world", Header{
			Type:     "feat",
			Scope:    "",
			Breaking: false,
			Message:  "hello world",
		}},
		{"feat(api): hello world", Header{
			Type:     "feat",
			Scope:    "api",
			Breaking: false,
			Message:  "hello world",
		}},
		{"feat!: drop A", Header{
			Type:     "feat",
			Scope:    "",
			Breaking: true,
			Message:  "drop A",
		}},
		{"feat(api)!: hello world", Header{
			Type:     "feat",
			Scope:    "api",
			Breaking: true,
			Message:  "hello world",
		}},
	}
	for _, set := range sets {
		got, _ := ParseHeader(set.input)
		if got != set.want {
			t.Errorf("Does not match:\ngot:  %+v\nwant: %+v", got, set.want)
		}
	}
}

func TestParseFooter(t *testing.T) {
	sets := []struct {
		input string
		want  Footer
	}{
		{"Reviewed-by: Z", Footer{
			Token:     "Reviewed-by",
			Separator: ": ",
			Value:     "Z",
		}},
		{"Refs #133", Footer{
			Token:     "Refs",
			Separator: " #",
			Value:     "133",
		}},
		{"Refs: #133", Footer{
			Token:     "Refs",
			Separator: ": ",
			Value:     "#133",
		}},
	}
	for _, set := range sets {
		got, _ := ParseFooter(set.input)
		if got != set.want {
			t.Errorf("Does not match:\ngot:  %+v\nwant: %+v", got, set.want)
		}
	}
}

func TestParseLines(t *testing.T) {
	sets := []struct {
		input string
		want  CommitMessage
	}{
		{`fix(typo): correct minor typos in code

see the issue for details

    /bin/sh test

on typos fixed.

Reviewed-by: Z
Refs #133`, CommitMessage{
			Header: Header{
				Type:     "fix",
				Scope:    "typo",
				Breaking: false,
				Message:  "correct minor typos in code",
			},
			Body: `see the issue for details

    /bin/sh test

on typos fixed.`,
			Footers: []Footer{
				Footer{
					Token:     "Reviewed-by",
					Separator: ": ",
					Value:     "Z",
				},
				Footer{
					Token:     "Refs",
					Separator: " #",
					Value:     "133",
				},
			},
		},
		},
		{`fix(typo): correct minor typos in code

Reviewed-by: Z
Refs #133

following
`, CommitMessage{
			Header: Header{
				Type:     "fix",
				Scope:    "typo",
				Breaking: false,
				Message:  "correct minor typos in code",
			},
			Body: "",
			Footers: []Footer{
				Footer{
					Token:     "Reviewed-by",
					Separator: ": ",
					Value:     "Z",
				},
				Footer{
					Token:     "Refs",
					Separator: " #",
					Value:     "133\n\nfollowing",
				},
			},
		},
		},
	}
	for _, set := range sets {
		lines := splitToLines(set.input)
		got, _ := ParseLines(lines)
		if !reflect.DeepEqual(got, set.want) {
			t.Errorf("Does not match:\ngot:  %+v\nwant: %+v", got, set.want)
		}
	}
}
