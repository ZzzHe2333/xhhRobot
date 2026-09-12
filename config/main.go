package config

import (
	"encoding/json"
	"os"

	"xhhrobot/loger"
)

const (
	DefaultXHHVersion    = "999.0.4"
	DefaultXHHWebVersion = "3.0"
	DefaultAIModel       = "deepseek-flash"
	DefaultAIBaseURL     = "https://api.deepseek.com/chat/completions"
	DefaultAIPrompt      = "你是小黑盒的回复bot，回复用户@你的评论，不要使用markdown格式，使用plaintext进行回复"
)

var ConfigStruct struct {
	Xhh struct {
		CheckTime   int    `json:"checkTime"`
		ReplyTime   int    `json:"replyTime"`
		Owner       string `json:"owner"`
		DeviceID    string `json:"deviceID"`
		BaseUrl     string `json:"baseUrl"`
		AutoVersion bool   `json:"autoVersion"`
		WebVer      string `json:"webver"`
		Ver         string `json:"version"`
	} `json:"xhh"`
	DataBase struct {
		Type   string `json:"type"`
		Db     string `json:"db"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		User   string `json:"user"`
		Passwd string `json:"passwd"`
	} `json:"database"`
	Ai struct {
		Model             string `json:"model"`
		Prompt            string `json:"prompt"`
		BaseUrl           string `json:"baseUrl"`
		APIKey            string `json:"apiKey"`
		Token             string `json:"token,omitempty"` // legacy: use apiKey for new configs
		WebSearch         bool   `json:"webSearch"`
		ForceWebSearch    bool   `json:"forceWebSearch"`
		SearchContextSize string `json:"searchContextSize"`
		MCP               struct {
			Enabled           bool `json:"enabled"`
			MaxRounds         int  `json:"maxRounds"`
			ToolCallTimeLimit int  `json:"toolCallTimeLimit"`
			UseOSEnv          bool `json:"useOSEnv"`
			MCPServers        map[string]struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env"`
			} `json:"mcpServers"`
		} `json:"mcp"`
	} `json:"ai"`
}

func applyDefaults() {
	ConfigStruct.Xhh.CheckTime = 10
	ConfigStruct.Xhh.ReplyTime = 30
	ConfigStruct.Xhh.BaseUrl = "https://api.xiaoheihe.cn"
	ConfigStruct.Xhh.AutoVersion = true
	ConfigStruct.Xhh.WebVer = DefaultXHHWebVersion
	ConfigStruct.Xhh.Ver = DefaultXHHVersion

	ConfigStruct.DataBase.Type = "sqlite"

	ConfigStruct.Ai.Model = DefaultAIModel
	ConfigStruct.Ai.Prompt = DefaultAIPrompt
	ConfigStruct.Ai.BaseUrl = DefaultAIBaseURL
	ConfigStruct.Ai.MCP.Enabled = false
	ConfigStruct.Ai.MCP.MaxRounds = 10
	ConfigStruct.Ai.MCP.ToolCallTimeLimit = 30
	ConfigStruct.Ai.MCP.UseOSEnv = true
	ConfigStruct.Ai.MCP.MCPServers = make(map[string]struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	})
}

func normalizeConfig() {
	if ConfigStruct.Xhh.CheckTime <= 0 {
		ConfigStruct.Xhh.CheckTime = 10
	}
	if ConfigStruct.Xhh.ReplyTime <= 0 {
		ConfigStruct.Xhh.ReplyTime = 30
	}
	if ConfigStruct.Xhh.BaseUrl == "" {
		ConfigStruct.Xhh.BaseUrl = "https://api.xiaoheihe.cn"
	}
	if ConfigStruct.Xhh.Ver == "" {
		ConfigStruct.Xhh.Ver = DefaultXHHVersion
	}
	if ConfigStruct.Xhh.WebVer == "" {
		ConfigStruct.Xhh.WebVer = DefaultXHHWebVersion
	}
	if ConfigStruct.DataBase.Type == "" {
		ConfigStruct.DataBase.Type = "sqlite"
	}
	if ConfigStruct.Ai.Model == "" {
		ConfigStruct.Ai.Model = DefaultAIModel
	}
	if ConfigStruct.Ai.Prompt == "" {
		ConfigStruct.Ai.Prompt = DefaultAIPrompt
	}
	if ConfigStruct.Ai.BaseUrl == "" {
		ConfigStruct.Ai.BaseUrl = DefaultAIBaseURL
	}
	if ConfigStruct.Ai.APIKey == "" && ConfigStruct.Ai.Token != "" {
		ConfigStruct.Ai.APIKey = ConfigStruct.Ai.Token
	}
	if ConfigStruct.Ai.Token == "" && ConfigStruct.Ai.APIKey != "" {
		ConfigStruct.Ai.Token = ConfigStruct.Ai.APIKey
	}
	if ConfigStruct.Ai.MCP.MaxRounds <= 0 {
		ConfigStruct.Ai.MCP.MaxRounds = 10
	}
	if ConfigStruct.Ai.MCP.ToolCallTimeLimit <= 0 {
		ConfigStruct.Ai.MCP.ToolCallTimeLimit = 30
	}
	if ConfigStruct.Ai.MCP.MCPServers == nil {
		ConfigStruct.Ai.MCP.MCPServers = make(map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
		})
	}
}

func AIKey() string {
	if ConfigStruct.Ai.APIKey != "" {
		return ConfigStruct.Ai.APIKey
	}
	return ConfigStruct.Ai.Token
}

func InitConfig() {
	applyDefaults()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	file, err := os.ReadFile(wd + "/config.json")
	if err != nil {
		if os.IsNotExist(err) {
			data, marshalErr := json.MarshalIndent(ConfigStruct, "", "  ")
			if marshalErr != nil {
				panic(marshalErr)
			}
			if writeErr := os.WriteFile("./config.json", data, 0600); writeErr != nil {
				panic(writeErr)
			}
			loger.Loger.Fatal("已生成 config.json。AI 默认使用 DeepSeek Flash（OpenAI 兼容格式），只需填写 ai.apiKey；xhh.owner 等业务项按需配置后重新启动")
		}
		panic(err)
	}
	if err = json.Unmarshal(file, &ConfigStruct); err != nil {
		panic(err)
	}
	normalizeConfig()
	loger.Loger.Info("[CFG]Init OK")
}
