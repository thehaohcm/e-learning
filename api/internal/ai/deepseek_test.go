package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const cardsJSON = `{"cards":[{"front":"leverage","payload":{"senses":[{"level":1,"meaning_vi":"tận dụng"}]}}]}`

func TestStreamedResponseIsJoined(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		half := len(cardsJSON) / 2
		for _, part := range []string{cardsJSON[:half], cardsJSON[half:]} {
			b := strings.ReplaceAll(part, `"`, `\"`)
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%s\"}}]}\n\n", b)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()
	c := NewClient("k", srv.URL, "m")
	got, err := c.GenerateCards(context.Background(), DeckContext{Kind: "vocab"}, 1, nil)
	if err != nil || len(got) != 1 || got[0].Front != "leverage" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestRetryWithoutJSONMode(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		if strings.Contains(string(body), "response_format") {
			http.Error(w, `{"error":"response_format not supported"}`, http.StatusBadRequest)
			return
		}
		content := strings.ReplaceAll("```json\n"+cardsJSON+"\n```", `"`, `\"`)
		content = strings.ReplaceAll(content, "\n", `\n`)
		fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":"%s"}}]}`, content)
	}))
	defer srv.Close()
	c := NewClient("k", srv.URL, "m")
	got, err := c.GenerateCards(context.Background(), DeckContext{Kind: "vocab"}, 1, nil)
	if err != nil || len(got) != 1 || calls != 2 {
		t.Fatalf("got %+v, %v, calls=%d", got, err, calls)
	}
}
