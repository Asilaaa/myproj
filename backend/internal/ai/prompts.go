package ai

import "fmt"

//Well-defined prompts for the description of the input image
func ImageDescriptionPrompt() string {
	return `You are an expert image analyst.

Describe the image in detail.

Return:
- Main subject
- Important objects
- Scene description
- Notable details

Be factual and avoid speculation.`
}

//Well-defined prompts for the user-asked question
func ImageQuestionPrompt(question string) string {
	return fmt.Sprintf(`You are answering questions about an image.

Answer the following question based only on what is visible in the image.

Question:
%s`, question)
}
