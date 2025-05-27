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

//go:build ignore_vet

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"google.golang.org/genai"
)

var (
	model = flag.String("model", genai.GeminiNativeAudioFlashModel, "the model name for native audio dialog")
	voiceID = flag.String("voice", "en-US-Standard-A", "the voice ID to use for audio responses")
	language = flag.String("language", "en-US", "the language for responses")
	affective = flag.Bool("affective", true, "enable affective dialog features")
)

func run(ctx context.Context) {
	// Create client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		// Set appropriate config parameters
		APIKey: os.Getenv("GOOGLE_API_KEY"),
		HTTPOptions: genai.HTTPOptions{
			APIVersion: "v1alpha", // Make sure to use the correct API version
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set up the live connection config
	liveConfig := &genai.LiveConnectConfig{
		// Set response modalities to include audio
		ResponseModalities: []genai.Modality{genai.ModalityAudio, genai.ModalityText},
		// Set temperature for response generation
		Temperature: genai.Ptr[float32](0.7),
		// Add any system instructions
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "You are a helpful assistant. Keep responses conversational and natural."},
			},
		},
	}

	// Set up native audio configuration
	audioConfig := &genai.NativeAudioConfig{
		VoiceID:              *voiceID,
		ResponseLanguage:     *language,
		EnableAffectiveDialog: *affective,
		ProactiveAudio: &genai.ProactiveAudioConfig{
			ProactivityLevel:   0.5, // Moderate proactivity
			AllowInterruptions: true,
		},
	}

	fmt.Println("Connecting to native audio dialog model:", *model)
	
	// Establish a native audio dialog session
	session, err := client.Live.ConnectNativeAudio(ctx, *model, liveConfig, audioConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer session.Close()

	fmt.Println("Connected successfully!")
	
	// Set up a goroutine to receive server messages
	go func() {
		for {
			msg, err := session.Receive()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}
			
			// Handle server content
			if msg.ServerContent != nil && msg.ServerContent.ModelTurn != nil {
				for _, part := range msg.ServerContent.ModelTurn.Parts {
					// Handle text content
					if part.Text != "" {
						fmt.Printf("Assistant (text): %s\n", part.Text)
					}
					
					// Handle audio content
					if part.InlineData != nil && part.InlineData.MIMEType == "audio/wav" {
						fmt.Printf("Assistant (audio): [%d bytes of audio data received]\n", len(part.InlineData.Data))
						// In a real application, you would save or play this audio
					}
				}
				
				if msg.ServerContent.TurnComplete {
					fmt.Println("Assistant turn complete")
				}
			}
		}
	}()

	// Send a simple greeting
	err = session.SendClientContent(genai.LiveClientContentInput{
		Turns: []*genai.Content{
			{
				Parts: []*genai.Part{
					{Text: "Hello! How are you today?"},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to send content: %v", err)
	}

	// Wait for Ctrl+C to terminate
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("\nShutting down...")
}

func main() {
	ctx := context.Background()
	flag.Parse()
	run(ctx)
}