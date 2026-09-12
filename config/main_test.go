package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	ConfigStruct.Ai.APIKey = ""
	ConfigStruct.Ai.Token = ""
	applyDefaults()
	if !ConfigStruct.Xhh.AutoVersion {
		t.Fatal("autoVersion should default to true")
	}
	if ConfigStruct.Xhh.Ver != "999.0.4" || ConfigStruct.Xhh.WebVer != "3.0" {
		t.Fatalf("unexpected XHH defaults: %s / %s", ConfigStruct.Xhh.Ver, ConfigStruct.Xhh.WebVer)
	}
	if ConfigStruct.Ai.Model != "deepseek-flash" {
		t.Fatalf("model=%q", ConfigStruct.Ai.Model)
	}
	if ConfigStruct.Ai.BaseUrl != "https://api.deepseek.com/chat/completions" {
		t.Fatalf("baseUrl=%q", ConfigStruct.Ai.BaseUrl)
	}
	if ConfigStruct.Ai.MCP.Enabled {
		t.Fatal("MCP should be disabled by default")
	}
}

func TestExistingConfigWithoutAutoVersionKeepsAutoDefault(t *testing.T) {
	applyDefaults()
	if err := json.Unmarshal([]byte(`{"xhh":{"webver":"2.5","version":"999.0.4"}}`), &ConfigStruct); err != nil {
		t.Fatal(err)
	}
	if !ConfigStruct.Xhh.AutoVersion {
		t.Fatal("old config without autoVersion should inherit true")
	}
}

func TestAPIKeyCompatibility(t *testing.T) {
	applyDefaults()
	ConfigStruct.Ai.APIKey = "new-key"
	ConfigStruct.Ai.Token = ""
	normalizeConfig()
	if AIKey() != "new-key" || ConfigStruct.Ai.Token != "new-key" {
		t.Fatal("apiKey should populate legacy token field at runtime")
	}

	ConfigStruct.Ai.APIKey = ""
	ConfigStruct.Ai.Token = "legacy-key"
	normalizeConfig()
	if AIKey() != "legacy-key" || ConfigStruct.Ai.APIKey != "legacy-key" {
		t.Fatal("legacy token should remain compatible")
	}
}

func TestGeneratedConfigUsesAPIKey(t *testing.T) {
	ConfigStruct.Ai.APIKey = ""
	ConfigStruct.Ai.Token = ""
	applyDefaults()
	data, err := json.Marshal(ConfigStruct)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `"apiKey":""`) {
		t.Fatalf("generated config missing apiKey: %s", text)
	}
	if strings.Contains(text, `"token":`) {
		t.Fatalf("generated config should omit legacy token: %s", text)
	}
}
