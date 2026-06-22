package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

type Service struct {
	apiKey     string
	guardrails GuardrailsConfig
}

// Created everytime application runs
func NewService(apiKey string, guardrails GuardrailsConfig) *Service {
	return &Service{apiKey: apiKey, guardrails: guardrails}
}

// private client creating from OpenAI
func (s *Service) client() openai.Client {
	return openai.NewClient(option.WithAPIKey(s.apiKey))
}

func (s *Service) DescribeImage(ctx context.Context, contentType string, imageData []byte) (string, error) {
	//validation
	if strings.TrimSpace(s.apiKey) == "" {
		return "", fmt.Errorf("openai api key is required")
	}

	if len(imageData) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	if strings.TrimSpace(contentType) == "" {
		contentType = "image/jpeg"
	}

	//Preparation of image without exposing it publicly
	//1. converting bytes to base64 for JSON
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData))
	//2. creating an image input
	imageInput := responses.ResponseInputContentParamOfInputImage(responses.ResponseInputImageDetailAuto)
	//3. attaching the actual image
	imageInput.OfInputImage.ImageURL = openai.String(dataURL)

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"main_subject": map[string]any{
				"type": "string",
			},
			"objects": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"scene": map[string]any{
				"type": "string",
			},
			"notable_details": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{
			"main_subject",
			"objects",
			"scene",
			"notable_details",
		},
		"additionalProperties": false,
	}

	client := s.client()
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: shared.ResponsesModel("gpt-4.1-mini"),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText(ImageDescriptionPrompt()),
						imageInput,
					},
					responses.EasyInputMessageRoleUser,
				),
			},
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema("image_description", schema),
		},
	})
	if err != nil {
		return "", err
	}

	var structured ImageDescription
	if err := json.Unmarshal([]byte(resp.OutputText()), &structured); err != nil {
		return "", fmt.Errorf("could not parse structured image description: %w", err)
	}

	if strings.TrimSpace(structured.MainSubject) == "" && strings.TrimSpace(structured.Scene) == "" {
		return "", fmt.Errorf("structured image description is missing required content")
	}

	description := fmt.Sprintf(
		"Main subject: %s. Objects: %s. Scene: %s. Notable details: %s.",
		strings.TrimSpace(structured.MainSubject),
		strings.Join(structured.Objects, ", "),
		strings.TrimSpace(structured.Scene),
		strings.Join(structured.NotableDetails, ", "),
	)

	return applyDescriptionGuardrails(description, s.guardrails), nil
}

// this is not for this project, it is just a test
func (s *Service) RunVisionCheck(ctx context.Context) (string, error) {
	imageInput := responses.ResponseInputContentParamOfInputImage(responses.ResponseInputImageDetailAuto)
	imageInput.OfInputImage.ImageURL = openai.String("https://openai-documentation.vercel.app/images/cat_and_otter.png")

	client := s.client()
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: shared.ResponsesModel("gpt-4.1-mini"),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				responses.ResponseInputItemParamOfMessage(
					responses.ResponseInputMessageContentListParam{
						responses.ResponseInputContentParamOfInputText("What is in this image?"),
						imageInput,
					},
					responses.EasyInputMessageRoleUser,
				),
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func RunVisionCheck() error {
	service := NewService(os.Getenv("OPENAI_API_KEY"), GuardrailsConfig{Enabled: true, MaxDescriptionChars: 500})
	output, err := service.RunVisionCheck(context.Background())
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
