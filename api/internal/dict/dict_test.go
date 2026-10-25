package dict

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sample = `{"exists":true,"word":"reluctant","results":[{"lang_code":"en","lang_name":"Tiếng Anh",
 "meanings":[
  {"definition":"Miễn cưỡng, bất đắc dĩ.","definition_lang":"vi","pos":"Tính từ","source":"x"},
  {"definition":"miễn cưỡng, bất đắc dĩ","definition_lang":"vi","pos":"Tính từ"},
  {"definition":"unwilling","definition_lang":"en","pos":"adjective"},
  {"definition":"Chống lại","definition_lang":"vi","pos":"Danh từ"}],
 "pronunciations":[{"ipa":"rɪˈlʌktənt","region":"UK"}],
 "translations":[{"lang_code":"vi","translation":"ngần ngại"},{"lang_code":"fr","translation":"réticent"}]},
 {"lang_code":"vi","meanings":[{"definition":"không liên quan","definition_lang":"vi","pos":"Danh từ"}]}]}`

func TestParse(t *testing.T) {
	e, err := Parse("reluctant", []byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if !e.Found || e.IPA != "/rɪˈlʌktənt/" {
		t.Fatalf("got %+v", e)
	}
	want := []Meaning{{"adjective", "Miễn cưỡng, bất đắc dĩ"}, {"noun", "Chống lại"}, {"", "ngần ngại"}}
	if len(e.Meanings) != len(want) {
		t.Fatalf("meanings %+v", e.Meanings)
	}
	for i := range want {
		if e.Meanings[i] != want[i] {
			t.Fatalf("meaning %d = %+v, want %+v", i, e.Meanings[i], want[i])
		}
	}
	if e.Short(2) != "Miễn cưỡng, bất đắc dĩ; Chống lại" || e.FirstPOS() != "adjective" {
		t.Fatalf("short %q", e.Short(2))
	}
}

func TestParseMissing(t *testing.T) {
	e, err := Parse("qwzx", []byte(`{"exists":false,"word":"qwzx","results":[]}`))
	if err != nil || e.Found {
		t.Fatalf("%+v %v", e, err)
	}
}

func TestLookupCaches(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/v1/lookup" || r.URL.Query().Get("word") != "reluctant" || r.URL.Query().Get("def_lang") != "vi" {
			t.Errorf("bad request %s", r.URL)
		}
		w.Write([]byte(sample))
	}))
	defer srv.Close()
	c := New(srv.URL)
	for i := 0; i < 3; i++ {
		e, err := c.Lookup(context.Background(), "  Reluctant ")
		if err != nil || !e.Found {
			t.Fatalf("%+v %v", e, err)
		}
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}
