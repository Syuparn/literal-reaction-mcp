package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	_ "github.com/mattn/go-sqlite3"

	"literal-reaction-mcp/model"
)

// setupTestHandler builds a WordHandler backed by an in-memory SQLite database.
func setupTestHandler(t *testing.T) *WordHandler {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	schema := `
		CREATE TABLE adjectives (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE adverbs   (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE nouns     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE verbs     (id INTEGER PRIMARY KEY, word TEXT);
		CREATE TABLE counts    (table_name VARCHAR(255), row_count INT);
		CREATE TABLE favorite_sentences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			former_pos TEXT, latter_pos TEXT, particle TEXT,
			former_word TEXT, latter_word TEXT
		);
		INSERT INTO adjectives VALUES (1, '美しい'), (2, '激しい');
		INSERT INTO adverbs   VALUES (1, '素早く'),  (2, '静かに');
		INSERT INTO nouns     VALUES (1, '猫'),      (2, '山');
		INSERT INTO verbs     VALUES (1, '走る'),    (2, '眠る');
		INSERT INTO counts VALUES
			('adjectives', 2), ('adverbs', 2), ('nouns', 2), ('verbs', 2);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	dbHandler, err := model.OpenDBFromConn(db)
	if err != nil {
		t.Fatalf("failed to create DBHandler: %v", err)
	}

	t.Cleanup(func() { dbHandler.Close() })
	return &WordHandler{db: dbHandler}
}

func TestGenerateAdjNounPhrase(t *testing.T) {
	h := setupTestHandler(t)

	result, err := h.generateAdjNounPhrase(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("generateAdjNounPhrase() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("generateAdjNounPhrase() returned nil result")
	}
	if result.IsError {
		t.Fatal("generateAdjNounPhrase() returned tool error")
	}

	text := extractText(result)
	if text == "" {
		t.Error("generateAdjNounPhrase() returned empty phrase")
	}
}

func TestGenerateAdvVerbPhrase(t *testing.T) {
	h := setupTestHandler(t)

	result, err := h.generateAdvVerbPhrase(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("generateAdvVerbPhrase() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("generateAdvVerbPhrase() returned nil result")
	}
	if result.IsError {
		t.Fatal("generateAdvVerbPhrase() returned tool error")
	}

	text := extractText(result)
	if text == "" {
		t.Error("generateAdvVerbPhrase() returned empty phrase")
	}
}

func TestGenerateNounVerbData(t *testing.T) {
	h := setupTestHandler(t)

	result, err := h.generateNounVerbData(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("generateNounVerbData() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("generateNounVerbData() returned nil result")
	}
	if result.IsError {
		t.Fatal("generateNounVerbData() returned tool error")
	}

	// Result must contain both "noun:" and "verb:" labels.
	text := extractText(result)
	if !strings.Contains(text, "noun:") {
		t.Errorf("generateNounVerbData() result missing \"noun:\": %q", text)
	}
	if !strings.Contains(text, "verb:") {
		t.Errorf("generateNounVerbData() result missing \"verb:\": %q", text)
	}
}

// extractText pulls the first text content from a CallToolResult.
func extractText(r *mcp.CallToolResult) string {
	for _, c := range r.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
