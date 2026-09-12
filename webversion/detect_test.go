package webversion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseObjectParams(t *testing.T) {
	text := `const common={os_type:"web",version:"999.0.4",web_version:"3.0",x_app:"heybox_website"};`
	got, ok := Parse(text)
	if !ok {
		t.Fatal("Parse returned ok=false")
	}
	if got.Version != "999.0.4" || got.WebVersion != "3.0" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseQueryParams(t *testing.T) {
	text := `fetch("https://api.xiaoheihe.cn/bbs/app/feeds?os_type=web&version=999.0.4&web_version=3.0&client_type=web")`
	got, ok := Parse(text)
	if !ok {
		t.Fatal("Parse returned ok=false")
	}
	if got.Version != "999.0.4" || got.WebVersion != "3.0" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseRequiresNearbyPair(t *testing.T) {
	text := `library={version:"18.3.1"};` + string(make([]byte, 3000)) + `cfg={web_version:"3.0"}`
	if _, ok := Parse(text); ok {
		t.Fatal("Parse should reject unrelated version values")
	}
}

func TestDetectFromReferencedScript(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><script src="/assets/app.js"></script></html>`))
	})
	mux.HandleFunc("/assets/app.js", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`axios.defaults.params={version:"999.0.4",web_version:"3.0"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	got, err := Detect(context.Background(), client, []string{server.URL + "/"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "999.0.4" || got.WebVersion != "3.0" {
		t.Fatalf("got %+v", got)
	}
	if got.Source != server.URL+"/assets/app.js" {
		t.Fatalf("unexpected source %q", got.Source)
	}
}

func TestDetectFallsThroughRoots(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusBadGateway)
	}))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`window.client={version:"999.0.4",web_version:"3.0"}`))
	}))
	defer good.Close()

	got, err := Detect(context.Background(), &http.Client{Timeout: time.Second}, []string{bad.URL, good.URL})
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "999.0.4" || got.WebVersion != "3.0" {
		t.Fatalf("got %+v", got)
	}
}

func TestParsePrefersXHHVersionNearWebVersion(t *testing.T) {
	text := `bundle={version:"18.3.1",version:"999.0.4",web_version:"3.0"}`
	got, ok := Parse(text)
	if !ok {
		t.Fatal("Parse returned ok=false")
	}
	if got.Version != "999.0.4" {
		t.Fatalf("version=%q", got.Version)
	}
}
