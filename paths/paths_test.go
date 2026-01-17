package paths_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/grokify/oscompat/paths"
)

func TestHome(t *testing.T) {
	home, err := paths.Home()
	if err != nil {
		t.Fatalf("Home() error: %v", err)
	}
	if home == "" {
		t.Error("Home() returned empty string")
	}
	if !filepath.IsAbs(home) {
		t.Errorf("Home() returned non-absolute path: %s", home)
	}
}

func TestUserConfig(t *testing.T) {
	dir, err := paths.UserConfig()
	if err != nil {
		t.Fatalf("UserConfig() error: %v", err)
	}
	if dir == "" {
		t.Error("UserConfig() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("UserConfig() returned non-absolute path: %s", dir)
	}
}

func TestUserData(t *testing.T) {
	dir, err := paths.UserData()
	if err != nil {
		t.Fatalf("UserData() error: %v", err)
	}
	if dir == "" {
		t.Error("UserData() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("UserData() returned non-absolute path: %s", dir)
	}
}

func TestUserCache(t *testing.T) {
	dir, err := paths.UserCache()
	if err != nil {
		t.Fatalf("UserCache() error: %v", err)
	}
	if dir == "" {
		t.Error("UserCache() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("UserCache() returned non-absolute path: %s", dir)
	}
}

func TestUserRuntime(t *testing.T) {
	dir, err := paths.UserRuntime()
	if err != nil {
		t.Fatalf("UserRuntime() error: %v", err)
	}
	if dir == "" {
		t.Error("UserRuntime() returned empty string")
	}
}

func TestSystemConfig(t *testing.T) {
	dir, err := paths.SystemConfig()
	if err != nil {
		t.Fatalf("SystemConfig() error: %v", err)
	}
	if dir == "" {
		t.Error("SystemConfig() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		// Should be %ProgramData% or C:\ProgramData
		if !strings.Contains(strings.ToLower(dir), "programdata") {
			t.Errorf("SystemConfig() on Windows expected ProgramData, got: %s", dir)
		}
	default:
		// Should be /etc on Unix-like systems
		if dir != "/etc" {
			t.Errorf("SystemConfig() on Unix expected /etc, got: %s", dir)
		}
	}
}

func TestAppConfig(t *testing.T) {
	// Use a temp directory to avoid polluting real config
	tmpDir := t.TempDir()

	// Override the base directory for testing
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	case "darwin":
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	default:
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	}

	dir, err := paths.AppConfig("testapp")
	if err != nil {
		t.Fatalf("AppConfig() error: %v", err)
	}

	if !strings.HasSuffix(dir, "testapp") {
		t.Errorf("AppConfig() should end with app name, got: %s", dir)
	}

	// Verify directory was created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("AppConfig() directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("AppConfig() did not create a directory")
	}
}

func TestAppConfigEmptyName(t *testing.T) {
	_, err := paths.AppConfig("")
	if err != paths.ErrInvalidAppName {
		t.Errorf("AppConfig('') expected ErrInvalidAppName, got: %v", err)
	}
}

func TestAppData(t *testing.T) {
	tmpDir := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		t.Setenv("LOCALAPPDATA", tmpDir)
	case "darwin":
		t.Setenv("XDG_DATA_HOME", tmpDir)
	default:
		t.Setenv("XDG_DATA_HOME", tmpDir)
	}

	dir, err := paths.AppData("testapp")
	if err != nil {
		t.Fatalf("AppData() error: %v", err)
	}

	if !strings.HasSuffix(dir, "testapp") {
		t.Errorf("AppData() should end with app name, got: %s", dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("AppData() directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("AppData() did not create a directory")
	}
}

func TestAppDataEmptyName(t *testing.T) {
	_, err := paths.AppData("")
	if err != paths.ErrInvalidAppName {
		t.Errorf("AppData('') expected ErrInvalidAppName, got: %v", err)
	}
}

func TestAppCache(t *testing.T) {
	tmpDir := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		t.Setenv("LOCALAPPDATA", tmpDir)
	case "darwin":
		t.Setenv("XDG_CACHE_HOME", tmpDir)
	default:
		t.Setenv("XDG_CACHE_HOME", tmpDir)
	}

	dir, err := paths.AppCache("testapp")
	if err != nil {
		t.Fatalf("AppCache() error: %v", err)
	}

	if !strings.HasSuffix(dir, "testapp") {
		t.Errorf("AppCache() should end with app name, got: %s", dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("AppCache() directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("AppCache() did not create a directory")
	}
}

func TestAppCacheEmptyName(t *testing.T) {
	_, err := paths.AppCache("")
	if err != paths.ErrInvalidAppName {
		t.Errorf("AppCache('') expected ErrInvalidAppName, got: %v", err)
	}
}

func TestAppRuntime(t *testing.T) {
	tmpDir := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		t.Setenv("LOCALAPPDATA", tmpDir)
	default:
		t.Setenv("XDG_RUNTIME_DIR", tmpDir)
	}

	dir, err := paths.AppRuntime("testapp")
	if err != nil {
		t.Fatalf("AppRuntime() error: %v", err)
	}

	if !strings.HasSuffix(dir, "testapp") {
		t.Errorf("AppRuntime() should end with app name, got: %s", dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("AppRuntime() directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("AppRuntime() did not create a directory")
	}
}

func TestAppRuntimeEmptyName(t *testing.T) {
	_, err := paths.AppRuntime("")
	if err != paths.ErrInvalidAppName {
		t.Errorf("AppRuntime('') expected ErrInvalidAppName, got: %v", err)
	}
}

func TestSystemAppConfig(t *testing.T) {
	dir, err := paths.SystemAppConfig("testapp")
	if err != nil {
		t.Fatalf("SystemAppConfig() error: %v", err)
	}

	if !strings.HasSuffix(dir, "testapp") {
		t.Errorf("SystemAppConfig() should end with app name, got: %s", dir)
	}

	// Note: We don't verify directory creation because system config
	// directories typically require elevated privileges to create
}

func TestSystemAppConfigEmptyName(t *testing.T) {
	_, err := paths.SystemAppConfig("")
	if err != paths.ErrInvalidAppName {
		t.Errorf("SystemAppConfig('') expected ErrInvalidAppName, got: %v", err)
	}
}

func TestXDGOverrides(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG overrides not applicable on Windows")
	}

	tmpDir := t.TempDir()

	tests := []struct {
		envVar string
		fn     func() (string, error)
	}{
		{"XDG_CONFIG_HOME", paths.UserConfig},
		{"XDG_DATA_HOME", paths.UserData},
		{"XDG_CACHE_HOME", paths.UserCache},
		{"XDG_RUNTIME_DIR", paths.UserRuntime},
	}

	for _, tt := range tests {
		t.Run(tt.envVar, func(t *testing.T) {
			customDir := filepath.Join(tmpDir, tt.envVar)
			t.Setenv(tt.envVar, customDir)

			dir, err := tt.fn()
			if err != nil {
				t.Fatalf("%s override error: %v", tt.envVar, err)
			}
			if dir != customDir {
				t.Errorf("%s override: got %s, want %s", tt.envVar, dir, customDir)
			}
		})
	}
}
