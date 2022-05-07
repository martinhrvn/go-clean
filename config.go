package goclean

import (
	"fmt"
	"regexp"
	"strings"
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

var leetSpeakMapping = map[string]string{
	"a": "[a4]",
	"s": "[s5$]",
}

func (c *Config) initializeMatchers(matchers []WordMatcher) []WordMatcher {
	for i, m := range matchers {
		if m.Regex != "" {
			matchers[i].Matcher = regexp.MustCompile("(?i)" + m.Regex)
		} else if m.Word != "" {
			split := strings.Split(m.Word, "")
			c.replaceLeetSpeak(split)
			c.configureDetectObfuscated(matchers, split, i)
		}
	}
	return matchers
}

func (c *Config) configureDetectObfuscated(matchers []WordMatcher, split []string, i int) {
	if c.DetectObfuscated {
		str := strings.Join(split, fmt.Sprintf("\\W{0,%d}", c.ObfuscationLength))
		matchers[i].Matcher = regexp.MustCompile("(?i)" + str)
	} else {
		str := strings.Join(split, "")
		matchers[i].Matcher = regexp.MustCompile("(?i)" + str)
	}
}

func (c *Config) replaceLeetSpeak(chars []string) {
	if c.DetectLeetSpeak {
		for i, ch := range chars {
			if leetSpeakMapping[ch] != "" {
				chars[i] = leetSpeakMapping[ch]
			} else {
				chars[i] = ch
			}
		}
	}
}
