![go-clean](/.github/assets/go-clean.png)

# go-clean

[![Maintainability](https://api.codeclimate.com/v1/badges/eeb2df6c6af1ebf7a9f0/maintainability)](https://codeclimate.com/github/martinhrvn/go-clean/maintainability)
[![codecov](https://codecov.io/gh/martinhrvn/go-clean/branch/main/graph/badge.svg?token=UHYSK2ZZ67)](https://codecov.io/gh/martinhrvn/go-clean)
[![Go Report Card](https://goreportcard.com/badge/github.com/martinhrvn/go-clean)](https://goreportcard.com/report/github.com/martinhrvn/go-clean)
[![Go Reference](https://pkg.go.dev/badge/github.com/martinhrvn/go-clean.svg)](https://pkg.go.dev/github.com/martinhrvn/go-clean)
[![Follow martinhrvn](https://img.shields.io/github/followers/martinhrvn?label=Follow&style=social)](https://github.com/martinhrvn)

`go-clean` is a flexible, stand-alone, lightweight library for detecting and censoring profanities in Go.

# Installation
```console
go get -u github.com/martinhrvn/go-clean
```
## Usage
By default 
```go
package main

import (
    goclean "github.com/martinhrvn/go-clean"
)

func main() {
    goclean.IsProfane("fuck this shit")
    // returns true  
    goclean.List("fuck this shit")         
    // returns "DetectedConcern{Word: "fuck", MatchedWord: "fuck", StartIndex: 0, EndIndex: 3}"
    goclean.Redact("fuck this shit")
    // returns "**** this shit"
}
```

Calling `goclean.IsProfane(s)`, `goclean.ExtractProfanity(s)` or `goclean.Redact(s)` will use the default profanity detector, 
that is configured in the `config.json` file.

If you'd like to disable leet speak, numerical character or special character sanitization, you have to create a
ProfanityDetector instead:
```go
profanityDetector := goclean.NewProfanitySanitizer(goclean.Config{
    // will not sanitize leet speak (a$$, b1tch, etc.)
    DetectLeetSpeak: false,
    // will not detect obfuscated words (f_u_c_k, etc.)
    DetectObfuscated: false,
    // replacement character for redacted words
    ReplacementCharacter: '*', 
    // Lenght for obfuscated characters (e.g. if set to "1" f_u_c_k will be detected but f___u___c___k won't)
    ObfuscationLength: 1,
	
    Profanities: []goclean.WordMatcher{
        { Word: "fuck", Regex: "f[u]+ck" }
    }
})
```

## Configuration

### Base configuration
- `DetectLeetSpeak`: sanitize leet speak (`a$$`, `b1tch`, etc.)
  - default: `true`
- `DetectObfuscated`: detect obfuscated words (`f_u_c_k`, etc.)
  - default: `true`
- `ObfuscationLength`: length for obfuscated characters (e.g. if set to "1" `f_u_c_k` will be detected but `f___u___c___k` won't)
  - default: `3`
- `ReplacementCharacter`: replacement character for redacted words
  - default: `*`

### WordMatchers
used for profanities and false negatives configuration

- `Regex`:
  - if found it will be used to match word instead of `Word`
- `Word`: 
    - word to detect, 
    - if `DetectObfuscated: true` it will also match words with `ObfuscationLength` characters in between letters
- `Level`:
  - optional profanity level that will be returned from `List` method

### False positive
These are words that contain words that are profanities but are not profane themselves.
For example word `bass` contains `ass` but is not profane.

### False negatives
These are words that may be incorrectly filtered as false positives and words that should always be treated as profane, regardless of false postives. 
These are matched before false positives are removed.

For example: `dumbass` is false negative, as `bass` is false positive so to be matched it needs to be added to false negatives.

## Methods

### List

Returns list of `DetectedConcerns` for profanities found in the given string.
This contains:
- `Word`: base word found (in case only regex is provided empty string will be returned, e.g. for `fuuuck` it will be `fuck`)
- `MatchedWord`: actual word found in string (e.g. for `fuuuck` it will be `fuuuck`)
- `StartIndex`: start index of word in string
- `EndIndex`: end index of word in string
- `Level`: profanity level (if provided, else it will be `0`)

If the configuration is:
```go
WordMatcher {
    Word: "fuck"
    Regex: "f[u]+ck"
    Level: 1
}
```
and the input string is `fuuuck`, it will return:
```go
DetectedEntity {
    Word: "fuck"
    MatchedWord: "fuuuck"
    StartIndex: 0
    EndIndex: 6
}
```
### Redact
It will return string with profanities replaced with `ReplacementCharacter` for each character of detected profanities.

The input string `"shit hit the fan"` will be returned as `"**** hit the fan"`.

### IsProfane
Returns `true` if the given string contains profanities.

The input string `"shit hit the fan"` returns `true`.


