package router

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"cdn/api/util"

	"github.com/gin-gonic/gin"
)

// Directory for uploaded images.
var folder = "./uploads/images"

// FileUploads maps an upload token to file metadata.
var FileUploads struct {
	Files map[string][]string // maps token to [fileName, clientID, fileSize, extension]
}

// UploadEndpoints holds active upload endpoints.
var UploadEndpoints struct {
	Uploads map[string]string
}

// ListAllFilesFromFolder prints all file paths in the upload folder.
func ListAllFilesFromFolder() {
	fmt.Println("Listing all files from folder")
	fmt.Println(FileUploads.Files)

	err := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("Error:", err)
			return err
		}
		if !d.IsDir() {
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// CleanAllFilesFromFolder removes and recreates the upload folder.
func CleanAllFilesFromFolder() {
	fmt.Println("Cleaning all files from folder")
	err := os.RemoveAll(folder)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(folder, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

// AllowedHost for upload endpoint requests.
var AllowedHost = "localhost:3002"

// getUploadEndpoint creates an upload token and returns the upload URL.
func getUploadEndpoint(c *gin.Context) {
	fmt.Println(c.Request.Header, c.Request.Header.Get("Origin"), c.Request.Host+c.Request.URL.Path)

	if c.Request.Host != AllowedHost {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	clientid := c.Query("clientid")
	randomHex := util.RandStringBytesMaskImprSrc(32)
	UploadEndpoints.Uploads = make(map[string]string)
	UploadEndpoints.Uploads[randomHex] = clientid

	fmt.Println(randomHex, UploadEndpoints.Uploads)
	dnsCdn := util.EnvGetString("DNS_CDN", true)
	c.JSON(http.StatusOK, gin.H{
		"upload_url": dnsCdn + `/api/upload/` + randomHex,
	})
}

// bufferPool provides fixed-size 32 KB buffers to reduce per-request allocations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 32 KB buffer
	},
}

// uploadEndpoint streams a large file upload directly to disk while limiting memory usage.
// It uses MultipartReader to process the file part and enforces a max upload size.
func uploadEndpoint(c *gin.Context) {
	const paramName = "file"
	const maxUploadSize = 1 << 30 // 1GB maximum upload size

	// Wrap the request body to enforce maximum size.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	// Validate token and extract upload token from URL.
	tokenHeader := c.Request.Header.Get("Authorization")
	uploadToken := c.Param("uploadtoken")
	if uploadToken == "" {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	// Optionally get fileName from form data.
	fileName := c.PostForm("fileName")

	// Validate Bearer token.
	if tokenHeader == "" || len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
		c.String(http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	tokenHeader = tokenHeader[7:] // Remove "Bearer " prefix.
	user := FetchTokenInfo(c, tokenHeader)
	if user == nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	// Ensure the uploads directory exists.
	if err := os.MkdirAll(folder, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	// Create a multipart reader to stream form parts without buffering entire file in memory.
	mr, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create multipart reader: " + err.Error()})
		return
	}

	var written int64
	// Get a buffer from the pool.
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	// Process each part.
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading multipart data: " + err.Error()})
			return
		}

		if part.FormName() == paramName {
			// Use the provided fileName or fallback to the part's filename.
			if fileName == "" {
				fileName = part.FileName()
			}
			// Define file path and create the destination file.
			filePath := filepath.Join(folder, fileName)
			dst, err := os.Create(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file: " + err.Error()})
				return
			}
			defer dst.Close()

			// Stream file data to disk using the buffer.
			written, err = io.CopyBuffer(dst, part, buf)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving file: " + err.Error()})
				return
			}
			// Exit loop once the file is processed.
			break
		} else {
			// Discard non-file parts.
			_, _ = io.Copy(io.Discard, part)
		}
	}

	// If no file data was written, return an error.
	if written == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Store file metadata.
	if FileUploads.Files == nil {
		FileUploads.Files = make(map[string][]string)
	}
	ext := filepath.Ext(fileName)
	metadata := []string{strconv.FormatInt(written, 10), ext}
	FileUploads.Files[uploadToken] = append([]string{fileName, user.ClientID}, metadata...)

	// Construct the file URL for download/callback.
	dnsCdn := util.EnvGetString("DNS_CDN", true)
	fileUrl := fmt.Sprintf("%s/api/download/%s?token=%s", dnsCdn, user.ClientID, uploadToken)

	FetchFileCallback(c, user.ClientID, fileUrl, fileName, written, ext)
	c.JSON(http.StatusOK, gin.H{"file_url": fileUrl})
}

// getFileEndpoint serves a file download based on clientID and token.
func getFileEndpoint(c *gin.Context) {
	clientid := c.Param("clientid")
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	fmt.Println(clientid, token)

	data, exists := FileUploads.Files[token]
	if !exists || data[1] != clientid {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	fileName := data[0]
	fmt.Println(fileName)

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.File("./uploads/images/" + fileName)
}
