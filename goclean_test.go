package goclean

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

var (
	file, _ = ioutil.ReadFile("config.json")
	config  = Config{}
	_       = json.Unmarshal(file, &config)
)

func TestGoClean_IsProfane(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"no profanity", "hello world", false},
		{"profanity", "hello world fuck", true},
		{"should match exact words", "ass", true},
		{"regex", "fuuuuck", true},
		{"should match obfuscated words", "a.s.s", true},
		{"should match obfuscated words", "a  s  s", true},
		{"should not match obfuscated words with length > set value", "a....s....s", false},
		{"should match leet speak", "4$$", true},
		{"should match leet speak and obfuscation", "a.$.$", true},
		{"should match false negatives", "dumbass", true},
		{"should match false positive", "bass", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gc := NewProfanitySanitizer(config)
			got := gc.IsProfane(test.text)
			if got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}

func TestGoClean_Redact(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"no profanity", "hello world", "hello world"},
		{"profanity", "hello world fuck", "hello world ****"},
		{"should match exact words", "ass", "***"},
		{"regex", "fuuuuck", "*******"},
		{"should match obfuscated words", "a.s.s", "*****"},
		{"should match obfuscated words", "a  s  s", "*******"},
		{"should not match obfuscated words with length > set value", "a....s....s", "a....s....s"},
		{"should match leet speak", "4$$", "***"},
		{"should match leet speak and obfuscation", "a.$.$", "*****"},
		{"should match false negatives", "dumbass", "*******"},
		{"should match false positive", "bass", "bass"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gc := NewProfanitySanitizer(config)
			got := gc.Redact(test.text)
			if got != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}

func TestGoClean_List(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []DetectedConcern
	}{
		{"no profanity", "hello world", []DetectedConcern{}},
		{"profanity", "hello world fuck", []DetectedConcern{{MatchedText: "fuck", StartIndex: 12, EndIndex: 16}}},
		{"should match exact words", "ass", []DetectedConcern{{Word: "ass", MatchedText: "ass", StartIndex: 0, EndIndex: 3, Level: 2}}},
		{"regex", "fuuuuck", []DetectedConcern{{MatchedText: "fuuuuck", StartIndex: 0, EndIndex: 7}}},
		{"regex with word", "daaaamn", []DetectedConcern{{Word: "damn", MatchedText: "daaaamn", StartIndex: 0, EndIndex: 7, Level: 2}}},
		{"should match obfuscated words", "a.s.s", []DetectedConcern{{Word: "ass", MatchedText: "a.s.s", StartIndex: 0, EndIndex: 5, Level: 2}}},
		{"should match obfuscated words", "a  s  s", []DetectedConcern{{Word: "ass", MatchedText: "a  s  s", StartIndex: 0, EndIndex: 7, Level: 2}}},
		{"should not match obfuscated words with length > set value", "a....s....s", []DetectedConcern{}},
		{"should match leet speak", "4$$", []DetectedConcern{{Word: "ass", MatchedText: "4$$", StartIndex: 0, EndIndex: 3, Level: 2}}},
		{"should match leet speak and obfuscation", "a.$.$", []DetectedConcern{{Word: "ass", MatchedText: "a.$.$", StartIndex: 0, EndIndex: 5, Level: 2}}},
		{"should match false negatives", "dumbass", []DetectedConcern{{Word: "dumbass", MatchedText: "dumbass", StartIndex: 0, EndIndex: 7, Level: 2}}},
		{"should match false positive", "bass", []DetectedConcern{}},
		{"should match case insensitive", "ASS", []DetectedConcern{{Word: "ass", MatchedText: "ASS", StartIndex: 0, EndIndex: 3, Level: 2}}},
		{"should handle multi-byte characters case insensitive", "世界 世界 ASS 世界", []DetectedConcern{{Word: "ass", MatchedText: "ASS", StartIndex: 14, EndIndex: 17, Level: 2}}},
		{"should sanitize special characters", "fûçk", []DetectedConcern{{MatchedText: "fuck", StartIndex: 0, EndIndex: 4}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gc := NewProfanitySanitizer(config)
			got := gc.List(test.text)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
