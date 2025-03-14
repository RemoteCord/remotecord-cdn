package router

import (

	// . "cdn/api/util"

	"cdn/api/util"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	tokenHeader := c.Request.Header.Get("Authorization")

	uploadToken := c.Param("uploadtoken")

	if uploadToken == "" {
		c.String(http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	realFileName := c.PostForm("fileName")
	// tokenFile := c.PostForm("tokenFile")

	fmt.Println(realFileName, uploadToken)

	if tokenHeader == "" || len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
		c.String(http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}

	tokenHeader = tokenHeader[7:] // Remove "Bearer " prefix

	fmt.Println(tokenHeader, uploadToken)

	user := FetchTokenInfo(c, tokenHeader)
	if user == nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}


	fileHeader, err := c.FormFile(paramName)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid file upload: %s", err.Error())
		return
	}

	fmt.Println(fileHeader.Filename, fileHeader.Size)

	file, err := fileHeader.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to open uploaded file: %s", err.Error())
		return
	}
	defer file.Close()

	// Validate file type based on file content
	fileBuffer := make([]byte, 512) // Read first 512 bytes of file content
	_, err = file.Read(fileBuffer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}
	// fileType := http.DetectContentType(fileBuffer)

	// // Check if file type is allowed (image MIME types whitelist)
	// allowedMIMETypes := map[string]bool{
	// 	"image/jpg":  true,
	// 	"image/jpeg": true,
	// 	"image/png":  true,
	// 	"image/gif":  true,
	// }
	// if !allowedMIMETypes[fileType] {
	// 	c.String(http.StatusBadRequest, "Invalid file upload: file type must be an image. Found '%s'.", fileType)
	// 	return
	// }

	// Upload the file if has not already been uploaded
	// fileHashBuffer := md5.Sum(fileBuffer)
		//err = c.SaveUploadedFile(fileHeader, "./uploads/images/"+fileName)

	// if !alreadyExists {
		// Create the destination file
		dst, err := os.Create("./uploads/images/" + realFileName)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to save uploaded file: %s", err.Error())
			return
		}
		defer dst.Close()

		// Reset the file pointer to beginning after earlier read
		file.Seek(0, 0)

		// Create a buffer for copying
		buf := make([]byte, 32*1024)
		var written int64

		// Copy the file in chunks and report progress
		for {
			n, err := file.Read(buf)
			if n > 0 {
				nw, err := dst.Write(buf[:n])
				if err != nil {
					return
				}
				written += int64(nw)
				progress := float64(written) / float64(fileHeader.Size) * 100
				fmt.Printf("\rUploading... %.2f%%", progress)
			}
			if err != nil {
				break
			}
		}
		fmt.Println("\nUpload complete!")

		
	// }

	if FileUploads.Files == nil {
		FileUploads.Files = make(map[string][]string)
	}
	metadataSplit := strings.Split(fileHeader.Filename, ".")

	extFile := metadataSplit[len(metadataSplit)-1]

	metadata := []string{strconv.FormatInt(fileHeader.Size, 10), extFile}

	data := append([]string{realFileName, user.ClientID}, metadata...)

	fmt.Print(data, metadata)

	FileUploads.Files[uploadToken] =  data

	dnsCdn := util.EnvGetString("DNS_CDN", true)

	fileUrl := dnsCdn + `/api/download/` + user.ClientID + `?token=` + uploadToken

	fmt.Println(fileUrl)
	
	FetchFileCallback(c, user.ClientID, fileUrl, fileHeader.Filename, fileHeader.Size, extFile)


	body := gin.H{
		"file_url": fileUrl,
	}
	c.JSON(http.StatusOK, body)
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
	c.File("uploads/images/" + fileName)
		// c.JSON(http.StatusOK, gin.H{
		// 	"status": "ok",
		// })

}