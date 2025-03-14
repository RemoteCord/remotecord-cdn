package router

import (
	"bytes"
	"cdn/api/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)


func FetchTokenInfo(c *gin.Context, token string, ) *struct {
	ClientID string `json:"clientid"`
	Email    string `json:"email"`
	Username string `json:"username"`
} {
	payload := []byte(`{"token": "` + token + `"}`)

		apiUrl := util.EnvGetString("API_URL", true)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiUrl + `/api/cdn/decode-token`, bytes.NewBuffer(payload))
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

func FetchFileCallback(c *gin.Context, clientid string, fileurl string, filename string, filesize int64, fileformat string ) {
	metadata := map[string]interface{}{
		"filename": filename,
		"size":    filesize,
		"format":  fileformat,
	}

	uploadCallback := map[string]interface{}{
		"fileurl":  fileurl,
		"clientid": clientid,
		"metadata": metadata,
	}

	payload, err := json.Marshal(uploadCallback)
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	fmt.Println(string(payload))

	apiUrl := util.EnvGetString("API_URL", true)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiUrl + `/api/cdn/upload`, bytes.NewBuffer(payload))
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
		return
	}
	defer resp.Body.Close()

	// Print the response body for debugging
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Error reading response:", err)
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}
	fmt.Println("Response JSON:", string(jsonBytes))
	
}

func FetchTokenFile(c *gin.Context, tokenFile string, clientid string) *struct {
	Status bool `json:"status"`
} {

		apiUrl := util.EnvGetString("API_URL", true)


	// Create HTTP request
	req, err := http.NewRequest("GET", apiUrl +`/api/cdn/verify-token-file?token=` + tokenFile + `&clientid=` + clientid, nil )
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