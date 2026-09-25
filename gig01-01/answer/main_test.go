// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	indexHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello, GIG!") {
		t.Fatalf("expected response body to contain 'Hello, GIG!', got %q", rec.Body.String())
	}
}

func TestAuthorizeUser(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		targetID string
		expected bool
	}{
		{
			name:     "No header passes in tutorial context",
			header:   "",
			targetID: "user123",
			expected: true,
		},
		{
			name:     "Matching user header passes",
			header:   "user123",
			targetID: "user123",
			expected: true,
		},
		{
			name:     "Mismatched user header is rejected (IDOR prevention)",
			header:   "attacker",
			targetID: "victim",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/firestore/victim", nil)
			if tc.header != "" {
				req.Header.Set("X-User-ID", tc.header)
			}
			result := authorizeUser(req, tc.targetID)
			if result != tc.expected {
				t.Fatalf("authorizeUser() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestGetUserBody(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		payload := `{"email":"test@example.com","name":"Test User"}`
		req := httptest.NewRequest(http.MethodPost, "/firestore", strings.NewReader(payload))
		u, err := getUserBody(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "test@example.com" || u.Name != "Test User" {
			t.Fatalf("unexpected user parsed: %+v", u)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/firestore", strings.NewReader(`{invalid}`))
		_, err := getUserBody(req)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("Empty Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/firestore", bytes.NewReader(nil))
		_, err := getUserBody(req)
		if err == nil {
			t.Fatal("expected error for empty body, got nil")
		}
	})
}
