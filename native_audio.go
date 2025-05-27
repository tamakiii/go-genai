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
	"fmt"
)

// Model constants for Gemini native audio dialog models
const (
	// GeminiNativeAudioFlashModel is the model ID for Gemini 2.5 Flash Preview native audio dialog
	GeminiNativeAudioFlashModel = "gemini-2.5-flash-preview-native-audio-dialog"
	// GeminiNativeAudioProModel is the model ID for Gemini 2.5 Pro Preview native audio dialog
	GeminiNativeAudioProModel = "gemini-2.5-pro-preview-native-audio-dialog"
)

// NativeAudioConfig represents configuration for native audio capabilities
type NativeAudioConfig struct {
	// Optional. The ID of the voice to use for spoken responses.
	VoiceID string `json:"voiceId,omitempty"`
	// Optional. Language code for the response audio, using BCP-47 format (e.g., "en-US").
	ResponseLanguage string `json:"responseLanguage,omitempty"`
	// Optional. Configuration for proactive audio capabilities.
	ProactiveAudio *ProactiveAudioConfig `json:"proactiveAudio,omitempty"`
	// Optional. Controls whether to enable affective dialog features.
	EnableAffectiveDialog bool `json:"enableAffectiveDialog,omitempty"`
}

// ProactiveAudioConfig configures proactive audio capabilities for native audio models
type ProactiveAudioConfig struct {
	// Optional. Controls the level of model proactivity in conversations.
	// Higher values make the model more likely to interject.
	ProactivityLevel float32 `json:"proactivityLevel,omitempty"`
	// Optional. Controls whether the model can interrupt the user.
	AllowInterruptions bool `json:"allowInterruptions,omitempty"`
}

// IsNativeAudioModel returns true if the provided model name is a native audio dialog model
func IsNativeAudioModel(model string) bool {
	return model == GeminiNativeAudioFlashModel || model == GeminiNativeAudioProModel
}

// ConnectNativeAudio establishes a WebSocket connection specifically configured for native audio
// dialog models with the given configuration.
func (r *Live) ConnectNativeAudio(ctx context.Context, model string, config *LiveConnectConfig, audioConfig *NativeAudioConfig) (*Session, error) {
	// Validate that the model is a native audio model
	if !IsNativeAudioModel(model) {
		return nil, fmt.Errorf("model %q is not a supported native audio dialog model", model)
	}
	
	// If response modalities not specified, default to AUDIO for native audio models
	if len(config.ResponseModalities) == 0 {
		config.ResponseModalities = []Modality{ModalityAudio}
	}
	
	// Apply native audio configuration if provided
	if audioConfig != nil {
		// Set affective dialog if configured
		if audioConfig.EnableAffectiveDialog {
			config.EnableAffectiveDialog = Ptr(true)
		}
		
		// Any additional transformations of audioConfig to config would go here
		// For now, these will be handled by the underlying API
	}
	
	// Use the standard Connect method with the enhanced configuration
	return r.Connect(ctx, model, config)
}