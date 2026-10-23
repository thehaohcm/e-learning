package study

import (
	"encoding/json"
	"testing"
)

func TestWithExample(t *testing.T) {
	raw := []byte(`{"senses":[{"level":2,"meaning_vi":"b","example_en":"old2"},{"level":1,"meaning_vi":"a","example_en":"old1","example_vi":"cũ"}],"ipa":"/x/"}`)
	out, ok := withExample(raw, 42, "The hypothesis was tested.", "Giả thuyết đã được kiểm chứng.")
	if !ok {
		t.Fatal("not applied")
	}
	var p struct {
		IPA    string `json:"ipa"`
		Senses []struct {
			Level     int    `json:"level"`
			ExampleEn string `json:"example_en"`
			ExampleVi string `json:"example_vi"`
			ExampleID int64  `json:"example_id"`
		} `json:"senses"`
	}
	json.Unmarshal(out, &p)
	if p.IPA != "/x/" || p.Senses[0].ExampleEn != "old2" || p.Senses[1].ExampleEn != "The hypothesis was tested." ||
		p.Senses[1].ExampleVi != "Giả thuyết đã được kiểm chứng." || p.Senses[1].ExampleID != 42 {
		t.Fatalf("%+v", p)
	}
	if _, ok := withExample([]byte(`{"back":"x"}`), 1, "a", "b"); ok {
		t.Fatal("card without senses must be left alone")
	}
}
