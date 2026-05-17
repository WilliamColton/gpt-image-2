package database

type User struct {
	ID          string `gorm:"primaryKey;type:text"`
	Label       string `gorm:"type:text;not null"`
	Role        string `gorm:"type:text;not null"`
	Status      string `gorm:"type:text;not null;default:active"`
	CreatedAt   int64  `gorm:"not null"`
	LastLoginAt *int64
	Quota       int `gorm:"not null;default:0"`
	UsedCount   int `gorm:"not null;default:0"`
}

func (User) TableName() string { return "users" }

type RedemptionCode struct {
	ID        string  `gorm:"primaryKey;type:text"`
	Code      string  `gorm:"type:text;uniqueIndex;not null"`
	Quota     int     `gorm:"not null"`
	UsedBy    *string `gorm:"type:text"`
	UsedAt    *int64
	CreatedAt int64 `gorm:"not null"`
}

func (RedemptionCode) TableName() string { return "redemption_codes" }

type Image struct {
	ID        string `gorm:"primaryKey;type:text"`
	UserID    string `gorm:"type:text;not null;index"`
	FilePath  string `gorm:"type:text;not null"`
	Mime      string `gorm:"type:text;not null"`
	Size      int64  `gorm:"not null"`
	Sha256    string `gorm:"type:text;not null"`
	Source    string `gorm:"type:text;not null"`
	CreatedAt int64  `gorm:"not null"`
}

func (Image) TableName() string { return "images" }

type Task struct {
	ID                       string  `gorm:"primaryKey;type:text"`
	UserID                   string  `gorm:"type:text;not null;index"`
	Prompt                   string  `gorm:"type:text;not null"`
	ParamsJSON               string  `gorm:"type:text;not null;column:params_json"`
	ActualParamsJSON         *string `gorm:"type:text;column:actual_params_json"`
	ActualParamsByImageJSON  *string `gorm:"type:text;column:actual_params_by_image_json"`
	RevisedPromptByImageJSON *string `gorm:"type:text;column:revised_prompt_by_image_json"`
	InputImageIDsJSON        string  `gorm:"type:text;not null;column:input_image_ids_json"`
	MaskTargetImageID        *string `gorm:"type:text"`
	MaskImageID              *string `gorm:"type:text"`
	OutputImageIDsJSON       string  `gorm:"type:text;not null;column:output_image_ids_json"`
	Status                   string  `gorm:"type:text;not null"`
	Error                    *string `gorm:"type:text"`
	IsFavorite               int     `gorm:"not null;default:0"`
	CreatedAt                int64   `gorm:"not null"`
	FinishedAt               *int64
	Elapsed                  *int64
	ApiMode                  *string `gorm:"type:text"`
	CodexCli                 int     `gorm:"not null;default:0"`
}

func (Task) TableName() string { return "tasks" }

type Announcement struct {
	ID        string `gorm:"primaryKey;type:text"`
	Content   string `gorm:"type:text;not null"`
	Enabled   int    `gorm:"not null;default:0"`
	UpdatedAt int64  `gorm:"not null"`
}

func (Announcement) TableName() string { return "announcements" }
