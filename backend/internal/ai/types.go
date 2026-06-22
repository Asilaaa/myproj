package ai

type ImageDescription struct {
	MainSubject    string   `json:"main_subject"`
	Objects        []string `json:"objects"`
	Scene          string   `json:"scene"`
	NotableDetails []string `json:"notable_details"`
}
