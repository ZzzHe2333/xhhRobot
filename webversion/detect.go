package webversion

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	DefaultVersion    = "999.0.4"
	DefaultWebVersion = "3.0"
	maxPageBytes      = 2 << 20
	maxScriptBytes    = 8 << 20
	maxScripts        = 24
)

var DefaultRoots = []string{
	"https://www.xiaoheihe.cn/",
	"https://xiaoheihe.cn/",
}

type Versions struct {
	Version    string
	WebVersion string
	Source     string
}

var (
	scriptSrcRE = regexp.MustCompile(`(?is)<script[^>]+src\s*=\s*["']([^"']+)["'][^>]*>`)
	webVerRE    = regexp.MustCompile(`(?i)["']?web_version["']?\s*[:=]\s*["']([0-9]+(?:\.[0-9]+){1,3})["']`)
	versionRE   = regexp.MustCompile(`(?i)(?:^|[^a-z0-9_])["']?version["']?\s*[:=]\s*["']([0-9]+(?:\.[0-9]+){1,3})["']`)
	queryWebRE  = regexp.MustCompile(`(?i)(?:[?&]|\\u0026|&amp;)web_version=([0-9]+(?:\.[0-9]+){1,3})`)
	queryVerRE  = regexp.MustCompile(`(?i)(?:[?&]|\\u0026|&amp;)version=([0-9]+(?:\.[0-9]+){1,3})`)
)

func Detect(ctx context.Context, client *http.Client, roots []string) (Versions, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if len(roots) == 0 {
		roots = DefaultRoots
	}

	var errs []error
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		versions, err := detectRoot(ctx, client, root)
		if err == nil {
			return versions, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", root, err))
	}
	if len(errs) == 0 {
		return Versions{}, errors.New("没有可用的小黑盒 Web 地址")
	}
	return Versions{}, errors.Join(errs...)
}

func detectRoot(ctx context.Context, client *http.Client, root string) (Versions, error) {
	body, finalURL, err := fetchText(ctx, client, root, maxPageBytes)
	if err != nil {
		return Versions{}, err
	}
	if versions, ok := Parse(body); ok {
		versions.Source = finalURL
		return versions, nil
	}
	partialVersion, partialWebVersion := parsePartial(body)

	base, err := url.Parse(finalURL)
	if err != nil {
		return Versions{}, fmt.Errorf("解析小黑盒页面地址失败: %w", err)
	}

	scripts := scriptSrcRE.FindAllStringSubmatch(body, -1)
	seen := make(map[string]struct{}, len(scripts))
	checked := 0
	for _, match := range scripts {
		if len(match) < 2 || checked >= maxScripts {
			break
		}
		scriptURL, err := resolveScriptURL(base, match[1])
		if err != nil {
			continue
		}
		if _, exists := seen[scriptURL]; exists {
			continue
		}
		seen[scriptURL] = struct{}{}
		checked++

		script, _, err := fetchText(ctx, client, scriptURL, maxScriptBytes)
		if err != nil {
			continue
		}
		if versions, ok := Parse(script); ok {
			versions.Source = scriptURL
			return versions, nil
		}
		version, webVersion := parsePartial(script)
		if partialVersion == "" && version != "" {
			partialVersion = version
		}
		if partialWebVersion == "" && webVersion != "" {
			partialWebVersion = webVersion
		}
		if partialVersion != "" && partialWebVersion != "" {
			return Versions{Version: partialVersion, WebVersion: partialWebVersion, Source: scriptURL}, nil
		}
	}
	return Versions{}, errors.New("在首页及脚本资源中未识别到 version/web_version")
}

func fetchText(ctx context.Context, client *http.Client, target string, limit int64) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/153 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/javascript,text/javascript,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %w", err)
	}
	if int64(len(data)) > limit {
		return "", "", fmt.Errorf("响应超过 %d 字节限制", limit)
	}
	finalURL := target
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return string(data), finalURL, nil
}

func resolveScriptURL(base *url.URL, src string) (string, error) {
	src = html.UnescapeString(strings.TrimSpace(strings.ReplaceAll(src, `\u0026`, "&")))
	ref, err := url.Parse(src)
	if err != nil {
		return "", err
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", fmt.Errorf("unsupported script scheme %q", resolved.Scheme)
	}
	return resolved.String(), nil
}

func Parse(text string) (Versions, bool) {
	if text == "" {
		return Versions{}, false
	}
	// Query-string form is the strongest signal because these values are used
	// directly by fetch/XHR requests.
	if v, w, ok := parsePair(text, queryVerRE, queryWebRE); ok {
		return Versions{Version: v, WebVersion: w}, true
	}
	if v, w, ok := parsePair(text, versionRE, webVerRE); ok {
		return Versions{Version: v, WebVersion: w}, true
	}
	return Versions{}, false
}

func parsePartial(text string) (string, string) {
	version := selectClientVersion(append(queryVerRE.FindAllStringSubmatch(text, -1), versionRE.FindAllStringSubmatch(text, -1)...))
	if !strings.HasPrefix(version, "999.") {
		version = ""
	}
	webVersion := ""
	if match := queryWebRE.FindStringSubmatch(text); len(match) > 1 {
		webVersion = match[1]
	} else if match := webVerRE.FindStringSubmatch(text); len(match) > 1 {
		webVersion = match[1]
	}
	return version, webVersion
}

func parsePair(text string, versionPattern, webPattern *regexp.Regexp) (string, string, bool) {
	webMatches := webPattern.FindAllStringSubmatchIndex(text, -1)
	for _, wm := range webMatches {
		if len(wm) < 4 {
			continue
		}
		start := wm[0] - 1200
		if start < 0 {
			start = 0
		}
		end := wm[1] + 1200
		if end > len(text) {
			end = len(text)
		}
		window := text[start:end]
		versionValue := selectClientVersion(versionPattern.FindAllStringSubmatch(window, -1))
		if versionValue == "" {
			continue
		}
		webValue := text[wm[2]:wm[3]]
		if validVersion(versionValue) && validVersion(webValue) {
			return versionValue, webValue, true
		}
	}
	return "", "", false
}

func selectClientVersion(matches [][]string) string {
	var fallback string
	for _, match := range matches {
		if len(match) < 2 || !validVersion(match[1]) {
			continue
		}
		value := match[1]
		if strings.HasPrefix(value, "999.") {
			return value
		}
		if fallback == "" {
			fallback = value
		}
	}
	return fallback
}

func validVersion(value string) bool {
	parts := strings.Split(value, ".")
	return len(parts) >= 2 && len(parts) <= 4
}
