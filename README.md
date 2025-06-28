# Wallabag RSS Tool

A tool to fetch articles from RSS feeds and automatically send them to your Wallabag instance.

## Features

- ✅ Fetch articles from multiple RSS feeds
- ✅ Automatically send articles to Wallabag
- ✅ Web interface for feed management
- ✅ Configurable polling intervals
- ✅ Article deduplication
- ✅ Built with Go and Templ for type-safe templates
- ✅ SQLite database for persistence
- ✅ HTMX for dynamic UI interactions

## Prerequisites

- Go 1.24.1 or later
- A running Wallabag instance
- Wallabag API credentials

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/wallabag-rss-tool.git
   cd wallabag-rss-tool
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install development tools:**
   ```bash
   make install-tools
   ```

## Configuration

Set the following environment variables with your Wallabag API credentials:

```bash
export WALLABAG_BASE_URL="https://your-wallabag-instance.com"
export WALLABAG_CLIENT_ID="your_client_id"
export WALLABAG_CLIENT_SECRET="your_client_secret"
export WALLABAG_USERNAME="your_username"
export WALLABAG_PASSWORD="your_password"
```

## Building and Running

### Using Make (recommended):

```bash
# Build the application
make build

# Run the application
make run

# Run in development mode
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Manual build:

```bash
# Generate templ files
templ generate ./views

# Build
go build -o wallabag-rss-tool .

# Run
./wallabag-rss-tool
```

## Usage

1. **Start the application:**
   ```bash
   make run
   ```

2. **Open your browser** and navigate to `http://localhost:8080`

3. **Add RSS feeds** through the web interface

4. **Configure settings** including default polling intervals

5. **Monitor processed articles** in the Articles section

## Development

### Project Structure

```
├── pkg/
│   ├── config/          # Configuration management
│   ├── database/        # Database operations and models
│   ├── models/          # Data structures
│   ├── rss/             # RSS feed processing
│   ├── server/          # HTTP server and handlers
│   ├── wallabag/        # Wallabag API client
│   └── worker/          # Background feed processing
├── views/               # Templ templates
├── templates/           # Legacy HTML templates (deprecated)
├── db/                  # Database schema
└── main.go             # Application entry point
```

### Technology Stack

- **Backend:** Go 1.24.1
- **Templates:** [Templ](https://github.com/a-h/templ) - Type-safe Go templates
- **Frontend:** Bootstrap 5 + HTMX
- **Database:** SQLite
- **Testing:** Go testing + testify + gomock

### Working with Templates

The project uses Templ for type-safe HTML templates. Template files are located in the `views/` directory with `.templ` extension.

To modify templates:
1. Edit the `.templ` files in the `views/` directory
2. Run `make generate` to regenerate Go code
3. Build and test your changes

For development with live reload:
```bash
make watch  # In one terminal
make dev    # In another terminal
```

### Database Schema

The application uses SQLite with the following tables:
- `feeds` - RSS feed configurations
- `articles` - Processed articles
- `settings` - Application settings

Schema is automatically applied on startup from `db/schema.sql`.

### Testing

The project has comprehensive unit tests with >80% coverage:

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run specific package tests
go test ./pkg/models/
```

## API Endpoints

- `GET /` - Dashboard
- `GET /feeds` - Feed management page
- `POST /feeds` - Add new feed
- `PUT /feeds/{id}` - Update feed
- `DELETE /feeds/{id}` - Delete feed
- `GET /articles` - View processed articles
- `GET /settings` - Application settings
- `POST /sync` - Trigger manual sync

## Configuration Options

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `WALLABAG_BASE_URL` | Your Wallabag instance URL | Yes |
| `WALLABAG_CLIENT_ID` | Wallabag API client ID | Yes |
| `WALLABAG_CLIENT_SECRET` | Wallabag API client secret | Yes |
| `WALLABAG_USERNAME` | Your Wallabag username | Yes |
| `WALLABAG_PASSWORD` | Your Wallabag password | Yes |

### Feed Settings

- **Name:** Display name for the feed
- **URL:** RSS/Atom feed URL
- **Poll Interval:** How often to check for new articles (minutes, 0 = use default)

## Troubleshooting

### Common Issues

1. **"Wallabag credentials not configured"**
   - Ensure all required environment variables are set
   - Verify your Wallabag instance is accessible
   - Check your API credentials

2. **"Failed to parse feed"**
   - Verify the RSS feed URL is accessible
   - Check if the feed format is valid RSS/Atom

3. **Database errors**
   - Ensure the application has write permissions in its directory
   - Check disk space availability

### Logs

The application logs important events to stdout. For debugging:

```bash
# Run with verbose logging
go run . 2>&1 | tee app.log
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Templ](https://github.com/a-h/templ) for type-safe Go templates
- [HTMX](https://htmx.org/) for dynamic interactions
- [Bootstrap](https://getbootstrap.com/) for UI components
- [Wallabag](https://wallabag.org/) for the read-it-later service