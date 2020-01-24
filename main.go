package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type CommitMessage struct {
	Header  Header
	Body    string
	Footers []Footer
}

type Header struct {
	Type     string
	Scope    string
	Breaking bool
	Message  string
}

type Footer struct {
	Token     string
	Value     string
	Separator string
}

func readLines(path string) ([]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return splitToLines(string(bytes)), nil
}

func splitToLines(str string) []string {
	var lines []string
	for _, line := range strings.Split(str, "\n") {
		lines = append(lines, strings.TrimSuffix(line, "\r"))
	}
	return lines
}

func findGroups(r *regexp.Regexp, str string) (map[string]string, error) {
	m := r.FindStringSubmatch(str)
	if m == nil {
		return nil, errors.New("does not match")
	}
	byName := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && name != "" {
			byName[name] = m[i]
		}
	}

	return byName, nil
}

func ParseHeader(line string) (Header, error) {
	pattern := regexp.MustCompile(`^(?P<type>\w+)(?:\((?P<scope>[\w$.\-*/ ]*)\))?(?P<breaking>!)?: (?P<msg>.*)$`)
	byName, err := findGroups(pattern, line)
	if err != nil {
		return Header{}, err
	}
	return Header{
		Type:     byName["type"],
		Scope:    byName["scope"],
		Breaking: byName["breaking"] == "!",
		Message:  byName["msg"],
	}, nil
}

func ParseFooter(line string) (Footer, error) {
	pattern := regexp.MustCompile(`^(?P<token>BREAKING CHANGE|[\w\-]+)(?P<separator>(: )|( #))(?P<value>.+)$`)
	byName, err := findGroups(pattern, line)
	if err != nil {
		return Footer{}, err
	}
	return Footer{
		Token:     byName["token"],
		Separator: byName["separator"],
		Value:     byName["value"],
	}, nil
}

func ParseLines(lines []string) (CommitMessage, error) {
	lineCount := len(lines)
	if lineCount == 0 {
		return CommitMessage{}, errors.New("empty message")
	}

	header, err := ParseHeader(lines[0])
	if err != nil {
		return CommitMessage{}, err
	}

	if lineCount > 1 && lines[1] != "" {
		return CommitMessage{}, errors.New("no empty line after header")
	}

	var bodyLines []string

	footerStarted := false
	var footer Footer
	var footerValueLines []string
	footers := []Footer{}

	lineCount -= 2 // header and empty line

	for i, line := range lines[2:] {
		parsed, err := ParseFooter(line)
		if err == nil { // footer found
			if footerStarted { // this is not the first footer
				footers = addFooter(footers, footer, footerValueLines)
			} else { // this is first footer
				footerStarted = true
			}
			footer = parsed
			footerValueLines = []string{parsed.Value}
		} else { // this is not a footer
			if footerStarted { // in the middle of footer
				footerValueLines = append(footerValueLines, line)
			} else { // this is body
				bodyLines = append(bodyLines, line)
			}
		}

		if footerStarted && i == lineCount-1 { // last line
			footers = addFooter(footers, footer, footerValueLines)
		}
	}

	return CommitMessage{
		Header:  header,
		Body:    removeTrailingNewline(joinLines(bodyLines)),
		Footers: footers,
	}, nil
}

func addFooter(footers []Footer, footer Footer, lines []string) []Footer {
	footer.Value = removeTrailingNewline(joinLines(lines))
	return append(footers, footer)
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func removeTrailingNewline(str string) string {
	return strings.TrimRight(str, "\r\n")
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("only one argument is expected")
	}

	lines, err := readLines(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := ParseLines(lines)
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, err := json.Marshal(parsed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonBytes))
}
