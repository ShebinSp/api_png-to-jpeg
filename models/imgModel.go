package models

type Image struct {
	ID        uint   `gorm:"primaryKey"`
	UserName  string `gorm:"not null"`
	Image      string `gorm:"not null"`
	IsVisible bool  
}
