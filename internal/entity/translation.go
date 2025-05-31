// Package entity defines main entities for business logic (services), data base mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

import "time"

// Translation -.
type Translation struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Source      string    `json:"source"       example:"auto"`
	Destination string    `json:"destination"  example:"en"`
	Original    string    `json:"original"     example:"текст для перевода"`
	Translation string    `json:"translation"  example:"text for translation"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TranslationHistory -.
type TranslationHistory struct {
	History []Translation `json:"history"`
}
