package config

import (
	"os"
	"os/exec"
	"testing"
)

// TestLookPath_InPath verifies that lookPath finds binaries already in PATH.
func TestLookPath_InPath(t *testing.T) {
	p, err := lookPath("ls")
	if err != nil {
		t.Fatalf("expected to find ls, got error: %v", err)
	}
	if p == "" {
		t.Fatal("expected non-empty path for ls")
	}
}

// TestLookPath_NotExist verifies that lookPath returns an error for missing binaries.
func TestLookPath_NotExist(t *testing.T) {
	_, err := lookPath("nonexistent-binary-xyz-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

// TestLookPath_LoginShellFallback reproduces the daemon scenario:
// PATH is stripped to system-only dirs (no nvm), so exec.LookPath fails,
// but lookPath resolves claude via login shell fallback.
func TestLookPath_LoginShellFallback(t *testing.T) {
	// Precondition: claude must be discoverable via login shell (i.e. nvm in .zshrc)
	fullPath, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude not installed, skipping login shell fallback test")
	}

	// Simulate daemon environment: strip PATH to system-only dirs
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin")
	defer os.Setenv("PATH", origPath)

	// Reproduce the bug: exec.LookPath must fail under stripped PATH
	_, err = exec.LookPath("claude")
	if err == nil {
		t.Skip("claude found in minimal PATH, cannot reproduce nvm issue")
	}

	// Verify fix: lookPath should find claude via login shell
	p, err := lookPath("claude")
	if err != nil {
		t.Fatalf("lookPath should find claude via login shell, got: %v", err)
	}
	if p != fullPath {
		t.Logf("resolved path differs: direct=%s, login-shell=%s (acceptable)", fullPath, p)
	}
	t.Logf("lookPath resolved claude via login shell: %s", p)
}

// TestDetectAndConfigure_StrippedPath is an end-to-end test:
// empty config + stripped PATH → DetectAndConfigure should still find claude.
func TestDetectAndConfigure_StrippedPath(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not installed, skipping")
	}

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin")
	defer os.Setenv("PATH", origPath)

	cfg := DefaultConfig()
	DetectAndConfigure(cfg)

	agent, ok := cfg.Agents["claude"]
	if !ok {
		t.Fatal("expected claude to be detected via login shell fallback")
	}
	if agent.Type != "cli" {
		t.Fatalf("expected type=cli, got %s", agent.Type)
	}
	t.Logf("detected claude: type=%s, command=%s", agent.Type, agent.Command)
}

// TestLoadMinimaxProvider verifies MiniMax provider preset resolution from env.
func TestLoadMinimaxProvider(t *testing.T) {
	t.Run("disabled without api key", func(t *testing.T) {
		t.Setenv("MINIMAX_API_KEY", "")
		if endpoint, _, _ := loadMinimaxProvider(); endpoint != "" {
			t.Fatalf("expected empty endpoint without API key, got %q", endpoint)
		}
	})

	t.Run("default region and model", func(t *testing.T) {
		t.Setenv("MINIMAX_API_KEY", "test-key")
		t.Setenv("MINIMAX_REGION", "")
		t.Setenv("MINIMAX_BASE_URL", "")
		t.Setenv("MINIMAX_MODEL", "")
		endpoint, apiKey, model := loadMinimaxProvider()
		if endpoint != "https://api.minimax.io/v1/chat/completions" {
			t.Fatalf("endpoint = %q", endpoint)
		}
		if apiKey != "test-key" {
			t.Fatalf("apiKey = %q", apiKey)
		}
		if model != "MiniMax-M3" {
			t.Fatalf("model = %q", model)
		}
	})

	t.Run("cn_zh region", func(t *testing.T) {
		t.Setenv("MINIMAX_API_KEY", "test-key")
		t.Setenv("MINIMAX_BASE_URL", "")
		t.Setenv("MINIMAX_REGION", "cn_zh")
		if endpoint, _, _ := loadMinimaxProvider(); endpoint != "https://api.minimaxi.com/v1/chat/completions" {
			t.Fatalf("cn endpoint = %q", endpoint)
		}
	})

	t.Run("base url and model override", func(t *testing.T) {
		t.Setenv("MINIMAX_API_KEY", "test-key")
		t.Setenv("MINIMAX_BASE_URL", "https://proxy.example.com/v1/chat/completions/")
		t.Setenv("MINIMAX_MODEL", "MiniMax-M2.7")
		endpoint, _, model := loadMinimaxProvider()
		if endpoint != "https://proxy.example.com/v1/chat/completions" {
			t.Fatalf("override endpoint = %q", endpoint)
		}
		if model != "MiniMax-M2.7" {
			t.Fatalf("override model = %q", model)
		}
	})
}

// TestDetectAndConfigure_Minimax verifies MINIMAX_API_KEY enables the HTTP provider.
func TestDetectAndConfigure_Minimax(t *testing.T) {
	t.Setenv("MINIMAX_API_KEY", "test-key")
	t.Setenv("MINIMAX_REGION", "")
	t.Setenv("MINIMAX_BASE_URL", "")
	t.Setenv("MINIMAX_MODEL", "")

	cfg := DefaultConfig()
	DetectAndConfigure(cfg)

	ag, ok := cfg.Agents["minimax"]
	if !ok {
		t.Fatal("expected minimax agent to be configured when MINIMAX_API_KEY is set")
	}
	if ag.Type != "http" {
		t.Fatalf("type = %q, want http", ag.Type)
	}
	if ag.Endpoint != "https://api.minimax.io/v1/chat/completions" {
		t.Fatalf("endpoint = %q", ag.Endpoint)
	}
	if ag.Model != "MiniMax-M3" {
		t.Fatalf("model = %q", ag.Model)
	}
}

// TestDetectAndConfigure_MinimaxDisabled verifies no provider without the key.
func TestDetectAndConfigure_MinimaxDisabled(t *testing.T) {
	t.Setenv("MINIMAX_API_KEY", "")

	cfg := DefaultConfig()
	DetectAndConfigure(cfg)

	if _, ok := cfg.Agents["minimax"]; ok {
		t.Fatal("did not expect minimax agent without MINIMAX_API_KEY")
	}
}
