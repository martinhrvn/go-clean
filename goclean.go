package goclean

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type WordMatcher struct {
	Word    string `json:"word,omitempty"`
	Regex   string `json:"regex,omitempty"`
	Level   int32  `json:"level,omitempty,default=1"`
	Matcher *regexp.Regexp
}

type Config struct {
	DetectLeetSpeak   bool   `json:"detectLeetSpeak"`
	DetectObfuscated  bool   `json:"detectObfuscated"`
	Replacement       string `json:"replacement"`
	ObfuscationLength int32  `json:"obfuscationLength,default=3"`

	Profanities    []WordMatcher `json:"profanities"`
	FalsePositives []string      `json:"falsePositives"`
	FalseNegatives []WordMatcher `json:"falseNegatives"`
}

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

func putIndexesToMap(indexes []int, mapIndexes map[int]bool) {
	for i := indexes[0]; i < indexes[1]; i++ {
		mapIndexes[i] = true
	}
}

func (gc *GoClean) List(message string) []DetectedConcern {
	str := sanitizeString(message)
	detected := make([]DetectedConcern, 0)
	matched := make(map[int]bool)
	for _, negative := range gc.config.FalseNegatives {
		if negative.Matcher != nil {
			indexes := negative.Matcher.FindAllStringIndex(str, -1)
			for _, index := range indexes {
				putIndexesToMap(index, matched)
				detected = append(detected, DetectedConcern{
					Word:        negative.Word,
					MatchedText: str[index[0]:index[1]],
					StartIndex:  int32(index[0]),
					EndIndex:    int32(index[1]),
					Level:       negative.Level,
				})
			}
		}
	}
	for _, falsePositive := range gc.config.FalsePositives {
		if falsePositive != "" {
			indexes := regexp.MustCompile(falsePositive).FindAllStringIndex(str, -1)
			for _, index := range indexes {
				putIndexesToMap(index, matched)
			}
		}
	}
	for _, profanity := range gc.config.Profanities {
		if profanity.Matcher != nil {
			indexes := profanity.Matcher.FindAllStringIndex(str, -1)
			for _, index := range indexes {
				_, startFound := matched[index[0]]
				_, endFound := matched[index[1]]
				if !startFound && !endFound {
					detected = append(detected, DetectedConcern{
						Word:        profanity.Word,
						MatchedText: str[index[0]:index[1]],
						StartIndex:  int32(index[0]),
						EndIndex:    int32(index[1]),
						Level:       profanity.Level,
					})
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

var leetSpeakMapping = map[string]string{
	"a": "[a4]",
	"s": "[s5$]",
}

func NewProfanitySanitizer(c Config) GoClean {
	c.Profanities = initializeMatchers(c.Profanities, c)
	c.FalseNegatives = initializeMatchers(c.FalseNegatives, c)
	return GoClean{
		config: c,
	}
}

func initializeMatchers(matchers []WordMatcher, c Config) []WordMatcher {
	for i, profanity := range matchers {
		if profanity.Regex != "" {
			matchers[i].Matcher = regexp.MustCompile("(?i)" + profanity.Regex)
		} else if profanity.Word != "" {
			split := strings.Split(profanity.Word, "")
			if c.DetectLeetSpeak {
				for i, ch := range split {
					if leetSpeakMapping[ch] != "" {
						split[i] = leetSpeakMapping[ch]
					} else {
						split[i] = ch
					}
				}
			}
			if c.DetectObfuscated {
				str := strings.Join(split, fmt.Sprintf("\\W{0,%d}", c.ObfuscationLength))
				matchers[i].Matcher = regexp.MustCompile("(?i)" + str)
			} else {
				str := strings.Join(split, "")
				matchers[i].Matcher = regexp.MustCompile("(?i)" + str)
			}

		}
	}
	return matchers
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
