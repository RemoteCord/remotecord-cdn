package router

import (
	"bytes"
	// . "cdn/api/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var folder = "./uploads/images"

var FileUploads struct {

	Files map[string][]string  // maps clientID to array of file paths
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


func FetchTokenInfo(c *gin.Context, token string, tokenFile string) *struct {
	ClientID string `json:"clientid"`
	Email    string `json:"email"`
	Username string `json:"username"`
} {
	payload := []byte(`{"token": "` + token + `", "tokenFile": "` + tokenFile + `"}`)

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api2.luqueee.dev/api/cdn/decode-token", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers (adjust as needed)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// log.Fatal("Error sending request:", err)
		return nil
	}
	defer resp.Body.Close()

	// Read response body
	var tokenResponse struct {
		ClientID string `json:"clientid"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	// Print the response body for debugging
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Error reading response:", err)
	}
	
	// Check if response contains error
	if _, hasError := result["error"]; hasError {
		return nil
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}
	fmt.Println("Response JSON:", string(jsonBytes))

	// Reset the response body for later decoding
	resp.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Fatal("Error decoding response:", err)
	}
	fmt.Printf("User: %s, Email: %s, ClientID: %s StatusCode: %d \n", tokenResponse.Username, tokenResponse.Email, tokenResponse.ClientID, resp.StatusCode)
	fmt.Println(http.StatusOK)
	// Check response status
	if resp.StatusCode != 201 {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return nil
	}

	// Return the token response

	return &tokenResponse
}

func FetchTokenFile(c *gin.Context, tokenFile string, clientid string) *struct {
	Status bool `json:"status"`
} {

	// Create HTTP request
	req, err := http.NewRequest("GET", `https://api2.luqueee.dev/api/cdn/verify-token-file?token=` + tokenFile + `&clientid=` + clientid, nil )
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers (adjust as needed)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// log.Fatal("Error sending request:", err)
		return nil
	}
	defer resp.Body.Close()

	// Read response body
	var tokenResponse struct {
		Status bool `json:"status"`
	}

	// Print the response body for debugging
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Error reading response:", err)
	}
	
	// Check if response contains error
	if _, hasError := result["error"]; hasError {
		return nil
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}
	fmt.Println("Response JSON:", string(jsonBytes))

	// Reset the response body for later decoding
	resp.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Fatal("Error decoding response:", err)
	}
	// Set the status field from the response
	tokenResponse.Status = result["status"].(bool)
	fmt.Printf("Status: %t StatusCode: %d \n", tokenResponse.Status, resp.StatusCode)
	fmt.Println(http.StatusOK)
	// Check response status
	if resp.StatusCode != 200 {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return nil
	}

	// Return the token response

	return &tokenResponse
}

func uploadEndpoint(c *gin.Context) {
	const paramName = "file"

	tokenHeader := c.Request.Header.Get("Authorization")

	realFileName := c.PostForm("fileName")
	tokenFile := c.PostForm("tokenFile")

	fmt.Println(realFileName, tokenFile)

	if tokenHeader == "" || len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
		c.String(http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}

	if tokenFile == ""  {
		c.String(http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	tokenHeader = tokenHeader[7:] // Remove "Bearer " prefix

	fmt.Println(tokenHeader, tokenFile)

	user := FetchTokenInfo(c, tokenHeader, tokenFile)
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
		err = c.SaveUploadedFile(fileHeader, "./uploads/images/"+realFileName)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to save uploaded file: %s", err.Error())
			return
		}
	// }

	if FileUploads.Files == nil {
		FileUploads.Files = make(map[string][]string)
	}
	FileUploads.Files[user.ClientID] = []string{realFileName, tokenFile}


	

	body := gin.H{
		"file_url": realFileName,
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

	verifyToken := FetchTokenFile(c, token, clientid)

	fmt.Println(verifyToken)

	if verifyToken == nil || !verifyToken.Status {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}


	data, exists := FileUploads.Files[clientid]
	if !exists  {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	fileName := data[0]
	tokenFile := data[1]

	fmt.Println(fileName, tokenFile)



	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.File("./uploads/images/" + fileName)
		// c.JSON(http.StatusOK, gin.H{
		// 	"status": "ok",
		// })

}