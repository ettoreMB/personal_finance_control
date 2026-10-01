package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Classification struct {
	Type     string
	Category string
}

type Classifier interface {
	Classify(ctx context.Context, text string, categories []string) (Classification, error)
}

type Sender interface {
	SendText(ctx context.Context, number, text string) (messageID string, err error)
}

type EvolutionSender struct {
	baseURL    string
	apiKey     string
	instanceID string
	client     *http.Client
}

func NewEvolutionSender(baseURL, apiKey, instanceID string) *EvolutionSender {
	return &EvolutionSender{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		instanceID: instanceID,
		client:     &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *EvolutionSender) SendText(ctx context.Context, number, text string) (string, error) {
	body, err := json.Marshal(map[string]string{"number": number, "text": text})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/send/text", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("instanceId", s.instanceID)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("evolution send status %d", resp.StatusCode)
	}

	var parsed struct {
		MessageID string `json:"messageId"`
		Data      struct {
			Info struct {
				ID string `json:"ID"`
			} `json:"Info"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return "", err
	}
	if parsed.Data.Info.ID != "" {
		return parsed.Data.Info.ID, nil
	}
	return parsed.MessageID, nil
}

type JevClassifier struct {
	apiKey string
	model  string
	client *http.Client
}

func NewJevClassifier(apiKey, model string) *JevClassifier {
	if model == "" {
		model = "jev-latest"
	}
	return &JevClassifier{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *JevClassifier) Classify(ctx context.Context, text string, categories []string) (Classification, error) {
	categoryCriteria := map[string]string{
		"nenhuma": "Nenhuma categoria da lista serve.",
	}
	for _, name := range categories {
		categoryCriteria[name] = "Categoria chamada " + name + "."
	}
	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"state": text,
		"questions": map[string]any{
			"tipo": map[string]any{
				"type":         "choice",
				"instructions": "A mensagem descreve entrada de dinheiro, saída de dinheiro, ou não dá para saber?",
				"criteria": map[string]string{
					"ganho":  "Entrada de dinheiro.",
					"gasto":  "Saída de dinheiro.",
					"nenhum": "Não dá para saber se é ganho ou gasto.",
				},
			},
			"categoria": map[string]any{
				"type":         "choice",
				"instructions": "Qual categoria da lista combina com a mensagem?",
				"criteria":     categoryCriteria,
			},
		},
	})
	if err != nil {
		return Classification{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.typesafe.ai/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return Classification{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Classification{}, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Classification{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Classification{}, fmt.Errorf("jev status %d", resp.StatusCode)
	}

	var parsed struct {
		Answers map[string]struct {
			Choice string `json:"choice"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return Classification{}, err
	}
	out := Classification{}
	switch parsed.Answers["tipo"].Choice {
	case "ganho":
		out.Type = "income"
	case "gasto":
		out.Type = "expense"
	}
	category := parsed.Answers["categoria"].Choice
	if category != "" && category != "nenhuma" {
		out.Category = category
	}
	return out, nil
}
