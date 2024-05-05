package initializers

import (
	"log"
	"os"

	"github.com/ShebinSp/convert_png-jpeg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnnectToDB() {
	var err error

	DSN := os.Getenv("dsn")

	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Panicln("Failed to connect to database")
	}

	DB.AutoMigrate(&models.Image{})
}
