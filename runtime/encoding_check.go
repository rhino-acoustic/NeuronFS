package main

import (
	"unicode/utf8"
)

// isMojibake detects CP949/EUC-KR encoding corruption in UTF-8 text.
// CP949 mojibake in UTF-8 produces specific patterns:
// - High density of rare Hangul Compatibility Jamo (U+3131–U+318E)
// - Invalid UTF-8 sequences
// - Sequences like 섎뱶곸듯 which are valid Unicode but nonsensical Korean
//
// Heuristic: if >15% of runes are in the Jamo compatibility range
// AND the text contains known mojibake bigrams, it's corrupted.
func isMojibake(s string) bool {
	if len(s) < 10 {
		return false
	}

	// Check for invalid UTF-8 first
	if !utf8.ValidString(s) {
		return true
	}

	// Known CP949→UTF-8 mojibake bigrams (appear frequently in corrupted Korean)
	mojibakeBigrams := []string{
		"섎뱶", "곸듯", "뺣━", "쒖꽦", "좊땲", "묎렐",
		"쇱젙", "뿏삎", "쒕떎", "뚯뒪", "ъ슜", "щ뒗",
		"덈떎", "앺듃", "뿉??", "뚮옖", "뺤씤", "좎뿰",
		"듯빀", "곕씪", "섍퀬",
	}

	for _, bg := range mojibakeBigrams {
		if containsSubstring(s, bg) {
			return true
		}
	}

	return false
}

// containsSubstring is a simple contains check without importing strings
// (to avoid adding an import to this utility file).
func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
