package a2a

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchAgentCard(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/agent.json" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name":            "Currency Agent",
			"description":     "Helps with exchange rates",
			"url":             "http://example.test/",
			"version":         "1.0.0",
			"protocolVersion": "0.3.0",
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	c.SetHTTPClient(srv.Client())

	card, err := c.FetchAgentCard(context.Background())
	if err != nil {
		t.Fatalf("FetchAgentCard: %v", err)
	}
	if card.Name != "Currency Agent" {
		t.Fatalf("name=%q", card.Name)
	}
	if c.rpcURL != "http://example.test" {
		t.Fatalf("rpcURL=%q", c.rpcURL)
	}
}

func TestSendText_TaskArtifacts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/.well-known/agent.json":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "Test", "url": r.Host, "protocolVersion": "0.3.0", "version": "1",
			})
		case r.Method == http.MethodPost:
			var req jsonRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode: %v", err)
				http.Error(w, "bad", 400)
				return
			}
			if req.Method != "message/send" {
				t.Errorf("method=%q", req.Method)
			}
			result := map[string]interface{}{
				"kind": "task",
				"id":   "task-1",
				"status": map[string]interface{}{
					"state": "completed",
				},
				"artifacts": []map[string]interface{}{
					{
						"artifactId": "a1",
						"name":       "conversion_result",
						"parts": []map[string]interface{}{
							{"kind": "text", "text": "10 USD is about 9.2 EUR"},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  result,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	c.SetHTTPClient(srv.Client())
	c.rpcURL = srv.URL

	reply, err := c.SendText(context.Background(), "How much is 10 USD to EUR?")
	if err != nil {
		t.Fatalf("SendText: %v", err)
	}
	if !strings.Contains(reply, "9.2 EUR") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestSendText_MessageResult(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "1",
			"result": map[string]interface{}{
				"kind":      "message",
				"messageId": "m1",
				"role":      "agent",
				"parts": []map[string]interface{}{
					{"kind": "text", "text": "hello back"},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	c.SetHTTPClient(srv.Client())

	reply, err := c.SendText(context.Background(), "hi")
	if err != nil {
		t.Fatalf("SendText: %v", err)
	}
	if reply != "hello back" {
		t.Fatalf("reply=%q", reply)
	}
}

func TestSendText_Empty(t *testing.T) {
	t.Parallel()
	c := NewClient("http://127.0.0.1:9")
	_, err := c.SendText(context.Background(), "  ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendText_RPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "1",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
		})
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	c.SetHTTPClient(srv.Client())
	_, err := c.SendText(context.Background(), "hi")
	if err == nil || !strings.Contains(err.Error(), "Invalid Request") {
		t.Fatalf("err=%v", err)
	}
}

func TestExtractFinalText_HistoryFallback(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"kind":"task","id":"t1",
		"status":{"state":"completed"},
		"history":[
			{"role":"user","parts":[{"kind":"text","text":"q"}]},
			{"role":"agent","parts":[{"kind":"text","text":"from history"}]}
		]
	}`)
	got, err := extractFinalText(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != "from history" {
		t.Fatalf("got=%q", got)
	}
}
