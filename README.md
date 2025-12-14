# Internship Alert Bot 🤖

A Go-based web scraper designed to fetch internship job listings from JobsDB. Currently configured to track "Backend Internship" positions.

## Features
- **Basic Scraping**: Fetches job details including Title, Company, Location, Posted Date, and URL.
- **Bot Protection Bypass**: Includes browser header simulation to avoid 403 Forbidden errors.
- **JSON Output**: Automatically saves results to `jobs_output.json` for further processing.
- **Console Summary**: Displays a clean summary of found jobs and top hiring companies in the terminal.

## Prerequisites
- Go 1.25 or higher

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/annop07/internship-alert-bot.git
   cd internship-alert-bot
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

Run the scraper using the following command:

```bash
go run cmd/scraper/main.go
```

The program will:
1. Test connection to JobsDB.
2. Scrape the first page of results.
3. Display the results in the terminal.
4. Save the detailed data to `jobs_output.json`.

## Roadmap (Phase A)
- [ ] Keyword filtering
- [ ] Multi-page scraping support (Page 2, 3, ...)
- [ ] Export to CSV
- [ ] Extract additional details (Salary, Requirements)

## Project Structure
```
├── cmd/
│   └── scraper/       # Main entry point (main.go)
├── pkg/
│   ├── models/        # Data structures (Job struct)
│   └── scraper/       # Core scraping logic
├── jobs_output.json   # Generated output file
└── todo-list.md       # Project roadmap
```

## Disclaimer
This project is for educational purposes only. Please respect the website's terms of service and usage policies.
