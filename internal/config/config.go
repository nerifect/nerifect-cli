package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// RepoConfig represents a tracked repository with its scan settings.
type RepoConfig struct {
	Name     string  `yaml:"name" json:"name"`
	Path     string  `yaml:"path,omitempty" json:"path,omitempty"`
	URL      string  `yaml:"url,omitempty" json:"url,omitempty"`
	Branch   string  `yaml:"branch,omitempty" json:"branch,omitempty"`
	ScanType string  `yaml:"scan_type,omitempty" json:"scan_type,omitempty"`
	Policies []int64 `yaml:"policies,omitempty" json:"policies,omitempty"`
}

type Config struct {
	LLMProvider     string       `yaml:"llm_provider" json:"llm_provider"`
	GeminiAPIKey    string       `yaml:"gemini_api_key" json:"-"`
	OpenAIAPIKey    string       `yaml:"openai_api_key" json:"-"`
	AnthropicAPIKey string       `yaml:"anthropic_api_key" json:"-"`
	GithubToken     string       `yaml:"github_token" json:"-"`
	DefaultModel    string       `yaml:"default_model" json:"default_model"`
	OutputFormat    string       `yaml:"output_format" json:"output_format"`
	DataDir         string       `yaml:"data_dir" json:"data_dir"`
	DatabasePath    string       `yaml:"database_path" json:"database_path"`
	MaxFilesPerScan int          `yaml:"max_files_per_scan" json:"max_files_per_scan"`
	MaxFileSizeKB   int          `yaml:"max_file_size_kb" json:"max_file_size_kb"`
	Repos           []RepoConfig `yaml:"repos,omitempty" json:"repos,omitempty"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".nerifect")
	return &Config{
		LLMProvider:     "gemini",
		DefaultModel:    "gemini-2.0-flash",
		OutputFormat:    "table",
		DataDir:         dataDir,
		DatabasePath:    filepath.Join(dataDir, "nerifect.db"),
		MaxFilesPerScan: 800,
		MaxFileSizeKB:   80,
	}
}

func configPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".nerifect.yaml")
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	// Env var overrides
	if v := os.Getenv("NERIFECT_PROVIDER"); v != "" {
		cfg.LLMProvider = v
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.GeminiAPIKey = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAIAPIKey = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		cfg.GithubToken = v
	}
	if v := os.Getenv("NERIFECT_MODEL"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("NERIFECT_OUTPUT"); v != "" {
		cfg.OutputFormat = v
	}
	if v := os.Getenv("NERIFECT_DATA_DIR"); v != "" {
		cfg.DataDir = v
		cfg.DatabasePath = filepath.Join(v, "nerifect.db")
	}

	// Ensure data directory
	if cfg.DataDir != "" {
		os.MkdirAll(cfg.DataDir, 0700)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	// Ensure data directory
	if c.DataDir != "" {
		os.MkdirAll(c.DataDir, 0700)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := configPath()
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func (c *Config) Set(key, value string) error {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == key {
			f := v.Field(i)
			if f.Kind() == reflect.String {
				f.SetString(value)
				return c.Save()
			}
			if f.Kind() == reflect.Int {
				var n int
				fmt.Sscanf(value, "%d", &n)
				f.SetInt(int64(n))
				return c.Save()
			}
			return fmt.Errorf("unsupported field type for %q", key)
		}
	}
	return fmt.Errorf("unknown config key %q", key)
}

func (c *Config) Get(key string) (string, error) {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == key {
			f := v.Field(i)
			if f.Kind() == reflect.String {
				return f.String(), nil
			}
			return fmt.Sprintf("%v", f.Interface()), nil
		}
	}
	return "", fmt.Errorf("unknown config key %q", key)
}

func (c *Config) Validate() error {
	if c.ActiveAPIKey() == "" {
		switch c.LLMProvider {
		case "openai":
			return fmt.Errorf("OPENAI_API_KEY is required for OpenAI provider (set via config or env var)")
		case "anthropic":
			return fmt.Errorf("ANTHROPIC_API_KEY is required for Anthropic provider (set via config or env var)")
		default:
			return fmt.Errorf("GEMINI_API_KEY is required for Gemini provider (set via config or env var)")
		}
	}
	return nil
}

// ActiveAPIKey returns the API key for the currently configured LLM provider.
func (c *Config) ActiveAPIKey() string {
	switch c.LLMProvider {
	case "openai":
		return c.OpenAIAPIKey
	case "anthropic":
		return c.AnthropicAPIKey
	default:
		return c.GeminiAPIKey
	}
}

func ValidConfigKeys() []string {
	t := reflect.TypeOf(Config{})
	keys := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("yaml")
		if tag != "" && tag != "-" {
			keys = append(keys, tag)
		}
	}
	return keys
}

func ValidConfigKeysStr() string {
	return strings.Join(ValidConfigKeys(), ", ")
}

// FindRepo looks up a repo config by name, path, or URL.
func (c *Config) FindRepo(nameOrTarget string) *RepoConfig {
	for i := range c.Repos {
		r := &c.Repos[i]
		if strings.EqualFold(r.Name, nameOrTarget) {
			return r
		}
		if r.Path != "" && r.Path == nameOrTarget {
			return r
		}
		if r.URL != "" && r.URL == nameOrTarget {
			return r
		}
	}
	return nil
}

// AddRepo appends a repo to the config and saves. Returns an error if a repo with the same name exists.
func (c *Config) AddRepo(repo RepoConfig) error {
	for _, r := range c.Repos {
		if strings.EqualFold(r.Name, repo.Name) {
			return fmt.Errorf("repo %q already exists", repo.Name)
		}
	}
	c.Repos = append(c.Repos, repo)
	return c.Save()
}

// UpdateRepo updates an existing repo by name and saves.
func (c *Config) UpdateRepo(name string, update func(*RepoConfig)) error {
	for i := range c.Repos {
		if strings.EqualFold(c.Repos[i].Name, name) {
			update(&c.Repos[i])
			return c.Save()
		}
	}
	return fmt.Errorf("repo %q not found", name)
}

// RemoveRepo removes a repo by name and saves.
func (c *Config) RemoveRepo(name string) error {
	for i, r := range c.Repos {
		if strings.EqualFold(r.Name, name) {
			c.Repos = append(c.Repos[:i], c.Repos[i+1:]...)
			return c.Save()
		}
	}
	return fmt.Errorf("repo %q not found", name)
}

// ListRepos returns all configured repos.
func (c *Config) ListRepos() []RepoConfig {
	return c.Repos
}
