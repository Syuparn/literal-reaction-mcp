package model

import "testing"

func TestPosName(t *testing.T) {
	tests := []struct {
		pos  POS
		want string
	}{
		{Adj, "adjective"},
		{Adverb, "adverb"},
		{Noun, "noun"},
		{Verb, "verb"},
	}

	for _, tt := range tests {
		got := posName(tt.pos)
		if got != tt.want {
			t.Errorf("posName(%v) = %q, want %q", tt.pos, got, tt.want)
		}
	}
}

func TestPosTableName(t *testing.T) {
	tests := []struct {
		pos  POS
		want string
	}{
		{Adj, "adjectives"},
		{Adverb, "adverbs"},
		{Noun, "nouns"},
		{Verb, "verbs"},
	}

	for _, tt := range tests {
		got := posTableName(tt.pos)
		if got != tt.want {
			t.Errorf("posTableName(%v) = %q, want %q", tt.pos, got, tt.want)
		}
	}
}

func TestNamePOS(t *testing.T) {
	tests := []struct {
		name string
		want POS
	}{
		{"adjective", Adj},
		{"adverb", Adverb},
		{"noun", Noun},
		{"verb", Verb},
	}

	for _, tt := range tests {
		got := namePOS(tt.name)
		if got != tt.want {
			t.Errorf("namePOS(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestPosNameRoundTrip(t *testing.T) {
	for _, pos := range []POS{Adj, Adverb, Noun, Verb} {
		name := posName(pos)
		if name == "" {
			t.Errorf("posName(%v) returned empty string", pos)
			continue
		}
		got := namePOS(name)
		if got != pos {
			t.Errorf("namePOS(posName(%v)) = %v, want %v", pos, got, pos)
		}
	}
}
