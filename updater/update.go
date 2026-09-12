package updater

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultRepository = "ZzzHe2333/xhhRobot"

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
	Size               int64  `json:"size"`
}

var httpClient = &http.Client{Timeout: 10 * time.Minute}

func Run() error {
	repository := strings.TrimSpace(os.Getenv("XHHROBOT_UPDATE_REPO"))
	if repository == "" {
		repository = DefaultRepository
	}

	fmt.Printf("正在检查云端最新版本：%s\n", repository)
	release, err := fetchLatestRelease(repository)
	if err != nil {
		return err
	}

	asset, err := selectAsset(release.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("最新版本 %s：%w", release.TagName, err)
	}

	fmt.Printf("发现最新版本：%s\n", release.TagName)
	fmt.Printf("更新包：%s（%.2f MB）\n", asset.Name, float64(asset.Size)/1024/1024)

	tempDir, err := os.MkdirTemp("", "xhhrobot-update-*")
	if err != nil {
		return fmt.Errorf("创建更新临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, asset.Name)
	fmt.Println("正在下载更新包...")
	if err := downloadFile(asset.BrowserDownloadURL, archivePath); err != nil {
		return err
	}

	if asset.Digest != "" {
		fmt.Println("正在校验更新包...")
		if err := verifyDigest(archivePath, asset.Digest); err != nil {
			return err
		}
	}

	extractDir := filepath.Join(tempDir, "package")
	if err := extractZip(archivePath, extractDir); err != nil {
		return fmt.Errorf("解压更新包失败: %w", err)
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("解析当前程序路径失败: %w", err)
	}

	newExe, err := findExecutable(extractDir, filepath.Base(currentExe), runtime.GOOS)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		if err := installWindows(newExe, currentExe); err != nil {
			return err
		}
		fmt.Println("更新包已准备完成。程序退出后会自动替换为最新版并重新启动。")
		return nil
	}

	if err := installUnix(newExe, currentExe); err != nil {
		return err
	}
	fmt.Println("更新完成，请重新启动 xhhRobot。")
	return nil
}

func fetchLatestRelease(repository string) (Release, error) {
	url := "https://api.github.com/repos/" + repository + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, fmt.Errorf("创建更新检查请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "xhhRobot-updater")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("访问云端更新源失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("更新源 %s 尚未发布 Release，暂无可下载的更新包", repository)
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("检查更新失败，GitHub 返回 HTTP %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("解析最新版本信息失败: %w", err)
	}
	if release.TagName == "" {
		return Release{}, errors.New("最新 Release 缺少版本标签")
	}
	return release, nil
}

func selectAsset(assets []Asset, goos, goarch string) (Asset, error) {
	osTokens := map[string][]string{
		"windows": {"windows", "win"},
		"linux":   {"linux"},
		"darwin":  {"darwin", "macos", "mac-os", "mac"},
	}
	archTokens := map[string][]string{
		"amd64": {"amd64", "amd-64", "x86_64", "x86-64", "x64"},
		"arm64": {"arm64", "arm-64", "aarch64"},
	}

	oses, ok := osTokens[goos]
	if !ok {
		return Asset{}, fmt.Errorf("暂不支持当前操作系统 %s", goos)
	}
	arches, ok := archTokens[goarch]
	if !ok {
		return Asset{}, fmt.Errorf("暂不支持当前 CPU 架构 %s", goarch)
	}

	bestScore := -1
	var best Asset
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if !strings.HasSuffix(name, ".zip") {
			continue
		}
		if !containsAny(name, oses) || !containsAny(name, arches) {
			continue
		}

		score := 0
		if strings.HasPrefix(name, goos) {
			score += 2
		}
		if strings.Contains(name, goarch) {
			score += 2
		}
		if strings.Contains(name, "xhhrobot") {
			score++
		}
		if score > bestScore {
			bestScore = score
			best = asset
		}
	}

	if bestScore < 0 {
		return Asset{}, fmt.Errorf("没有找到适用于 %s/%s 的 ZIP 更新包", goos, goarch)
	}
	if best.BrowserDownloadURL == "" {
		return Asset{}, fmt.Errorf("更新包 %s 缺少下载地址", best.Name)
	}
	return best, nil
}

func containsAny(value string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func downloadFile(url, destination string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建更新下载请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "xhhRobot-updater")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载更新包失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载更新包失败，服务器返回 HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("创建更新包文件失败: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("保存更新包失败: %w", err)
	}
	return nil
}

func verifyDigest(path, digest string) error {
	const prefix = "sha256:"
	if !strings.HasPrefix(strings.ToLower(digest), prefix) {
		return fmt.Errorf("不支持的更新包摘要格式 %q", digest)
	}
	want := strings.TrimSpace(digest[len(prefix):])
	if len(want) != sha256.Size*2 {
		return errors.New("更新包 SHA-256 摘要长度无效")
	}
	if _, err := hex.DecodeString(want); err != nil {
		return fmt.Errorf("更新包 SHA-256 摘要无效: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("读取更新包失败: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("计算更新包 SHA-256 失败: %w", err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("更新包 SHA-256 校验失败：期望 %s，实际 %s", want, got)
	}
	return nil
}

func extractZip(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	cleanDestination, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cleanDestination, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("更新包包含不允许的符号链接: %s", file.Name)
		}
		target := filepath.Join(cleanDestination, file.Name)
		cleanTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if cleanTarget != cleanDestination && !strings.HasPrefix(cleanTarget, cleanDestination+string(os.PathSeparator)) {
			return fmt.Errorf("更新包包含非法路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, 0755); err != nil {
				return err
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return fmt.Errorf("更新包包含不支持的文件类型: %s", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		mode := file.Mode().Perm()
		if mode == 0 {
			mode = 0644
		}
		dst, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close()
		src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func findExecutable(root, currentBase, goos string) (string, error) {
	var exact string
	var candidates []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		base := filepath.Base(path)
		if strings.EqualFold(base, currentBase) {
			exact = path
			return nil
		}

		if goos == "windows" {
			if strings.EqualFold(filepath.Ext(base), ".exe") {
				candidates = append(candidates, path)
			}
			return nil
		}

		if info.Mode().Perm()&0111 != 0 || strings.Contains(strings.ToLower(base), "xhhrobot") {
			candidates = append(candidates, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("查找新版程序失败: %w", err)
	}
	if exact != "" {
		return exact, nil
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) == 0 {
		return "", errors.New("更新包中未找到可执行程序")
	}
	return "", fmt.Errorf("更新包中找到多个可执行程序，无法确定主程序: %v", candidates)
}

func stageExecutable(source, currentExe string) (string, error) {
	staged := currentExe + ".update-new"
	if err := os.Remove(staged); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	src, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer src.Close()

	mode := os.FileMode(0755)
	if info, err := os.Stat(currentExe); err == nil {
		mode = info.Mode().Perm()
	}
	dst, err := os.OpenFile(staged, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return "", fmt.Errorf("无法在程序目录写入更新文件: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(staged)
		return "", err
	}
	if err := dst.Close(); err != nil {
		os.Remove(staged)
		return "", err
	}
	if err := os.Chmod(staged, mode); err != nil {
		os.Remove(staged)
		return "", err
	}
	return staged, nil
}

func installUnix(newExe, currentExe string) error {
	staged, err := stageExecutable(newExe, currentExe)
	if err != nil {
		return fmt.Errorf("准备新版程序失败: %w", err)
	}
	if err := os.Rename(staged, currentExe); err != nil {
		os.Remove(staged)
		return fmt.Errorf("替换当前程序失败: %w", err)
	}
	return nil
}

func installWindows(newExe, currentExe string) error {
	staged, err := stageExecutable(newExe, currentExe)
	if err != nil {
		return fmt.Errorf("准备新版程序失败: %w", err)
	}

	scriptPath := currentExe + ".update.cmd"
	script := fmt.Sprintf(`@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -Command "try { Wait-Process -Id %d -ErrorAction SilentlyContinue } catch {}"
for /L %%%%I in (1,1,30) do (
  copy /Y "%s" "%s" >nul 2>&1 && goto updated
  timeout /T 1 /NOBREAK >nul
)
exit /b 1
:updated
del /Q "%s" >nul 2>&1
start "" "%s"
del /Q "%%~f0" >nul 2>&1
`, os.Getpid(), staged, currentExe, staged, currentExe)

	if err := os.WriteFile(scriptPath, []byte(script), 0600); err != nil {
		os.Remove(staged)
		return fmt.Errorf("创建 Windows 更新辅助脚本失败: %w", err)
	}

	cmd := exec.Command("cmd.exe", "/C", scriptPath)
	if err := cmd.Start(); err != nil {
		os.Remove(staged)
		os.Remove(scriptPath)
		return fmt.Errorf("启动 Windows 更新辅助程序失败: %w", err)
	}
	return nil
}
