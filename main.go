package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./words.sqlite3"
	}

	handler, err := NewWordHandler(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer handler.Close()

	if err := handler.db.Validate(); err != nil {
		log.Fatalf("database validation failed: %v", err)
	}

	s := server.NewMCPServer(
		"literal-reaction-mcp",
		"1.0.0",
	)

	RegisterTools(s, handler)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	httpServer := server.NewStreamableHTTPServer(s)
	log.Printf("Starting LiteralReaction MCP server on %s", addr)
	log.Fatal(httpServer.Start(addr))
}
