package generator

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type nameParts struct {
	first  string
	middle []string
	last   string
}

func parseName(fullName string) nameParts {
	fullName = normalize(fullName)
	words := strings.Fields(fullName)
	if len(words) == 0 {
		return nameParts{}
	}

	np := nameParts{
		first: strings.ToLower(words[0]),
	}
	if len(words) == 1 {
		return np
	}

	np.last = strings.ToLower(words[len(words)-1])
	if len(words) > 2 {
		for _, w := range words[1 : len(words)-1] {
			np.middle = append(np.middle, strings.ToLower(w))
		}
	}
	return np
}

func normalize(s string) string {
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)))
	result, _, _ := transform.String(t, s)
	return result
}

func (np nameParts) initials() (firstInit, lastInit string) {
	if len(np.first) > 0 {
		firstInit = string(np.first[0])
	}
	if len(np.last) > 0 {
		lastInit = string(np.last[0])
	}
	return
}

func (np nameParts) middleInitials() string {
	var initials []string
	for _, m := range np.middle {
		if len(m) > 0 {
			initials = append(initials, string(m[0]))
		}
	}
	return strings.Join(initials, "")
}

func (np nameParts) middleJoined(sep string) string {
	return strings.Join(np.middle, sep)
}
