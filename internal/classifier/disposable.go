package classifier

import (
	"bufio"
	"bytes"
	_ "embed"
	"strings"
)

//go:embed disposable_domains.txt
var disposableDomainsData []byte

var disposableSet map[string]bool

func init() {
	disposableSet = make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(disposableDomainsData))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		disposableSet[strings.ToLower(line)] = true
	}
}
