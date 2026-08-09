package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if cfg.TargetLanguage != "c" {
		t.Errorf("TargetLanguage = %q, want %q", cfg.TargetLanguage, "c")
	}
	if cfg.QueueSize != 100 {
		t.Errorf("QueueSize = %d, want 100", cfg.QueueSize)
	}
	if cfg.SpendableSize != 10 {
		t.Errorf("SpendableSize = %d, want 10", cfg.SpendableSize)
	}
	if cfg.MemoryLimitMB != 4096 {
		t.Errorf("MemoryLimitMB = %d, want 4096", cfg.MemoryLimitMB)
	}
	if cfg.TimeoutSec != 120 {
		t.Errorf("TimeoutSec = %d, want 120", cfg.TimeoutSec)
	}
	if cfg.TemplatePath != "templates" {
		t.Errorf("TemplatePath = %q, want %q", cfg.TemplatePath, "templates")
	}
	if cfg.IncludePath != "../std" {
		t.Errorf("IncludePath = %q, want %q", cfg.IncludePath, "../std")
	}
}

func TestDefaultConfigBasePath(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BasePath == "" {
		t.Error("BasePath should not be empty")
	}
}

func TestResolveOptLevel_Default(t *testing.T) {
	cfg := DefaultConfig()
	level := cfg.ResolveOptLevel("")
	if level != "-O2" {
		t.Errorf("default opt level = %q, want %q", level, "-O2")
	}
}

func TestResolveOptLevel_SOR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SOR = true
	level := cfg.ResolveOptLevel("")
	if level != "-O3" {
		t.Errorf("SOR opt level = %q, want %q", level, "-O3")
	}
}

func TestResolveOptLevel_Release(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Release = true
	level := cfg.ResolveOptLevel("")
	if level != "-O3" {
		t.Errorf("release opt level = %q, want %q", level, "-O3")
	}
}

func TestResolveOptLevel_Override(t *testing.T) {
	cfg := DefaultConfig()
	// Override takes precedence
	level := cfg.ResolveOptLevel("O1")
	if level != "-O1" {
		t.Errorf("override opt level = %q, want %q", level, "-O1")
	}
}

func TestResolveOptLevel_OverrideWithSOR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SOR = true
	// Override should take precedence over SOR
	level := cfg.ResolveOptLevel("O0")
	if level != "-O0" {
		t.Errorf("override with SOR = %q, want %q", level, "-O0")
	}
}

func TestResolveOptLevel_Invalid(t *testing.T) {
	cfg := DefaultConfig()
	level := cfg.ResolveOptLevel("O5")
	if level != "-O2" {
		t.Errorf("invalid opt level = %q, want %q (fallback)", level, "-O2")
	}
}

func TestResolveOptLevel_ReleaseOverrideSOR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SOR = true
	cfg.Release = true
	// Override should take precedence over both
	level := cfg.ResolveOptLevel("O1")
	if level != "-O1" {
		t.Errorf("release+SOR+override = %q, want %q", level, "-O1")
	}
}

func TestOutputFile_Default(t *testing.T) {
	cfg := DefaultConfig()
	path := cfg.OutputFile("/path/to/input.kl")
	if path == "" {
		t.Fatal("OutputFile() returned empty string")
	}
	// Should strip .kl extension
	base := filepath.Base(path)
	if base == "input.kl" {
		t.Errorf("output filename = %q, should not have .kl extension", base)
	}
}

func TestOutputFile_WithOutputDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OutputDir = "/output"
	path := cfg.OutputFile("/path/to/input.kl")
	if !filepath.HasPrefix(path, "/output") {
		t.Errorf("output path = %q, should start with /output", path)
	}
}

func TestSplitList_Empty(t *testing.T) {
	result := splitList("")
	if result != nil {
		t.Errorf("splitList('') = %v, want nil", result)
	}
}

func TestSplitList_Single(t *testing.T) {
	result := splitList("foo")
	if len(result) != 1 || result[0] != "foo" {
		t.Errorf("splitList('foo') = %v, want [foo]", result)
	}
}

func TestSplitList_SpaceSeparated(t *testing.T) {
	result := splitList("foo bar baz")
	if len(result) != 3 {
		t.Errorf("splitList('foo bar baz') = %v, want 3 elements", result)
	}
}

func TestSplitList_CommaSeparated(t *testing.T) {
	result := splitList("foo,bar,baz")
	if len(result) != 3 {
		t.Errorf("splitList('foo,bar,baz') = %v, want 3 elements", result)
	}
}

func TestSplitList_Mixed(t *testing.T) {
	result := splitList("foo, bar, baz")
	if len(result) != 3 {
		t.Errorf("splitList('foo, bar, baz') = %v, want 3 elements", result)
	}
}

func TestSplitList_WhitespaceOnly(t *testing.T) {
	result := splitList("   ")
	if result != nil {
		t.Errorf("splitList('   ') = %v, want nil", result)
	}
}

func TestNormalizePaths_Relative(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TemplatePath = "relative/path"
	normalizePaths(cfg)
	// Should be absolute now
	if !filepath.IsAbs(cfg.TemplatePath) {
		t.Errorf("TemplatePath should be absolute after normalizePaths, got %q", cfg.TemplatePath)
	}
}

func TestNormalizePaths_Absolute(t *testing.T) {
	cfg := DefaultConfig()
	absPath := "/absolute/path"
	cfg.TemplatePath = absPath
	normalizePaths(cfg)
	if cfg.TemplatePath != absPath {
		t.Errorf("TemplatePath = %q, want %q (should remain unchanged)", cfg.TemplatePath, absPath)
	}
}

func TestNormalizePaths_Empty(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TemplatePath = ""
	normalizePaths(cfg)
	// Should not change empty paths
	if cfg.TemplatePath != "" {
		t.Errorf("TemplatePath should remain empty, got %q", cfg.TemplatePath)
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "kaula.json")
	err := GenerateDefaultConfig(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() returned error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("kaula.json was not created")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_config.json")
	cfg := DefaultConfig()
	cfg.TargetLanguage = "rust"
	err := SaveConfig(cfg, path)
	if err != nil {
		t.Fatalf("SaveConfig() returned error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := DefaultConfig()

	// Test setting various fields
	cfg.Freestanding = true
	if !cfg.Freestanding {
		t.Error("Freestanding should be true")
	}

	cfg.SOR = true
	if !cfg.SOR {
		t.Error("SOR should be true")
	}

	cfg.Release = true
	if !cfg.Release {
		t.Error("Release should be true")
	}

	cfg.NoCache = true
	if !cfg.NoCache {
		t.Error("NoCache should be true")
	}

	cfg.SourceMap = true
	if !cfg.SourceMap {
		t.Error("SourceMap should be true")
	}

	cfg.OptLevel = "O3"
	if cfg.OptLevel != "O3" {
		t.Errorf("OptLevel = %q, want %q", cfg.OptLevel, "O3")
	}
}

func TestConfig_PlatformFields(t *testing.T) {
	cfg := DefaultConfig()

	cfg.TargetTriple = "x86_64-unknown-elf"
	if cfg.TargetTriple != "x86_64-unknown-elf" {
		t.Errorf("TargetTriple = %q", cfg.TargetTriple)
	}

	cfg.Boot = "multiboot"
	if cfg.Boot != "multiboot" {
		t.Errorf("Boot = %q", cfg.Boot)
	}

	cfg.BootArch = "x86_64"
	if cfg.BootArch != "x86_64" {
		t.Errorf("BootArch = %q", cfg.BootArch)
	}

	cfg.Entry = "_start"
	if cfg.Entry != "_start" {
		t.Errorf("Entry = %q", cfg.Entry)
	}

	cfg.OutputFormat = "elf"
	if cfg.OutputFormat != "elf" {
		t.Errorf("OutputFormat = %q", cfg.OutputFormat)
	}
}

func TestConfig_CLibFields(t *testing.T) {
	cfg := DefaultConfig()

	cfg.CFlags = []string{"-Wall", "-O2"}
	if len(cfg.CFlags) != 2 {
		t.Errorf("CFlags len = %d, want 2", len(cfg.CFlags))
	}

	cfg.CDefines = []string{"DEBUG", "VERSION=2"}
	if len(cfg.CDefines) != 2 {
		t.Errorf("CDefines len = %d, want 2", len(cfg.CDefines))
	}

	cfg.CLibs = []string{"m", "pthread"}
	if len(cfg.CLibs) != 2 {
		t.Errorf("CLibs len = %d, want 2", len(cfg.CLibs))
	}
}

func TestConfig_AnalyzeFields(t *testing.T) {
	cfg := DefaultConfig()

	cfg.AnalyzePkg = "cjson"
	if cfg.AnalyzePkg != "cjson" {
		t.Errorf("AnalyzePkg = %q", cfg.AnalyzePkg)
	}

	cfg.AnalyzePkgAll = true
	if !cfg.AnalyzePkgAll {
		t.Error("AnalyzePkgAll should be true")
	}

	cfg.BuildPkglib = "all"
	if cfg.BuildPkglib != "all" {
		t.Errorf("BuildPkglib = %q", cfg.BuildPkglib)
	}

	cfg.ForcePKG = true
	if !cfg.ForcePKG {
		t.Error("ForcePKG should be true")
	}
}

func TestConfig_LoadProjectConfig_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	// When kaula.json doesn't exist, loadProjectConfig should return nil error
	err := loadProjectConfig(cfg)
	if err != nil {
		t.Fatalf("loadProjectConfig when file not found: %v", err)
	}
}

func TestConfig_NormalizePaths_AllEmpty(t *testing.T) {
	cfg := &Config{}
	normalizePaths(cfg)
	// Should not panic with empty fields
}