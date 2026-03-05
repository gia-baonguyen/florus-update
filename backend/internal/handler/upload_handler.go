package handler

import (
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storageService service.StorageService
}

func NewUploadHandler(storageService service.StorageService) *UploadHandler {
	return &UploadHandler{storageService: storageService}
}

// UploadImage godoc
// @Summary Upload an image file
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file to upload"
// @Param folder formData string false "Folder to store the image (default: general)"
// @Success 200 {object} map[string]string "Returns the uploaded image URL"
// @Router /api/upload [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "No file provided")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		utils.BadRequest(c, "Invalid file type. Only images are allowed")
		return
	}

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		utils.BadRequest(c, "File too large. Maximum size is 10MB")
		return
	}

	// Get folder from form data (default: "general")
	folder := c.DefaultPostForm("folder", "general")

	// Upload to MinIO
	url, err := h.storageService.UploadFile(file, header, folder)
	if err != nil {
		utils.InternalServerError(c, "Failed to upload file: "+err.Error())
		return
	}

	utils.OK(c, "File uploaded successfully", map[string]string{
		"url": url,
	})
}

// UploadMultipleImages godoc
// @Summary Upload multiple image files
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param files formData file true "Image files to upload (multiple)"
// @Param folder formData string false "Folder to store the images (default: general)"
// @Success 200 {object} map[string][]string "Returns the uploaded image URLs"
// @Router /api/upload/multiple [post]
func (h *UploadHandler) UploadMultipleImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "Failed to parse form data")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequest(c, "No files provided")
		return
	}

	// Limit to 10 files at once
	if len(files) > 10 {
		utils.BadRequest(c, "Too many files. Maximum is 10 files at once")
		return
	}

	folder := c.DefaultPostForm("folder", "general")
	var urls []string

	for _, header := range files {
		// Validate file type
		contentType := header.Header.Get("Content-Type")
		if !isValidImageType(contentType) {
			continue // Skip invalid files
		}

		// Validate file size (max 10MB)
		if header.Size > 10*1024*1024 {
			continue // Skip files that are too large
		}

		file, err := header.Open()
		if err != nil {
			continue
		}

		url, err := h.storageService.UploadFile(file, header, folder)
		file.Close()
		if err != nil {
			continue
		}

		urls = append(urls, url)
	}

	if len(urls) == 0 {
		utils.BadRequest(c, "No valid files were uploaded")
		return
	}

	utils.OK(c, "Files uploaded successfully", map[string][]string{
		"urls": urls,
	})
}

// UploadBase64 godoc
// @Summary Upload an image from base64 data
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body uploadBase64Request true "Base64 image data"
// @Success 200 {object} map[string]string "Returns the uploaded image URL"
// @Router /api/upload/base64 [post]
func (h *UploadHandler) UploadBase64(c *gin.Context) {
	var req uploadBase64Request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if req.Image == "" {
		utils.BadRequest(c, "No image data provided")
		return
	}

	folder := req.Folder
	if folder == "" {
		folder = "general"
	}

	url, err := h.storageService.UploadBase64(req.Image, folder)
	if err != nil {
		utils.InternalServerError(c, "Failed to upload image: "+err.Error())
		return
	}

	utils.OK(c, "Image uploaded successfully", map[string]string{
		"url": url,
	})
}

// UploadMultipleBase64 godoc
// @Summary Upload multiple images from base64 data
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body uploadMultipleBase64Request true "Array of base64 image data"
// @Success 200 {object} map[string][]string "Returns the uploaded image URLs"
// @Router /api/upload/base64/multiple [post]
func (h *UploadHandler) UploadMultipleBase64(c *gin.Context) {
	var req uploadMultipleBase64Request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if len(req.Images) == 0 {
		utils.BadRequest(c, "No images provided")
		return
	}

	// Limit to 10 images at once
	if len(req.Images) > 10 {
		utils.BadRequest(c, "Too many images. Maximum is 10 at once")
		return
	}

	folder := req.Folder
	if folder == "" {
		folder = "general"
	}

	var urls []string
	for _, base64Data := range req.Images {
		if base64Data == "" {
			continue
		}

		url, err := h.storageService.UploadBase64(base64Data, folder)
		if err != nil {
			continue
		}

		urls = append(urls, url)
	}

	if len(urls) == 0 {
		utils.BadRequest(c, "No valid images were uploaded")
		return
	}

	utils.OK(c, "Images uploaded successfully", map[string][]string{
		"urls": urls,
	})
}

type uploadBase64Request struct {
	Image  string `json:"image" binding:"required"`
	Folder string `json:"folder"`
}

type uploadMultipleBase64Request struct {
	Images []string `json:"images" binding:"required"`
	Folder string   `json:"folder"`
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}
