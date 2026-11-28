package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetSuffixFromUrl 回归：url 不含点时老实现会 panic（LastIndex 返回 -1 直接切片）
func TestGetSuffixFromUrl(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://a.b/x.png", ".png"},
		{"https://a.b/x.png?v=1", ".png"},
		{"https://a.b/x.PNG", ".PNG"},
		{"https://example.com", ""},
		{"https://a.b/no-ext", ""},
		{"", ""},
		{"no-dot-at-all", ""},
		{".hidden", ""},
	}
	for _, c := range cases {
		if got := GetSuffixFromUrl(c.in); got != c.want {
			t.Fatalf("GetSuffixFromUrl(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestDataDirAndFilesDir 数据目录默认 ./data，可用环境变量覆盖；
// 未设置 NAV_FILES_DIR 时用户文件目录跟随数据目录（飞牛原生安装依赖这个行为）
func TestDataDirAndFilesDir(t *testing.T) {
	t.Setenv("NAV_DATA_DIR", "")
	t.Setenv("NAV_FILES_DIR", "")

	if got := DataDir(); got != "./data" {
		t.Fatalf("DataDir() = %q, want ./data", got)
	}
	if got := FilesDir(); got != "./data" {
		t.Fatalf("FilesDir() = %q, want ./data", got)
	}

	dataDir := filepath.Join(t.TempDir(), "internal")
	t.Setenv("NAV_DATA_DIR", dataDir+string(os.PathSeparator))
	if got := DataDir(); got != dataDir {
		t.Fatalf("DataDir() = %q, want %q", got, dataDir)
	}
	if got := FilesDir(); got != dataDir {
		t.Fatalf("FilesDir() 未设置时应跟随数据目录，got %q, want %q", got, dataDir)
	}

	filesDir := filepath.Join(t.TempDir(), "files")
	t.Setenv("NAV_FILES_DIR", filesDir)
	if got := FilesDir(); got != filesDir {
		t.Fatalf("FilesDir() = %q, want %q", got, filesDir)
	}
	if got := DataDir(); got != dataDir {
		t.Fatalf("设置 NAV_FILES_DIR 不应影响数据目录，got %q, want %q", got, dataDir)
	}
}
