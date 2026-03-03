# Reporting Engine (HTML to PDF)

## Overview

A robust Enterprise Resource Planning (ERP) or Point of Sale (POS) system requires the ability to instantly generate pixel-perfect, printable documents like Invoices, Receipts, and Financial Reports.

ZenoEngine provides a native **Reporting Engine** via the `pdf.*` module. It uses `wkhtmltopdf` under the hood to perform high-speed, headless HTML-to-PDF conversions directly within your ZenoLang execution flow.

## Prerequisites

Because ZenoEngine leverages the industry-standard `wkhtmltopdf` engine, you must install the `wkhtmltopdf` binary on your host server.

- **Ubuntu/Debian:** `sudo apt-get install wkhtmltopdf`
- **MacOS:** `brew install wkhtmltopdf`
- **Windows:** Download the installer from the [official website](https://wkhtmltopdf.org/).

## Generating a PDF (In-Memory)

The `pdf.generate` slot takes raw HTML strings and converts them into a PDF byte array. This is extremely powerful when combined with `view.render` to turn dynamic Blade templates into PDFs.

```zeno
// 1. Render an HTML invoice using Blade
view.render: 'reports.invoice' {
    data: $invoice_data
    as: $html_body
}

// 2. Convert the HTML to a PDF byte array
pdf.generate: {
    html: $html_body
    orientation: 'Portrait'
    page_size: 'A4'
    as: $pdf_bytes
}

// You can now save $pdf_bytes to S3 or attach it to an email!
```

## Instant Downloading

If you want to generate a PDF and instantly force the user's browser to download it, use the `pdf.download` slot. It automatically sets the correct HTTP `Content-Type` and `Content-Disposition` headers.

```zeno
http.get: '/invoices/{id}/download' {
    orm.model: 'invoices'
    orm.find: $id { as: $invoice }
    
    // Render HTML
    view.render: 'reports.invoice' {
        data: { invoice: $invoice }
        as: $html_body
    }
    
    // Force browser download
    pdf.download: {
        html: $html_body
        filename: 'Invoice-' + $invoice.invoice_number + '.pdf'
        orientation: 'Portrait'    // Optional: 'Landscape'
        page_size: 'Letter'        // Optional: 'A4', 'Legal', etc
    }
}
```

## Configuration Options

Both `pdf.generate` and `pdf.download` accept the following optional parameters to customize the output document:

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `orientation` | `string` | `Portrait` | Page orientation (`Portrait` or `Landscape`) |
| `page_size` | `string` | `A4` | Page dimensions (`A4`, `Letter`, `Legal`, etc) |

With this engine, generating dynamic, beautifully styled PDF reports takes less than 5 lines of code!
