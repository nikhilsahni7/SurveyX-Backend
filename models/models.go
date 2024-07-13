package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Name     string
	GoogleID string `gorm:"uniqueIndex"`
	Picture  string
	Surveys  []Survey
}

type Survey struct {
	gorm.Model
	UserID        uint
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
}

type SurveyLink struct {
	gorm.Model
	SurveyID uint
	Link     string `gorm:"uniqueIndex"`
	IsActive bool
}
