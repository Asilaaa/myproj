package ai

import "fmt"

// Well-defined prompts for the description of the input image
func ImageDescriptionPrompt() string {
	return `You are an expert image analyst.

Analyze the image and return only valid JSON.

Use exactly this structure:
{
	"main_subject": "string",
	"objects": ["string"],
	"scene": "string",
	"notable_details": ["string"]
}

Be factual, concise, and avoid speculation.
Do not return markdown.
Do not return any text outside the JSON object.`
}

// Well-defined prompts for the user-asked question
func ImageQuestionPrompt(question string) string {
	return fmt.Sprintf(`You are answering questions about an image.

Answer the following question based only on what is visible in the image.

Question:
%s`, question)
}
