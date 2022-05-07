package goclean

import (
	"regexp"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type GoClean struct {
	config Config
}

type DetectedConcern struct {
	Word        string
	MatchedText string
	StartIndex  int32
	EndIndex    int32
	Level       int32
}

func (gc *GoClean) List(message string) []DetectedConcern {
	str := sanitizeString(message)
	detected := make([]DetectedConcern, 0)
	matched := make(map[int]bool)
	detected = append(detected, gc.detectConcerns(str, gc.config.FalseNegatives, matched)...)
	for _, falsePositive := range gc.config.FalsePositives {
		if falsePositive != "" {
			indexes := regexp.MustCompile(falsePositive).FindAllStringIndex(str, -1)
			for _, index := range indexes {
				putIndexesToMap(index, matched)
			}
		}
	}
	detected = append(detected, gc.detectConcerns(str, gc.config.Profanities, matched)...)
	return detected
}

func (gc GoClean) detectConcerns(message string, matchers []WordMatcher, matched map[int]bool) []DetectedConcern {
	detected := make([]DetectedConcern, 0)
	for _, profanity := range matchers {
		if profanity.Matcher != nil {
			indexes := profanity.Matcher.FindAllStringIndex(message, -1)
			for _, index := range indexes {
				start := index[0]
				end := index[1]
				if !isAlreadyMatched(start, end, matched) {
					detected = append(detected, DetectedConcern{
						Word:        profanity.Word,
						MatchedText: message[start:end],
						StartIndex:  int32(start),
						EndIndex:    int32(end),
						Level:       profanity.Level,
					})
					putIndexesToMap(index, matched)
				}
			}
		}
	}
	return detected
}

func (gc *GoClean) Redact(str string) string {
	redacted := sanitizeString(str)
	for _, profanity := range gc.config.Profanities {
		if profanity.Matcher != nil {
			redacted = profanity.Matcher.ReplaceAllStringFunc(redacted, func(s string) string {
				str := ""
				for i := 0; i < len(s); i++ {
					str += gc.config.Replacement
				}
				return str
			})
		}
	}
	return redacted
}

func (gc *GoClean) IsProfane() bool {
	return false
}

func NewProfanitySanitizer(c Config) GoClean {
	c.Profanities = c.initializeMatchers(c.Profanities)
	c.FalseNegatives = c.initializeMatchers(c.FalseNegatives)
	return GoClean{
		config: c,
	}
}

func sanitizeString(message string) string {
	bytes := make([]byte, len(message))
	normalize := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	_, _, err := normalize.Transform(bytes, []byte(message), true)
	if err != nil {
		return message
	}
	message = string(bytes)
	return message
}

func putIndexesToMap(indexes []int, mapIndexes map[int]bool) {
	for i := indexes[0]; i < indexes[1]; i++ {
		mapIndexes[i] = true
	}
}

func isAlreadyMatched(start, end int, matched map[int]bool) bool {
	_, startFound := matched[start]
	_, endFound := matched[end]
	return startFound || endFound
}
