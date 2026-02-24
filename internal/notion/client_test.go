package notion

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"empty string contains empty", "", "", true},
		{"abc contains a", "abc", "a", true},
		{"abc contains bc", "abc", "bc", true},
		{"abc contains abc", "abc", "abc", true},
		{"abc does not contain d", "abc", "d", false},
		{"abc does not contain abcd", "abc", "abcd", false},
		{"hello world contains world", "hello world", "world", true},
		{"empty string does not contain x", "", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestSearchStr(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"found at start", "hello", "he", true},
		{"found at middle", "hello", "ll", true},
		{"found at end", "hello", "lo", true},
		{"not found", "hello", "world", false},
		{"empty string", "", "a", false},
		{"substring longer", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searchStr(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("searchStr(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestGetMapString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{"key exists", map[string]any{"key": "value"}, "key", "value"},
		{"key does not exist", map[string]any{"other": "value"}, "key", ""},
		{"key is not string", map[string]any{"key": 123}, "key", ""},
		{"empty map", map[string]any{}, "key", ""},
		{"nil map", nil, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapString(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getMapString(%v, %q) = %q, want %q", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetMapBool(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected bool
	}{
		{"key exists true", map[string]any{"key": true}, "key", true},
		{"key exists false", map[string]any{"key": false}, "key", false},
		{"key does not exist", map[string]any{"other": true}, "key", false},
		{"key is not bool", map[string]any{"key": "true"}, "key", false},
		{"key is int", map[string]any{"key": 1}, "key", false},
		{"empty map", map[string]any{}, "key", false},
		{"nil map", nil, "key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapBool(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getMapBool(%v, %q) = %v, want %v", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"broken pipe", &testError{msg: "broken pipe"}, true},
		{"connection reset", &testError{msg: "connection reset by peer"}, true},
		{"EOF error", &testError{msg: "EOF"}, true},
		{"use of closed network connection", &testError{msg: "use of closed network connection"}, true},
		{"timeout error", &testError{msg: "timeout"}, false},
		{"not found", &testError{msg: "404 not found"}, false},
		{"normal error", &testError{msg: "some error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
