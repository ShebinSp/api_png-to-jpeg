package controllers

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ShebinSp/convert_png-jpeg/initializers"
	"github.com/ShebinSp/convert_png-jpeg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UploadImage(c *fiber.Ctx) error {
	var images models.Image
	db := initializers.DB

	if err := c.BodyParser(&images); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	// Parse the uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		log.Println("Error in uploading Image: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// creating a unique name for the file
	uniqueId := uuid.New()

	// removing hyphens from the uniqueId(name)
	filename := strings.Replace(uniqueId.String(), "-", "", -1)

	// spliting the file name into slice of substrings. and returns second substring
	fileExt := strings.Split(file.Filename, ".")[1]

	// creating a new file name
	image := fmt.Sprintf("%s.%s", filename, fileExt)

	// saving the file
	err = c.SaveFile(file, fmt.Sprintf("./images/%s", image))
	if err != nil {
		log.Println("Error in saving Image: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// CONVERT .PNG TO .JPEG
	// Open the PNG file to convert to jpg
	pngFile, err := os.Open("./images/" + image)
	if err != nil {
		log.Println("Error: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error"})
	}
	defer pngFile.Close()

	// Decode the png image
	pngImage, err := png.Decode(pngFile)
	if err != nil {
		log.Println("Error Decode: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Internal server error"})
	}

	// Change the extension in file name
	jpegImage := fmt.Sprintf("%s.%s", filename, "jpeg")

	// Create a new JPEG file
	jpegFile, err := os.Create("./images/" + jpegImage)
	if err != nil {
		log.Println("Error Create: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Internal server error"})
	}
	defer jpegFile.Close()

	// Encode the PNG image as JPEG
	err = jpeg.Encode(jpegFile, pngImage, nil)
	if err != nil {
		log.Println("Error Encode: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Internal server error"})
	}

	// remove the png file
	err = os.Remove("./images/" + image)
	if err != nil {
		log.Println("Error: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Internal server error"})
	}

	imageUrl := fmt.Sprintf("http://localhost:8080/images/%s", jpegImage)

	data := map[string]interface{}{
		"imageName": image,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	images.Image = imageUrl
	r := db.Create(&images)
	if r.Error != nil {
		log.Println("err on create table: ", r.Error)
		return c.JSON(fiber.Map{"status": 500, "message": "Internal server error"})

	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

func GetImageById(c *fiber.Ctx) error {
	var images models.Image
	db := initializers.DB

	imgid := c.FormValue("id")
	fmt.Println("imgid:", imgid)
	if imgid == "" {
		return c.JSON(fiber.Map{"status": http.StatusBadRequest, "message": "No id found"})
	}

	id, err := strconv.Atoi(imgid)
	if err != nil {
		fmt.Println("err: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error"})
	}

	res := db.Table("images").Select("image").Where("id = ?", id).Scan(&images.Image)
	if res.Error != nil {
		return c.JSON(fiber.Map{"status": 404, "messge": "Id not found"})
	}
	log.Println("image: ", images.Image)
	return c.JSON(fiber.Map{"status": 201, "message": "Image found", "data": images.Image})

}

func GetImageByUsername(c *fiber.Ctx) error {
	db := initializers.DB
	var allImages []string

	username := c.FormValue("username")

	res := db.Table("images").Where("user_name = ?", username).Where("is_visible = ?", true).Select("image").Scan(&allImages)
	if res.Error != nil {
		log.Println("Error: ", res.Error)
		return res.Error
	} else if res.RowsAffected == 0 {
		log.Println("username not found")
		return c.JSON(fiber.Map{"status": 404, "message": "username not found"})
	}

	return c.JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "image found for username " + username,
		"data":    allImages,
	})
}

func DeleteImgById(c *fiber.Ctx) error {
	img := ""
	db := initializers.DB
	userId := c.FormValue("id")

	id, err := strconv.Atoi(userId)
	if err != nil {
		log.Println("error while converting id")
		return c.JSON(fiber.Map{"status": 500, "message": "Server error"})
	}

	_ = db.Table("images").Where("id = ?", id).Select("image").Scan(&img)
	// formating name of the file
	prefix := "http://localhost:8080/images/"
	imgName := strings.TrimPrefix(img, prefix)

	res := db.Table("images").Where("id = ?", id).Delete("image")
	if res.Error != nil {
		log.Println("error on deleting images: ", res.Error)
		return err
	} else if res.RowsAffected == 0 {
		log.Println("id not found")
		return c.JSON(fiber.Map{"status": http.StatusNotFound, "message": "id not found"})
	}

	// deleting internally
	err = os.Remove("./images/" + imgName)
	if err != nil {
		log.Println("failed to remove the file internally: ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "failed to remove the file internally"})
	}
	return c.JSON(fiber.Map{"status": http.StatusOK, "message": "image " + imgName + " deleted successfully"})
}

func ToggleVisibility(c *fiber.Ctx) error {
	imgId := c.FormValue("id")
	db := initializers.DB
	var cStatus bool
	vStatus := ""

	id, err := strconv.Atoi(imgId)
	if err != nil {
		log.Println("Error while converting id")
	}

	res := db.Table("images").Where("id = ?", id).Select("is_visible").Find(&cStatus)
	if res.Error != nil {
		log.Println("error: ", res.Error)
		return c.JSON(fiber.Map{"status": http.StatusInternalServerError, "message": "Internal server error"})
	} else if res.RowsAffected == 0 {
		log.Println("Id does not exist")
		return c.JSON(fiber.Map{"status": 404, "message": "Id does not exist"})
	}
	if cStatus {
		cStatus = false
		vStatus = "false"
	} else {
		cStatus = true
		vStatus = "true"
	}
	_ = db.Table("images").Where("id = ?", id).Select("is_visible").Update("is_visible", cStatus)

	return c.JSON(fiber.Map{"status": 201, "message": "Image visibility changed to " + vStatus})

}
