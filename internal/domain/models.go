package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string `gorm:"not null"`
	Email      string `gorm:"uniqueIndex;not null"`
	AvatarURL  *string
	Provider   string `gorm:"not null"`
	ProviderID string `gorm:"not null"`
}

type Organization struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Description    *string
	WhatsappNumber string `gorm:"not null"`
	Prefix         string `gorm:"not null;default:'TSK'"`
	OwnerID        uint   `gorm:"not null;index"`
	Owner          User   `gorm:"foreignKey:OwnerID"`
}

type TeamMember struct {
	gorm.Model
	UserID         uint `gorm:"not null;index"`
	User           User
	OrganizationID uint `gorm:"not null;index"`
	Organization   Organization
	Role           string `gorm:"not null;default:'user'"`
}

type Project struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Description    *string
	StartDate      *time.Time
	OrganizationID uint `gorm:"not null;index"`
	Organization   Organization
	CreatedByID    uint `gorm:"not null"`
	CreatedBy      User `gorm:"foreignKey:CreatedByID"`
}

type State struct {
	gorm.Model
	Name      string `gorm:"not null"`
	Color     string `gorm:"not null;default:'#6B7280'"`
	Position  uint   `gorm:"not null;default:0"`
	IsDone    bool   `gorm:"not null;default:false"`
	ProjectID uint   `gorm:"not null;index"`
	Project   Project
}

type Sprint struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Goal        *string
	ProjectID   uint `gorm:"not null;index"`
	Project     Project
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	CompletedAt *time.Time
}

type BurndownSnapshot struct {
	gorm.Model
	SprintID       uint `gorm:"not null;index"`
	Sprint         Sprint
	Date           time.Time `gorm:"not null"`
	RemainingTasks int       `gorm:"not null"`
	CompletedTasks int       `gorm:"not null"`
}

type Task struct {
	gorm.Model
	TaskNumber  int    `gorm:"not null;default:0"`
	Title       string `gorm:"not null"`
	Priority    string `gorm:"not null;default:'medium'"`
	SprintID    *uint  `gorm:"index"`
	Description *string
	StateID     uint `gorm:"not null;index"`
	State       State
	ProjectID   uint `gorm:"not null;index"`
	Project     Project
	AssignedTo  *uint `gorm:"index"`
	Assignee    *User `gorm:"foreignKey:AssignedTo"`
	CreatedByID uint  `gorm:"not null"`
	CreatedBy   User  `gorm:"foreignKey:CreatedByID"`
	StartDate   *time.Time
	DueDate     *time.Time `gorm:"index"`
	CompletedAt *time.Time `gorm:"index"`
}

type Comment struct {
	gorm.Model
	Content  string `gorm:"not null"`
	TaskID   uint   `gorm:"not null;index"`
	Task     Task
	AuthorID uint `gorm:"not null"`
	Author   User `gorm:"foreignKey:AuthorID"`
}

type Invitation struct {
	gorm.Model
	Email          string `gorm:"not null"`
	OrganizationID uint   `gorm:"not null;index"`
	Organization   Organization
	Token          string    `gorm:"uniqueIndex;not null"`
	ExpiresAt      time.Time `gorm:"not null"`
	AcceptedAt     *time.Time
}

type GitHubIntegration struct {
	gorm.Model
	OrganizationID  uint `gorm:"not null;index"`
	Organization    Organization
	RepoOwner       string `gorm:"not null"`
	RepoName        string `gorm:"not null"`
	ProjectID       uint   `gorm:"not null;index"`
	Project         Project
	WebhookSecret   string `gorm:"not null"`
	InReviewStateID uint   `gorm:"not null"`
	DoneStateID     uint   `gorm:"not null"`
}

type APIKey struct {
	gorm.Model
	Name       string `gorm:"not null"`
	KeyHash    string `gorm:"uniqueIndex;not null"`
	KeyPrefix  string `gorm:"not null"`
	UserID     uint   `gorm:"not null;index"`
	User       User
	OrgID      uint `gorm:"not null;index"`
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}

type OutboundWebhook struct {
	gorm.Model
	OrgID        uint         `gorm:"not null;index"`
	Organization Organization `gorm:"foreignKey:OrgID"`
	URL          string       `gorm:"not null"`
	Secret       string       `gorm:"not null"`
	Events       string       `gorm:"not null;default:'task.moved,task.created'"`
	Active       bool         `gorm:"not null;default:true"`
}

type NotificationConfig struct {
	gorm.Model
	OrgID      uint   `gorm:"not null;index"`
	Channel    string `gorm:"not null"`
	WebhookURL string
	Enabled    bool `gorm:"not null;default:true"`
}

type NotificationPreference struct {
	gorm.Model
	UserID  uint   `gorm:"not null;index"`
	OrgID   uint   `gorm:"not null;index"`
	Event   string `gorm:"not null"`
	Channel string `gorm:"not null"`
	Enabled bool   `gorm:"not null;default:true"`
}

type StateTransition struct {
	gorm.Model
	TaskID             uint `gorm:"not null;index"`
	Task               Task
	FromStateID        *uint
	ToStateID          uint `gorm:"not null"`
	ChangedByID        uint `gorm:"not null"`
	SecondsInFromState int64
}

type Subtask struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description *string
	Done        bool `gorm:"not null;default:false"`
	TaskID      uint `gorm:"not null;index"`
	Task        Task
	CreatedByID uint `gorm:"not null"`
	CreatedBy   User `gorm:"foreignKey:CreatedByID"`
}

type SubtaskComment struct {
	gorm.Model
	Content   string `gorm:"not null"`
	SubtaskID uint   `gorm:"not null;index"`
	Subtask   Subtask
	AuthorID  uint `gorm:"not null"`
	Author    User `gorm:"foreignKey:AuthorID"`
}

type TimeEntry struct {
	gorm.Model
	TaskID    *uint    `gorm:"index"`
	Task      *Task    `gorm:"foreignKey:TaskID"`
	SubtaskID *uint    `gorm:"index"`
	Subtask   *Subtask `gorm:"foreignKey:SubtaskID"`
	UserID    uint     `gorm:"not null;index"`
	User      User
	Minutes   int `gorm:"not null"`
	Note      *string
	LoggedAt  time.Time `gorm:"not null;index"`
}

type ActiveTimer struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex;not null"`
	User      User
	TaskID    *uint     `gorm:"index"`
	Task      *Task     `gorm:"foreignKey:TaskID"`
	SubtaskID *uint     `gorm:"index"`
	Subtask   *Subtask  `gorm:"foreignKey:SubtaskID"`
	StartedAt time.Time `gorm:"not null"`
}
