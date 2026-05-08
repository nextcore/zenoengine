package slots

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

// RegisterPDFSlots registers the pdf.* slots for generating PDF documents.
func RegisterPDFSlots(eng *engine.Engine) {
	eng.Register("pdf.generate", handlePDFGenerate, engine.SlotMeta{
		Description: "Convert HTML string to PDF bytes using wkhtmltopdf.",
		Inputs: map[string]engine.InputMeta{
			"html":        {Type: "string", Required: true, Description: "The raw HTML string to convert."},
			"as":          {Type: "any", Required: true, Description: "Variable to store the resulting PDF byte array."},
			"orientation": {Type: "string", Required: false, Description: "Page orientation: 'Portrait' or 'Landscape' (Default: Portrait)."},
			"page_size":   {Type: "string", Required: false, Description: "Page size dimension: 'A4', 'Letter', etc (Default: A4)."},
			"error":       {Type: "any", Required: false, Description: "Variable to store any conversion errors."},
		},
		Example: `pdf.generate:
  html: $html_body
  orientation: 'Landscape'
  as: $pdf_bytes
  error: $pdf_error`,
	})

	eng.Register("pdf.download", handlePDFDownload, engine.SlotMeta{
		Description: "Generates a PDF from an HTML string and directly triggers a file download to the client's browser.",
		Inputs: map[string]engine.InputMeta{
			"html":        {Type: "string", Required: true, Description: "The HTML structure to render."},
			"filename":    {Type: "string", Required: true, Description: "The name of the file to be downloaded (e.g. 'report.pdf')."},
			"orientation": {Type: "string", Required: false, Description: "Page orientation (Portrait/Landscape)."},
			"page_size":   {Type: "string", Required: false, Description: "Page size (A4, Letter)."},
		},
		Example: `pdf.download:
  html: $invoice_html
  filename: 'Invoice-101.pdf'`,
	})
}

func handlePDFGenerate(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
	var htmlRaw string
	var orientation = "Portrait"
	var pageSize = "A4"
	var asVar = ""
	var errVar = ""

	for _, c := range node.Children {
		val := parseNodeValue(c, scope)
		switch c.Name {
		case "html":
			htmlRaw = coerce.ToString(val)
		case "orientation":
			orientation = coerce.ToString(val)
		case "page_size":
			pageSize = coerce.ToString(val)
		case "as":
			asVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
		case "error":
			errVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
		}
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		if errVar != "" {
			scope.Set(errVar, err.Error())
		}
		if asVar != "" {
			scope.Set(asVar, nil)
		}
		return nil
	}

	pdfg.Orientation.Set(orientation)
	pdfg.PageSize.Set(pageSize)
	pdfg.MarginTop.Set(10)
	pdfg.MarginBottom.Set(10)
	pdfg.MarginLeft.Set(10)
	pdfg.MarginRight.Set(10)

	page := wkhtmltopdf.NewPageReader(strings.NewReader(htmlRaw))
	pdfg.AddPage(page)

	err = pdfg.Create()
	if err != nil {
		if errVar != "" {
			scope.Set(errVar, err.Error())
		}
		if asVar != "" {
			scope.Set(asVar, nil)
		}
		return nil
	}

	if asVar != "" {
		scope.Set(asVar, pdfg.Bytes())
	}
	if errVar != "" {
		scope.Set(errVar, nil)
	}

	return nil
}

func handlePDFDownload(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
	var htmlRaw string
	var filename = "document.pdf"
	var orientation = "Portrait"
	var pageSize = "A4"

	for _, c := range node.Children {
		val := parseNodeValue(c, scope)
		switch c.Name {
		case "html":
			htmlRaw = coerce.ToString(val)
		case "filename":
			filename = coerce.ToString(val)
		case "orientation":
			orientation = coerce.ToString(val)
		case "page_size":
			pageSize = coerce.ToString(val)
		}
	}

	if !strings.HasSuffix(filename, ".pdf") {
		filename += ".pdf"
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		writeHTTPError(scope, "WKHTMLTOPDF Error: "+err.Error(), 500)
		return nil
	}

	pdfg.Orientation.Set(orientation)
	pdfg.PageSize.Set(pageSize)

	page := wkhtmltopdf.NewPageReader(strings.NewReader(htmlRaw))
	pdfg.AddPage(page)

	err = pdfg.Create()
	if err != nil {
		writeHTTPError(scope, "Failed to render PDF: "+err.Error(), 500)
		return nil
	}

	resRaw, hasRes := scope.Get("http.response")
	if !hasRes {
		return nil // Not in HTTP context
	}
	w, ok := resRaw.(http.ResponseWriter)
	if !ok {
		return nil
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", coerce.ToString(pdfg.Buffer().Len()))

	w.WriteHeader(http.StatusOK)
	_, _ = bytes.NewReader(pdfg.Bytes()).WriteTo(w)

	return nil
}

func writeHTTPError(scope *engine.Scope, msg string, code int) {
	resRaw, hasRes := scope.Get("http.response")
	if !hasRes {
		return
	}
	w, ok := resRaw.(http.ResponseWriter)
	if !ok {
		return
	}
	http.Error(w, msg, code)
}
