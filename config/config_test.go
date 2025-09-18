package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.CacheType != CacheTypeMemory {
		t.Errorf("Expected CacheType to be 'memory', got %s", cfg.CacheType)
	}
	
	if cfg.CacheTTL != 10*time.Minute {
		t.Errorf("Expected CacheTTL to be 10 minutes, got %v", cfg.CacheTTL)
	}
	
	if cfg.SyncInterval != 5*time.Minute {
		t.Errorf("Expected SyncInterval to be 5 minutes, got %v", cfg.SyncInterval)
	}
	
	if cfg.MaxCacheSize != 1000 {
		t.Errorf("Expected MaxCacheSize to be 1000, got %d", cfg.MaxCacheSize)
	}
}

func TestCacheType(t *testing.T) {
	tests := []struct {
		name     string
		cacheType CacheType
		expected  string
	}{
		{"memory cache", CacheTypeMemory, "memory"},
		{"redis cache", CacheTypeRedis, "redis"},
		{"no cache", CacheTypeNone, "none"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.cacheType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.cacheType))
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "valid memory cache config",
			config:    &Config{DSN: "test", CacheType: CacheTypeMemory, MaxCacheSize: 100},
			expectErr: false,
		},
		{
			name:      "valid redis cache config", 
			config:    &Config{DSN: "test", CacheType: CacheTypeRedis, RedisAddr: "localhost:6379"},
			expectErr: false,
		},
		{
			name:      "valid no cache config",
			config:    &Config{DSN: "test", CacheType: CacheTypeNone},
			expectErr: false,
		},
		{
			name:      "empty DSN",
			config:    &Config{CacheType: CacheTypeMemory},
			expectErr: true,
		},
		{
			name:      "invalid cache type",
			config:    &Config{DSN: "test", CacheType: "invalid"},
			expectErr: true,
		},
		{
			name:      "redis cache without address",
			config:    &Config{DSN: "test", CacheType: CacheTypeRedis},
			expectErr: true,
		},
		{
			name:      "memory cache with zero size",
			config:    &Config{DSN: "test", CacheType: CacheTypeMemory, MaxCacheSize: 0},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Expected error message to be 'test error', got %s", err.Error())
	}
}