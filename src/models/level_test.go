package models

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelStream, "STREAM"},
		{LevelInfo, "INFO"},
		{LevelLogon, "LOGON"},
		{LevelLogout, "LOGOUT"},
		{LevelTrade, "TRADE"},
		{LevelSchedule, "SCHEDULE"},
		{LevelReport, "REPORT"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{LevelCritical, "CRITICAL"},
		{LevelNotSet, "INFO"}, // Default behavior
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("%v.String() = %v, want %v", tt.level, got, tt.expected)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"DEBUG", LevelDebug},
		{"debug", LevelDebug},
		{"INFO", LevelInfo},
		{"warning", LevelWarning},
		{"ERROR", LevelError},
		{"UNKNOWN", LevelInfo}, // Default behavior
	}

	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.expected {
			t.Errorf("ParseLevel(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
