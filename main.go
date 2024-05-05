package main

import (
	"log"

	"github.com/ShebinSp/convert_png-jpeg/controllers"
	"github.com/ShebinSp/convert_png-jpeg/initializers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Image struct {
	ID   uint   `gorm:"primaryKey"`
	Data []byte `gorm:"not null"`
}

func main() {

	// loading env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading env file: ", err)
	}
	initializers.ConnnectToDB()

	app := fiber.New()

	api := app.Group("/image")
	api.Post("/upload", controllers.UploadImage)
	api.Get("/view/id", controllers.GetImageById)
	api.Get("/view/username", controllers.GetImageByUsername)
	api.Delete("/delete", controllers.DeleteImgById)
	api.Patch("/visibility", controllers.ToggleVisibility)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
