package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"xhhrobot/config"
)

func TestSendReqUsesOpenAIChatCompletionsFormat(t *testing.T) {
	oldAI := config.ConfigStruct.Ai
	oldMgr := mcpMgr
	defer func() {
		config.ConfigStruct.Ai = oldAI
		mcpMgr = oldMgr
	}()
	mcpMgr = nil

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method=%s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization=%q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type=%q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body["model"] != "deepseek-flash" {
			t.Errorf("model=%v", body["model"])
		}
		messages, ok := body["messages"].([]any)
		if !ok || len(messages) != 2 {
			t.Errorf("messages=%#v", body["messages"])
		}
		if _, exists := body["input"]; exists {
			t.Error("Chat Completions request must not use Responses API input field")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"total_tokens":12}}`))
	}))
	defer server.Close()

	config.ConfigStruct.Ai.BaseUrl = server.URL
	config.ConfigStruct.Ai.Token = "test-key"
	config.ConfigStruct.Ai.WebSearch = false
	config.ConfigStruct.Ai.ForceWebSearch = false

	msgs := []any{
		SysMsg{Role: "system", Content: "system prompt"},
		Messages[string]{Role: "user", Content: "hello"},
	}
	resp := SendReq("deepseek-flash", msgs)
	if len(resp.Choices) != 1 {
		t.Fatalf("choices=%d", len(resp.Choices))
	}
	if resp.Choices[0].Msg.Content != "ok" {
		t.Fatalf("content=%q", resp.Choices[0].Msg.Content)
	}
	if resp.Usage.TotalToken != 12 {
		t.Fatalf("tokens=%d", resp.Usage.TotalToken)
	}
}
