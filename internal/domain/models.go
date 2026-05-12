package domain

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Name       string  `gorm:"not null"`
	Email      string  `gorm:"uniqueIndex;not null"`
	AvatarURL  *string // pointer = nullable
	Provider   string  `gorm:"not null"`
	ProviderID string  `gorm:"not null"`
}

type Organization struct {
	gorm.Model
	Name           string  `gorm:"not null"`
	Description    *string // pointer = nullable
	WhatsappNumber string  `gorm:"not null"`
	OwnerID        uint    `gorm:"not null"`
	Owner          User    `gorm:"foreignKey:OwnerID"`
}

type TeamMember struct {
	gorm.Model
	UserID         uint         `gorm:"not null"`
	User           User         // relationship
	OrganizationID uint         `gorm:"not null"`
	Organization   Organization // relationship
	Role           string       `gorm:"not null;default:'user'"`
}

type Project struct {
	gorm.Model
	Name           string       `gorm:"not null"`
	Description    *string      // pointer = nullable
	StartDate      *time.Time   // pointer = nullable
	OrganizationID uint         `gorm:"not null"`
	Organization   Organization // relationship
	CreatedByID    uint         `gorm:"not null"`
	CreatedBy      User         `gorm:"foreignKey:CreatedByID"`
}

type State struct {
	gorm.Model
	Name      string  `gorm:"not null"`
	Color     string  `gorm:"not null;default:'#6B7280'"` // hex color
	Position  uint    `gorm:"not null;default:0"`         // order in the board
	IsDone    bool    `gorm:"not null;default:false"`     // triggers WhatsApp notification
	ProjectID uint    `gorm:"not null"`
	Project   Project // relationship
}

type Task struct {
	gorm.Model
	Title       string     `gorm:"not null"`
	Description *string    // pointer = nullable
	StateID     uint       `gorm:"not null"`
	State       State      // relationship
	ProjectID   uint       `gorm:"not null"`
	Project     Project    // relationship
	AssignedTo  *uint      // pointer = nullable
	Assignee    *User      `gorm:"foreignKey:AssignedTo"`
	CreatedByID uint       `gorm:"not null"`
	CreatedBy   User       `gorm:"foreignKey:CreatedByID"`
	StartDate   *time.Time // pointer = nullable
	DueDate     *time.Time // pointer = nullable
	CompletedAt *time.Time // pointer = nullable
}

type Comment struct {
	gorm.Model
	Content  string `gorm:"not null"`
	TaskID   uint   `gorm:"not null"`
	Task     Task   // relationship
	AuthorID uint   `gorm:"not null"`
	Author   User   `gorm:"foreignKey:AuthorID"`
}

type Invitation struct {
	gorm.Model
	Email          string       `gorm:"not null"`
	OrganizationID uint         `gorm:"not null"`
	Organization   Organization
	Token          string       `gorm:"uniqueIndex;not null"`
	ExpiresAt      time.Time    `gorm:"not null"`
	AcceptedAt     *time.Time
}
