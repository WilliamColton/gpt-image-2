package service

type User struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Role      string `json:"role"`
	Status    string `json:"-"`
	ApikeyCipher string `json:"-"`
}

type AuthUser struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Role       string `json:"role"`
	ImageCount int    `json:"imageCount"`
}

type AppConfig struct {
	BaseURL          string `json:"baseUrl"`
	CodexCLI         bool   `json:"codexCli"`
	APIMode          string `json:"apiMode"`
	Model            string `json:"model"`
	Timeout          int    `json:"timeout"`
	OpenAIConfigured bool   `json:"openAIConfigured"`
}

type Image struct {
	ID        string `json:"id"`
	UserID    string `json:"userId,omitempty"`
	FilePath  string `json:"filePath,omitempty"`
	Mime      string `json:"mime"`
	Size      int64  `json:"size"`
	Sha256    string `json:"sha256,omitempty"`
	Source    string `json:"source"`
	CreatedAt int64  `json:"createdAt"`
}

type TaskParams struct {
	Size              string  `json:"size"`
	Quality           string  `json:"quality"`
	OutputFormat      string  `json:"output_format"`
	OutputCompression *int    `json:"output_compression"`
	Moderation        string  `json:"moderation"`
	N                 int     `json:"n"`
}

type TaskRecord struct {
	ID                   string              `json:"id"`
	Prompt               string              `json:"prompt"`
	Params               interface{}         `json:"params"`
	ActualParams         interface{}         `json:"actualParams,omitempty"`
	ActualParamsByImage  interface{}         `json:"actualParamsByImage,omitempty"`
	RevisedPromptByImage interface{}        `json:"revisedPromptByImage,omitempty"`
	InputImageIDs        []string            `json:"inputImageIds"`
	MaskTargetImageID    *string             `json:"maskTargetImageId"`
	MaskImageID          *string             `json:"maskImageId"`
	OutputImages         []string            `json:"outputImages"`
	Status               string              `json:"status"`
	Error                *string             `json:"error"`
	IsFavorite           bool                `json:"isFavorite"`
	CreatedAt            int64               `json:"createdAt"`
	FinishedAt           *int64              `json:"finishedAt"`
	Elapsed              *int64              `json:"elapsed"`
	ApiMode              string              `json:"apiMode,omitempty"`
	CodexCli             bool                `json:"codexCli,omitempty"`
}
