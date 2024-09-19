package azure_models

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-models/internal/sse"
)

type Client struct {
	client *http.Client
	token  string
}

const (
	prodInferenceURL = "https://models.inference.ai.azure.com/chat/completions"
	prodModelsURL    = "https://api.catalog.azureml.ms/asset-gallery/v1.0/models"
)

func NewClient(authToken string) *Client {
	httpClient, _ := api.DefaultHTTPClient()
	return &Client{
		client: httpClient,
		token:  authToken,
	}
}

func (c *Client) GetChatCompletionStream(req ChatCompletionOptions) (*ChatCompletionResponse, error) {
	req.Stream = true

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(bodyBytes)

	httpReq, err := http.NewRequest("POST", prodInferenceURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// If we aren't going to return an SSE stream, then ensure the response body is closed.
		defer resp.Body.Close()
		return nil, c.handleHTTPError(resp)
	}

	var chatCompletionResponse ChatCompletionResponse
	chatCompletionResponse.Reader = sse.NewEventReader[ChatCompletion](resp.Body)

	return &chatCompletionResponse, nil
}

func (c *Client) ListModels() ([]*ModelSummary, error) {
	body := bytes.NewReader([]byte(`
		{
			"filters": [
				{ "field": "freePlayground", "values": ["true"], "operator": "eq"},
				{ "field": "labels", "values": ["latest"], "operator": "eq"}
			],
			"order": [
				{ "field": "displayName", "direction": "asc" }
			]
		}
	`))

	httpReq, err := http.NewRequest("POST", prodModelsURL, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var searchResponse modelCatalogSearchResponse
	err = decoder.Decode(&searchResponse)
	if err != nil {
		return nil, err
	}

	models := make([]*ModelSummary, 0, len(searchResponse.Summaries))
	for _, summary := range searchResponse.Summaries {
		inferenceTask := ""
		if len(summary.InferenceTasks) > 0 {
			inferenceTask = summary.InferenceTasks[0]
		}

		models = append(models, &ModelSummary{
			ID:           summary.AssetID,
			Name:         summary.Name,
			FriendlyName: summary.DisplayName,
			Task:         inferenceTask,
			Publisher:    summary.Publisher,
			Summary:      summary.Summary,
		})
	}

	return models, nil
}

func (c *Client) handleHTTPError(resp *http.Response) error {

	sb := strings.Builder{}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		sb.WriteString("unauthorized")

	case http.StatusBadRequest:
		sb.WriteString("bad request")

	default:
		sb.WriteString("unexpected response from the server: " + resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) > 0 {
		sb.WriteString("\n")
		sb.Write(body)
		sb.WriteString("\n")
	}

	return errors.New(sb.String())
}
