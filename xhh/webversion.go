package xhh

import (
	"context"
	"net/http"
	"time"

	"xhhrobot/config"
	"xhhrobot/loger"
	"xhhrobot/webversion"

	"go.uber.org/zap"
)

func RefreshWebClientVersion() {
	cfg := &config.ConfigStruct.Xhh
	if !cfg.AutoVersion {
		loger.Loger.Info("[XHH]使用手工 Web 版本参数", zap.String("version", cfg.Ver), zap.String("web_version", cfg.WebVer))
		return
	}

	// Auto mode never depends on stale values left in an old config file.
	cfg.Ver = webversion.DefaultVersion
	cfg.WebVer = webversion.DefaultWebVersion

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 5 * time.Second}
	versions, err := webversion.Detect(ctx, client, nil)
	if err != nil {
		loger.Loger.Warn(
			"[XHH]自动获取 Web 版本失败，使用内置回退值",
			zap.String("version", cfg.Ver),
			zap.String("web_version", cfg.WebVer),
			zap.Error(err),
		)
		return
	}
	if versions.Version != "" {
		cfg.Ver = versions.Version
	}
	if versions.WebVersion != "" {
		cfg.WebVer = versions.WebVersion
	}
	loger.Loger.Info(
		"[XHH]已自动获取 Web 版本参数",
		zap.String("version", cfg.Ver),
		zap.String("web_version", cfg.WebVer),
		zap.String("source", versions.Source),
	)
}
