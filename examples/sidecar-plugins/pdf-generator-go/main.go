package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
)

// Request structure
type Request struct {
	ID         string          `json:"id"`
	SlotName   string          `json:"slot_name"`
	Parameters json.RawMessage `json:"parameters"`
	Type       string          `json:"type"`
}

// Response structure
type Response struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type InvoiceParams struct {
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		if req.SlotName == "plugin_init" {
			sendResponse(req.ID, map[string]string{
				"name":        "pdf-generator",
				"version":     "1.0.0",
				"description": "Native Go PDF Generator",
			})
			continue
		}

		if req.SlotName == "plugin_register_slots" {
			slots := []map[string]interface{}{
				{
					"name":        "pdf.generate_invoice",
					"description": "Generate a simple invoice PDF",
					"inputs": map[string]interface{}{
						"customer": map[string]interface{}{"type": "string"},
						"amount":   map[string]interface{}{"type": "number"},
					},
				},
			}
			sendResponse(req.ID, map[string]interface{}{"slots": slots})
			continue
		}

		if req.Type == "guest_call" && req.SlotName == "pdf.generate_invoice" {
			var params InvoiceParams
			if err := json.Unmarshal(req.Parameters, &params); err != nil {
				sendError(req.ID, "Invalid parameters: "+err.Error())
				continue
			}

			pdfBase64, err := generateInvoice(params)
			if err != nil {
				sendError(req.ID, "Failed to generate PDF: "+err.Error())
				continue
			}

			sendResponse(req.ID, map[string]string{
				"pdf_base64": pdfBase64,
			})
		} else {
			sendError(req.ID, "Unknown slot or request type")
		}
	}
}

func generateInvoice(params InvoiceParams) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "INVOICE")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s", params.Customer))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: $%.2f", params.Amount))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(100, 10, "Item")
	pdf.Cell(40, 10, "Price")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 10, "Consulting Services")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", params.Amount))
	pdf.Ln(8)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func sendResponse(id string, data interface{}) {
	resp := Response{
		Type:    "guest_response",
		ID:      id,
		Success: true,
		Data:    data,
	}
	jsonBytes, _ := json.Marshal(resp)
	fmt.Println(string(jsonBytes))
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
