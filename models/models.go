package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex"`
	Name         string
	GoogleID     *string `gorm:"uniqueIndex"`
	Picture      string
	PasswordHash string
	Surveys      []Survey
	Teams        []Team `gorm:"many2many:user_teams;"`
}
type Team struct {
	gorm.Model
	Name    string
	OwnerID uint
	Users   []User `gorm:"many2many:user_teams;"`
	Surveys []Survey
}

type Survey struct {
	gorm.Model
	UserID        uint
	TeamID        *uint
	Title         string
	Description   string
	Questions     []Question
	ResponseLimit *int
	ReleaseDate   *time.Time
	CloseDate     *time.Time
	RedirectURL   string
	ClosedMessage string
	CustomStyles  string
	Responses     []Response
	Link          string
	IsPublished   bool
	Version       int
}

type Question struct {
	gorm.Model
	SurveyID      uint
	Text          string
	Type          string
	Options       []Option `gorm:"foreignKey:QuestionID"`
	IsRequired    bool
	Order         int
	MinValue      *int
	MaxValue      *int
	AllowMultiple bool
	MaxFileSize   *int
	Conditions    []Condition
}

type Condition struct {
	gorm.Model
	QuestionID       uint
	DependentOnID    uint
	DependentOnValue string
	Operator         string // e.g., "equals", "not equals", "greater than", etc.
}

type Option struct {
	gorm.Model
	QuestionID uint
	Text       string
	Value      string
}

type Response struct {
	gorm.Model
	SurveyID  uint
	Answers   []Answer
	IP        string
	UserAgent string
}

type Answer struct {
	gorm.Model
	ResponseID uint
	QuestionID uint
	Value      string
	Question   Question `gorm:"foreignKey:QuestionID"`
}

type SurveyLink struct {
	gorm.Model
	SurveyID uint
	Link     string `gorm:"uniqueIndex"`
	IsActive bool
}

type Webhook struct {
	gorm.Model
	UserID   uint
	SurveyID uint
	URL      string
	Events   string
	Secret   string
}
