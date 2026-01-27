# GoCMS: Google Sheets Powered CMS

GoCMS is a production-ready tutorial demonstrating how to use Google Sheets as a low-cost, high-flexibility backend for a modern web application powered by **ZenoEngine**.

## Features
- **Dynamic Content**: Pages are managed via Google Sheets.
- **Lead Generation**: Contact form data is appended directly to a spreadsheet.
- **Modern UI**: Styled with Tailwind CSS via CDN.
- **Type Safety**: Ensured by ZenoEngine's `gsheet` slots.

## 1. Google Sheets Setup

### A. Spreadsheet Structure
Create a new Google Spreadsheet and add two sheets:

#### 1. "Pages" Sheet
Headers (Row 1): `Title`, `Content`, `Slug`
Example Data:
- `Home`, `Welcome to our site!`, `home`
- `About`, `We are a high-performance team.`, `about`

#### 2. "Contacts" Sheet
Headers (Row 1): `Name`, `Email`, `Message`, `Date`

### B. Service Account
1. Go to [Google Cloud Console](https://console.cloud.google.com/).
2. Create a Project and enable **Google Sheets API**.
3. Create a **Service Account**, download the **JSON credentials**.
4. **Important**: Share your Spreadsheet with the Service Account email (Editor access).

## 2. ZenoEngine Setup

Add these variables to your `.env` file:

```env
GSHEET_ID=your_spreadsheet_id_here
GSHEET_CREDS=/path/to/your/service-account.json
```

## 3. How to Run

Execute the following command from the project root:

```bash
go run cmd/zeno/main.go --path src/tutorial/gocms
```

Visit `http://localhost:8080` to see your CMS in action!
