package speakers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/martinlindhe/unit"
)

type openAITTSSpeaker struct {
	baseURL string
	apiKey  string
	model   string
	voice   string
	speed   float64
}

var _ Speaker = (*openAITTSSpeaker)(nil)

// NewOpenAICompatibleTTSSpeaker creates a Speaker powered by an OpenAI-compatible API endpoint.
// The baseURL should point to the compatible API (e.g., "http://localhost:8000/v1").
// Supported models depend on the backend (e.g., "tts-1", "tts-1-hd" for OpenAI Platform).
// Supported voices depend on the model (e.g., "alloy", "echo", "fable", "onyx", "nova", "shimmer").
func NewOpenAICompatibleTTSSpeaker(apiKey, baseURL, model, voice string, playbackSpeed float64) Speaker {
	return &openAITTSSpeaker{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		voice:   voice,
		speed:   playbackSpeed,
	}
}

type speechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed,omitempty"`
}

// SayContext implements [Speaker.SayContext].
func (s *openAITTSSpeaker) SayContext(ctx context.Context, text string) ([]float32, error) {
	reqBody := speechRequest{
		Model:          s.model,
		Input:          text,
		Voice:          s.voice,
		ResponseFormat: "pcm",
		Speed:          s.speed,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal speech request: %w", err)
	}

	url := s.baseURL + "/audio/speech"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create speech request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("speech API returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read synthesized audio: %w", err)
	}

	// OpenAI TTS returns 24kHz mono PCM. Downsample to 16kHz.
	downsampled, err := downsample(data, 24000*unit.Hertz)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample synthesized audio: %w", err)
	}

	f32le := pcm.S16LEBytesToF32LE(downsampled)
	return f32le, nil
}

// Say implements [Speaker.Say].
func (s *openAITTSSpeaker) Say(text string) ([]float32, error) {
	return s.SayContext(context.Background(), text)
}
