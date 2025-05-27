// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package genai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestIsNativeAudioModel(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		wantResult bool
	}{
		{
			name:       "flash native audio model",
			model:      GeminiNativeAudioFlashModel,
			wantResult: true,
		},
		{
			name:       "pro native audio model",
			model:      GeminiNativeAudioProModel,
			wantResult: true,
		},
		{
			name:       "standard model",
			model:      "gemini-2.0-flash",
			wantResult: false,
		},
		{
			name:       "empty string",
			model:      "",
			wantResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsNativeAudioModel(test.model)
			if result != test.wantResult {
				t.Errorf("IsNativeAudioModel(%q) = %v, want %v", test.model, result, test.wantResult)
			}
		})
	}
}

func TestLiveConnectNativeAudio(t *testing.T) {
	// Create a test server that simulates a WebSocket server
	var upgrader = websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		// Echo back the first message
		_, message, err := c.ReadMessage()
		if err != nil {
			return
		}
		err = c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	// Convert http to ws
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create a client with the test server URL
	httpOpts := HTTPOptions{
		BaseURL:    wsURL,
		APIVersion: "v1alpha",
	}
	clientCfg := ClientConfig{
		APIKey:      "test-api-key",
		Backend:     BackendGeminiAPI,
		HTTPOptions: httpOpts,
	}
	
	clientCfgPtr := &clientCfg
	client := &Client{
		clientConfig: clientCfg,
		Live: &Live{
			apiClient: &apiClient{
				clientConfig: clientCfgPtr,
			},
		},
	}

	tests := []struct {
		name          string
		model         string
		config        *LiveConnectConfig
		audioConfig   *NativeAudioConfig
		wantErr       bool
		wantErrPrefix string
	}{
		{
			name:  "valid native audio flash model",
			model: GeminiNativeAudioFlashModel,
			config: &LiveConnectConfig{
				Temperature: Ptr[float32](0.7),
			},
			audioConfig: &NativeAudioConfig{
				VoiceID:              "en-US-Standard-A",
				ResponseLanguage:     "en-US",
				EnableAffectiveDialog: true,
			},
			wantErr: false,
		},
		{
			name:  "valid native audio pro model",
			model: GeminiNativeAudioProModel,
			config: &LiveConnectConfig{
				Temperature: Ptr[float32](0.5),
			},
			audioConfig: &NativeAudioConfig{
				VoiceID:              "en-US-Standard-B",
				ResponseLanguage:     "en-US",
				EnableAffectiveDialog: true,
				ProactiveAudio: &ProactiveAudioConfig{
					ProactivityLevel:   0.8,
					AllowInterruptions: true,
				},
			},
			wantErr: false,
		},
		{
			name:  "invalid model",
			model: "gemini-2.0-flash",
			config: &LiveConnectConfig{
				Temperature: Ptr[float32](0.5),
			},
			audioConfig: &NativeAudioConfig{
				VoiceID:          "en-US-Standard-A",
				ResponseLanguage: "en-US",
			},
			wantErr:       true,
			wantErrPrefix: "model \"gemini-2.0-flash\" is not a supported native audio dialog model",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			session, err := client.Live.ConnectNativeAudio(ctx, test.model, test.config, test.audioConfig)
			
			// Skip actual WebSocket connection tests since we can't fully simulate that in unit tests
			if err != nil {
				if !test.wantErr {
					t.Errorf("ConnectNativeAudio() error = %v, wantErr %v", err, test.wantErr)
				} else if !strings.HasPrefix(err.Error(), test.wantErrPrefix) {
					t.Errorf("ConnectNativeAudio() error = %v, want error prefix %v", err, test.wantErrPrefix)
				}
				return
			}
			
			if test.wantErr {
				t.Errorf("ConnectNativeAudio() error = nil, wantErr %v", test.wantErr)
				return
			}
			
			// Verify session is created (we won't be able to test an actual connection in unit tests)
			if session == nil {
				t.Errorf("ConnectNativeAudio() returned nil session")
			}
			
			// Clean up
			if session != nil {
				session.Close()
			}
		})
	}
}

// Remove this duplicate function since it exists in common.go
// Helper function to create a pointer to a float32
/*func Ptr[T any](v T) *T {
	return &v
}*/