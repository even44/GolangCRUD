package models

import (
	"gorm.io/gorm"
)

type JobApplication struct {
	gorm.Model
	UserID     uint      `json:"user_id" gorm:"index:idx_application"`
	CompanyID  uint      `json:"company_id" gorm:"index:idx_application"`
	Position   string    `json:"position"`
	Company    Company   `json:"company"`
	SearchDate CivilTime `json:"search_date"`
	Deadline   CivilTime `json:"deadline"`
	Response   bool      `json:"response"`
	Interview  bool      `json:"interview"`
	Done       bool      `json:"done"`
	Link       string    `json:"link"`
	ContactID  uint      `json:"contact_id"`
	Contact    Contact   `json:"contact"`
}
