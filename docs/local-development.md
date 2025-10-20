# Local Development Guide

## Quick Start with Docker

### Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

### Starting the Database

**Option 1: Basic PostgreSQL only**
```bash
docker-compose up -d
```

This starts:
- PostgreSQL 15 on `localhost:5432`
- Default credentials: `postgres/postgres`
- Database name: `esg_tube`

**Option 2: With pgAdmin (database management UI)**
```bash
docker-compose --profile tools up -d
```

This additionally starts:
- pgAdmin on `http://localhost:5050`
- Login: `admin@localhost.com` / `admin`

**Option 3: Development mode with verbose logging**
```bash
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

### Stopping the Database

```bash
# Stop containers but keep data
docker-compose down

# Stop and remove all data (fresh start)
docker-compose down -v
```

### Checking Database Status

```bash
# View logs
docker-compose logs -f postgres

# Check if database is ready
docker-compose exec postgres pg_isready -U postgres

# Connect to database with psql
docker-compose exec postgres psql -U postgres -d esg_tube
```

## Database Setup

### 1. Restore Existing Schema

If you have an existing database backup:

```bash
# Copy backup file into container
docker cp docs/db_cluster-06-06-2025@08-19-26.backup esg-tube-postgres:/tmp/backup.sql

# Restore the backup
docker-compose exec postgres psql -U postgres -d esg_tube -f /tmp/backup.sql
```

**Or** uncomment the volume mount in `docker-compose.yml`:
```yaml
volumes:
  - ./docs/db_cluster-06-06-2025@08-19-26.backup:/docker-entrypoint-initdb.d/01-init.sql
```

Then recreate the container:
```bash
docker-compose down -v
docker-compose up -d
```

### 2. Verify Tables Exist

```bash
docker-compose exec postgres psql -U postgres -d esg_tube -c "\dt"
```

Expected tables:
- `episodes`
- `playlists`
- `playlist_items`
- `presenters`
- `organizations`
- `video_presenters`
- `video_organizations`

### 3. Create Tables Manually (if needed)

If the schema doesn't exist, create the tables manually:

```sql
-- Connect to database
docker-compose exec postgres psql -U postgres -d esg_tube

-- Create episodes table
CREATE TABLE IF NOT EXISTS episodes (
    id SERIAL PRIMARY KEY,
    video_id VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    published_at TIMESTAMP,
    image1 VARCHAR(500),
    image2 VARCHAR(500),
    image3 VARCHAR(500),
    view_count INTEGER DEFAULT 0,
    video_site INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT true,
    sort_order INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create playlists table
CREATE TABLE IF NOT EXISTS playlists (
    id SERIAL PRIMARY KEY,
    playlist_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    publish_date DATE,
    image VARCHAR(500),
    website VARCHAR(500),
    hashtag VARCHAR(100),
    playlist_category_id INTEGER,
    slug VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create playlist_items table
CREATE TABLE IF NOT EXISTS playlist_items (
    id SERIAL PRIMARY KEY,
    playlist_id INTEGER REFERENCES playlists(id) ON DELETE CASCADE,
    episode_id INTEGER REFERENCES episodes(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(playlist_id, episode_id)
);

-- Create presenters table (manually maintained)
CREATE TABLE IF NOT EXISTS presenters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    biography TEXT,
    twitter VARCHAR(100),
    email VARCHAR(255),
    website VARCHAR(500),
    avatar_url VARCHAR(500),
    byline VARCHAR(500),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create organizations table (manually maintained)
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    website VARCHAR(500),
    twitter VARCHAR(100),
    contact_person VARCHAR(255),
    image VARCHAR(500),
    slug VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Environment Configuration

### 1. Copy Environment File

```bash
cp .env.example .env
```

### 2. Update .env with Docker Settings

```bash
# Database configuration for Docker
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=esg_tube
POSTGRES_SSLMODE=disable

# YouTube API key (get from Google Cloud Console)
YOUTUBE_API_KEY=your-youtube-api-key-here
```

### 3. Get YouTube API Key

1. Go to [Google Cloud Console](https://console.developers.google.com/)
2. Create a new project or select existing
3. Enable **YouTube Data API v3**
4. Create credentials → API key
5. Copy the API key to `.env`

## Running the Application

### 1. Build the Application

```bash
go build -o youtube-sync cmd/youtube-sync/main.go
```

### 2. Run Commands

```bash
# Test configuration
./youtube-sync version

# Sync a single video
./youtube-sync video dQw4w9WgXcQ --verbose

# Sync a channel
./youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --limit 10
```

## Testing

### Run Unit Tests

```bash
go test ./tests/unit/... -v
```

### Run Integration Tests (requires database)

```bash
# Make sure Docker Compose is running
docker-compose up -d

# Run integration tests
go test ./tests/integration/... -v

# Run all tests
go test ./... -v
```

### Run Tests with Coverage

```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting

### Port 5432 Already in Use

If you have PostgreSQL already running locally:

**Option 1:** Stop local PostgreSQL
```bash
# macOS
brew services stop postgresql

# Linux (systemd)
sudo systemctl stop postgresql

# Linux (init.d)
sudo service postgresql stop
```

**Option 2:** Change the port in `docker-compose.yml`
```yaml
ports:
  - "5433:5432"  # Use 5433 externally
```

Then update `.env`:
```bash
POSTGRES_PORT=5433
```

### Database Connection Refused

```bash
# Check if container is running
docker-compose ps

# Check logs
docker-compose logs postgres

# Restart container
docker-compose restart postgres
```

### Reset Database to Clean State

```bash
# Stop and remove all data
docker-compose down -v

# Start fresh
docker-compose up -d

# Restore schema if needed
docker cp docs/db_cluster-06-06-2025@08-19-26.backup esg-tube-postgres:/tmp/backup.sql
docker-compose exec postgres psql -U postgres -d esg_tube -f /tmp/backup.sql
```

### pgAdmin Cannot Connect to PostgreSQL

When using pgAdmin in Docker, use these connection settings:
- **Host**: `postgres` (use container name, not localhost)
- **Port**: `5432`
- **Username**: `postgres`
- **Password**: `postgres`
- **Database**: `esg_tube`

### Permission Issues with Volumes

```bash
# Linux: Fix permissions
sudo chown -R $USER:$USER postgres_data/

# Or recreate with proper permissions
docker-compose down -v
docker-compose up -d
```

## Development Workflow

### 1. Start Database
```bash
docker-compose up -d
```

### 2. Make Code Changes
Edit files in your IDE

### 3. Run Tests
```bash
go test ./... -v
```

### 4. Build and Test Locally
```bash
go build -o youtube-sync cmd/youtube-sync/main.go
./youtube-sync video dQw4w9WgXcQ --verbose
```

### 5. View Database Changes
```bash
# Via psql
docker-compose exec postgres psql -U postgres -d esg_tube -c "SELECT * FROM episodes LIMIT 5;"

# Or via pgAdmin
open http://localhost:5050
```

### 6. Stop Database When Done
```bash
docker-compose down
```

## Useful Database Queries

```sql
-- Check synced videos
SELECT video_id, title, view_count, active, created_at
FROM episodes
ORDER BY created_at DESC
LIMIT 10;

-- Count episodes by active status
SELECT active, COUNT(*)
FROM episodes
GROUP BY active;

-- View playlists with video count
SELECT p.playlist_id, p.name, COUNT(pi.episode_id) as video_count
FROM playlists p
LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
GROUP BY p.id, p.playlist_id, p.name
ORDER BY video_count DESC;

-- Find videos synced today
SELECT video_id, title, created_at
FROM episodes
WHERE created_at::date = CURRENT_DATE
ORDER BY created_at DESC;
```

## Next Steps

- Review the [README.md](../README.md) for application usage
- Check [quickstart.md](../specs/001-youtube-video-sync/quickstart.md) for testing scenarios
- Review [spec.md](../specs/001-youtube-video-sync/spec.md) for feature requirements
