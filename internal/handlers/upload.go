package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const UploadDir = "./uploads"

func ensureUploadDir() error {
	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		return os.MkdirAll(UploadDir, 0755)
	}
	return nil
}

func generateFilename(ext string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%s_%d%s", hex.EncodeToString(b), time.Now().UnixNano(), ext)
}

// UploadFile godoc
//
// @Summary      Upload file
// @Description  Upload an image or signature file
// @Tags         Upload
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "File to upload"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /upload [post]
func UploadFile(c *gin.Context) {
	if err := ensureUploadDir(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required field 'file' is missing. Please upload a file using a multipart/form-data request with field name 'file'."})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB. Received: " + fmt.Sprintf("%d bytes", file.Size)})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpg, jpeg, png, gif, webp, svg files are allowed"})
		return
	}

	filename := generateFilename(ext)
	savePath := filepath.Join(UploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	url := fmt.Sprintf("/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url, "filename": filename})
}
