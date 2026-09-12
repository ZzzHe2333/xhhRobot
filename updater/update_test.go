package updater

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectAsset(t *testing.T) {
	assets := []Asset{
		{Name: "linux-amd-64.zip", BrowserDownloadURL: "https://example.com/linux.zip"},
		{Name: "windows-amd-64.zip", BrowserDownloadURL: "https://example.com/windows.zip"},
		{Name: "windows-arm64.zip", BrowserDownloadURL: "https://example.com/windows-arm64.zip"},
	}

	tests := []struct {
		goos   string
		arch   string
		wanted string
	}{
		{"windows", "amd64", "windows-amd-64.zip"},
		{"windows", "arm64", "windows-arm64.zip"},
		{"linux", "amd64", "linux-amd-64.zip"},
	}

	for _, tt := range tests {
		asset, err := selectAsset(assets, tt.goos, tt.arch)
		if err != nil {
			t.Fatalf("selectAsset(%s/%s) returned error: %v", tt.goos, tt.arch, err)
		}
		if asset.Name != tt.wanted {
			t.Fatalf("selectAsset(%s/%s) = %q, want %q", tt.goos, tt.arch, asset.Name, tt.wanted)
		}
	}
}

func TestVerifyDigest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.zip")
	content := []byte("xhhRobot update package")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	if err := verifyDigest(path, digest); err != nil {
		t.Fatalf("valid digest rejected: %v", err)
	}
	if err := verifyDigest(path, "sha256:"+string(make([]byte, 64))); err == nil {
		t.Fatal("invalid digest should fail")
	}
}

func TestExtractZipRejectsZipSlip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "bad.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../outside.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if err := extractZip(archivePath, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("Zip Slip archive should be rejected")
	}
}

func TestExtractZipAndFindExecutable(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "good.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "xhhRobot.exe", Method: zip.Deflate}
	header.SetMode(0755)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("binary")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "out")
	if err := extractZip(archivePath, out); err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}
	path, err := findExecutable(out, "xhhRobot.exe", "windows")
	if err != nil {
		t.Fatalf("findExecutable failed: %v", err)
	}
	if filepath.Base(path) != "xhhRobot.exe" {
		t.Fatalf("unexpected executable: %s", path)
	}
}
