# literal-reaction-mcp

**ランダムに単語を「反応」させて、パワーワードを作ろう(MCP版)**

## 概要

日本語のパワーワードをランダム生成する MCP サーバーです。MeCab IPADIC の語彙から完全ランダムに生成されます。

Webアプリとして動作する[LiteralReaction](https://github.com/Syuparn/LiteralReaction/tree/master)のMCPサーバー版です。

## 提供ツール

| ツール名 | 説明 |
|---|---|
| `generate_adj_noun_phrase` | ランダムな形容詞＋名詞のフレーズを生成 |
| `generate_adv_verb_phrase` | ランダムな副詞＋動詞のフレーズを生成 |
| `generate_noun_verb_phrase` | ランダムな名詞と動詞を返す（助詞の選択は LLM が担当）|

## プロンプト例

```
literal-reaction generate_noun_verb_phrase ランダムなフレーズを作ってください
```

```
literal-reaction generate_adj_noun_phrase ランダムなフレーズを100個作り、一番面白かったもの
```

## 生成結果例

- <span style="font-size: x-large;">**`「烏滸がましい練り製品」`**</span>

- <span style="font-size: x-large;">**`「他愛ないしょうぶ湯」`**</span>

- <span style="font-size: x-large;">**`「待ち遠しい木魚」`**</span>

- <span style="font-size: x-large;">**`「人でなしな売り手」`**</span>

## セットアップ

### サーバーを起動

```bash
docker run -p 8080:8080 ghcr.io/syuparn/literal-reaction-mcp:main
```

## MCPサーバーの登録

- Claude Codeの場合

```bash
claude mcp add --transport http literal-reaction http://localhost:8080/mcp
```

接続できたら準備完了

```bash
$ claude mcp list
Checking MCP server health...

literal-reaction: http://localhost:8080/mcp (HTTP) - ✓ Connected
```
