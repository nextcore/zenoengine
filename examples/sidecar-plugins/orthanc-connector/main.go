package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Global configuration
var config struct {
	URL      string
	User     string
	Password string
}

type Request struct {
	ID         string          `json:"id"`
	SlotName   string          `json:"slot_name"`
	Parameters json.RawMessage `json:"parameters"`
	Type       string          `json:"type"`
}

type Response struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	// Initialize config from environment variables
	config.URL = os.Getenv("ORTHANC_URL")
	if config.URL == "" {
		config.URL = "http://localhost:8042"
	}
	// Remove trailing slash
	config.URL = strings.TrimSuffix(config.URL, "/")

	config.User = os.Getenv("ORTHANC_USER")
	config.Password = os.Getenv("ORTHANC_PASSWORD")

	scanner := bufio.NewScanner(os.Stdin)
	// Important: Increase scanner buffer size to handle larger requests if needed,
	// though standard input is usually streamed differently in Go.
	// But since we fixed the host side to use Decoder, the plugin side using Scanner is fine
	// AS LONG AS the HOST sends newline-delimited JSON.
	// Wait, the host sends newline-delimited JSON. So Scanner is okay for reading REQUESTS
	// (requests are usually small).
	// But RESPONSES (from plugin to host) can be large. We use fmt.Println which adds newline.

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(req.ID, "Invalid JSON request: "+err.Error())
			continue
		}

		handleRequest(req)
	}
}

func handleRequest(req Request) {
	switch req.SlotName {
	case "plugin_init":
		sendResponse(req.ID, map[string]string{
			"name":        "orthanc-connector",
			"version":     "1.0.0",
			"description": "Orthanc PACS Connector",
		})

	case "plugin_register_slots":
		slots := []map[string]interface{}{
			{
				"name":        "pacs.search_patient",
				"description": "Search patients in Orthanc",
				"inputs": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Patient Name or ID"},
				},
			},
			{
				"name":        "pacs.upload_instance",
				"description": "Upload a DICOM file to Orthanc",
				"inputs": map[string]interface{}{
					"filepath": map[string]interface{}{"type": "string", "description": "Absolute path to .dcm file"},
				},
			},
			{
				"name":        "pacs.get_study",
				"description": "Get details of a study",
				"inputs": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Study UUID"},
				},
			},
			{
				"name":        "pacs.preview_instance",
				"description": "Get preview image (PNG base64) of an instance",
				"inputs": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Instance UUID"},
				},
			},
			{
				"name":        "pacs.system_info",
				"description": "Get Orthanc system status",
			},
		}
		sendResponse(req.ID, map[string]interface{}{"slots": slots})

	case "pacs.system_info":
		resp, err := callOrthanc("GET", "/system", nil)
		if err != nil {
			sendError(req.ID, "Orthanc connection failed: "+err.Error())
			return
		}
		var info interface{}
		json.Unmarshal(resp, &info)
		sendResponse(req.ID, info)

	case "pacs.search_patient":
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(req.Parameters, &params); err != nil {
			sendError(req.ID, "Invalid parameters")
			return
		}

		// Orthanc Tools/Find API
		// Payload: {"Level":"Patient", "Query": {"PatientName": "*query*"}}
		payload := map[string]interface{}{
			"Level": "Patient",
			"Query": map[string]string{
				"PatientName": "*" + params.Query + "*",
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		resp, err := callOrthanc("POST", "/tools/find", bytes.NewReader(payloadBytes))
		if err != nil {
			sendError(req.ID, "Search failed: "+err.Error())
			return
		}

		// Response is list of IDs. We should fetch details for each (or at least some).
		// For MVP, just return the list of IDs
		var ids []string
		json.Unmarshal(resp, &ids)
		sendResponse(req.ID, map[string]interface{}{"ids": ids, "count": len(ids)})

	case "pacs.upload_instance":
		var params struct {
			Filepath string `json:"filepath"`
		}
		if err := json.Unmarshal(req.Parameters, &params); err != nil {
			sendError(req.ID, "Invalid parameters")
			return
		}

		// Read file
		fileBytes, err := os.ReadFile(params.Filepath)
		if err != nil {
			sendError(req.ID, "Failed to read file: "+err.Error())
			return
		}

		// Upload to /instances
		resp, err := callOrthanc("POST", "/instances", bytes.NewReader(fileBytes))
		if err != nil {
			sendError(req.ID, "Upload failed: "+err.Error())
			return
		}

		var uploadResp interface{}
		json.Unmarshal(resp, &uploadResp)
		sendResponse(req.ID, uploadResp)

	case "pacs.get_study":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(req.Parameters, &params); err != nil {
			sendError(req.ID, "Invalid parameters")
			return
		}

		resp, err := callOrthanc("GET", "/studies/"+params.ID, nil)
		if err != nil {
			sendError(req.ID, "Get study failed: "+err.Error())
			return
		}
		var study interface{}
		json.Unmarshal(resp, &study)
		sendResponse(req.ID, study)

	case "pacs.preview_instance":
		var params struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(req.Parameters, &params); err != nil {
			sendError(req.ID, "Invalid parameters")
			return
		}

		// Get preview (PNG)
		resp, err := callOrthanc("GET", "/instances/"+params.ID+"/preview", nil)
		if err != nil {
			sendError(req.ID, "Get preview failed: "+err.Error())
			return
		}

		// Encode to Base64
		encoded := base64.StdEncoding.EncodeToString(resp)
		sendResponse(req.ID, map[string]string{
			"mime": "image/png",
			"data": encoded,
		})

	default:
		sendError(req.ID, "Unknown slot: "+req.SlotName)
	}
}

func callOrthanc(method, path string, body io.Reader) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := config.URL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if config.User != "" {
		req.SetBasicAuth(config.User, config.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("orthanc returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func sendResponse(id string, data interface{}) {
	resp := Response{
		Type:    "guest_response",
		ID:      id,
		Success: true,
		Data:    data,
	}
	jsonBytes, _ := json.Marshal(resp)
	fmt.Println(string(jsonBytes)) // Newline is important
}

func sendError(id string, msg string) {
	resp := Response{
		Type:    "guest_response",
		ID:      id,
		Success: false,
		Error:   msg,
	}
	jsonBytes, _ := json.Marshal(resp)
	fmt.Println(string(jsonBytes))
}
