package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"literal-reaction-mcp/model"
)

// WordHandler holds the database connection used by all MCP tool handlers.
type WordHandler struct {
	db *model.DBHandler
}

// NewWordHandler opens the database and returns a WordHandler.
func NewWordHandler(dbPath string) (*WordHandler, error) {
	db, err := model.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &WordHandler{db: db}, nil
}

// Close releases the underlying database connection.
func (h *WordHandler) Close() {
	h.db.Close()
}

// RegisterTools registers all LiteralReaction MCP tools onto the server.
func RegisterTools(s *server.MCPServer, h *WordHandler) {
	// Tool 1: 形容詞 + 名詞
	s.AddTool(
		mcp.NewTool(
			"generate_adj_noun_phrase",
			mcp.WithDescription(
				"ランダムな形容詞と名詞を組み合わせてパワーワードを生成します。"+
					"形容詞が名詞を直接修飾する形（例: 「美しい夜」「激しい嵐」）になります。",
			),
		),
		h.generateAdjNounPhrase,
	)

	// Tool 2: 副詞 + 動詞
	s.AddTool(
		mcp.NewTool(
			"generate_adv_verb_phrase",
			mcp.WithDescription(
				"ランダムな副詞と動詞を組み合わせてパワーワードを生成します。"+
					"副詞が動詞を修飾する形（例: 「素早く走る」「静かに眠る」）になります。",
			),
		),
		h.generateAdvVerbPhrase,
	)

	// Tool 3: 名詞 + 動詞（助詞はLLM選択）
	s.AddTool(
		mcp.NewTool(
			"generate_noun_verb_phrase",
			mcp.WithDescription(
				"ランダムな名詞と動詞を取得し、パワーワードを生成します。\n"+
					"名詞と動詞の間には日本語の格助詞（を・に・が・で・と）が必要です。\n"+
					"返された noun と verb の意味を考慮して、最も自然で意味が通る助詞を選び、\n"+
					"「{noun}[助詞]{verb}」の形式で完成したフレーズを返答してください。\n"+
					"例: noun=「猫」、verb=「食べる」→「猫を食べる」\n"+
					"例: noun=「山」、verb=「登る」→「山に登る」",
			),
		),
		h.generateNounVerbData,
	)
}

func (h *WordHandler) generateAdjNounPhrase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	adj, err := h.db.GetRandWord(model.Adj)
	if err != nil {
		return mcp.NewToolResultErrorf("形容詞の取得に失敗しました: %v", err), nil
	}

	noun, err := h.db.GetRandWord(model.Noun)
	if err != nil {
		return mcp.NewToolResultErrorf("名詞の取得に失敗しました: %v", err), nil
	}

	phrase := fmt.Sprintf("%s%s", adj.Word, noun.Word)
	return mcp.NewToolResultText(phrase), nil
}

func (h *WordHandler) generateAdvVerbPhrase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	adv, err := h.db.GetRandWord(model.Adverb)
	if err != nil {
		return mcp.NewToolResultErrorf("副詞の取得に失敗しました: %v", err), nil
	}

	verb, err := h.db.GetRandWord(model.Verb)
	if err != nil {
		return mcp.NewToolResultErrorf("動詞の取得に失敗しました: %v", err), nil
	}

	phrase := fmt.Sprintf("%s%s", adv.Word, verb.Word)
	return mcp.NewToolResultText(phrase), nil
}

func (h *WordHandler) generateNounVerbData(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	noun, err := h.db.GetRandWord(model.Noun)
	if err != nil {
		return mcp.NewToolResultErrorf("名詞の取得に失敗しました: %v", err), nil
	}

	verb, err := h.db.GetRandWord(model.Verb)
	if err != nil {
		return mcp.NewToolResultErrorf("動詞の取得に失敗しました: %v", err), nil
	}

	result := fmt.Sprintf("noun: %s\nverb: %s", noun.Word, verb.Word)
	return mcp.NewToolResultText(result), nil
}
