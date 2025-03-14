package router

import (

	// . "cdn/api/util"

	"cdn/api/util"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var folder = "./uploads/images"

var FileUploads struct {

	Files map[string][]string  // maps clientID to array of file paths
}

var UploadEndpoints struct {
	Uploads map[string]string
}


func ListAllFilesFromFolder() {

	fmt.Println("Listing all files from folder")
	fmt.Println(FileUploads.Files)

	err := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("Error:", err)
			return err
		}

		// Only print files (excluding directories)
		if !d.IsDir() {
			fmt.Println(path) // You can print just d.Name() for file names
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

} 

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


var AllowedHost = "localhost:3002"



func getUploadEndpoint(c *gin.Context) {
	fmt.Println(c.Request.Header, c.Request.Header.Get("Origin"),c.Request.Host+c.Request.URL.Path)
	
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
		"upload_url": dnsCdn + `/api/upload/`+randomHex,
	})
}


func uploadEndpoint(c *gin.Context) {
	const paramName = "file"
	const folder = "./uploads/images/"

	tokenHeader := c.Request.Header.Get("Authorization")
	uploadToken := c.Param("uploadtoken")

	if uploadToken == "" {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	realFileName := c.PostForm("fileName")

	if tokenHeader == "" || len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
		c.String(http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}

	tokenHeader = tokenHeader[7:] // Remove "Bearer " prefix
	user := FetchTokenInfo(c, tokenHeader)
	if user == nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	fileHeader, err := c.FormFile(paramName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload: " + err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file: " + err.Error()})
		return
	}
	defer file.Close()

	// Ensure the uploads directory exists
	err = os.MkdirAll(folder, 0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	// Define the full path for the uploaded file
	filePath := filepath.Join(folder, realFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file: " + err.Error()})
		return
	}
	defer dst.Close()

	// Reset file pointer after initial read
	file.Seek(0, 0)

	// Efficiently copy file to disk
	written, err := io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying file: " + err.Error()})
		return
	}

	// Force garbage collection after processing large files
	runtime.GC()

	// Store file metadata
	if FileUploads.Files == nil {
		FileUploads.Files = make(map[string][]string)
	}
	metadataSplit := strings.Split(fileHeader.Filename, ".")
	extFile := metadataSplit[len(metadataSplit)-1]

	metadata := []string{strconv.FormatInt(written, 10), extFile}
	data := append([]string{realFileName, user.ClientID}, metadata...)

	FileUploads.Files[uploadToken] = data

	dnsCdn := util.EnvGetString("DNS_CDN", true)
	fileUrl := fmt.Sprintf("%s/api/download/%s?token=%s", dnsCdn, user.ClientID, uploadToken)

	FetchFileCallback(c, user.ClientID, fileUrl, fileHeader.Filename, written, extFile)

	c.JSON(http.StatusOK, gin.H{"file_url": fileUrl})
}


func getFileEndpoint(c *gin.Context) {
	clientid := c.Param("clientid")
	token := c.Query("token")

	if token == "" {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	
	fmt.Println(clientid, token)

	// verifyToken := FetchTokenFile(c, token, clientid)

	// fmt.Println(verifyToken)

	// if verifyToken == nil || !verifyToken.Status {
	// 	c.String(http.StatusUnauthorized, "Invalid token")
	// 	return
	// }


	data, exists := FileUploads.Files[token]
	if !exists || data[1] != clientid {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	fileName := data[0]

	fmt.Println(fileName)



	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.File("./uploads/images/" + fileName)
		// c.JSON(http.StatusOK, gin.H{
		// 	"status": "ok",
		// })

}