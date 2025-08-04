# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

M4K (Media for Kiosks) is a comprehensive media processing and management platform built around PocketBase with Go backend extensions. The system handles media upload, processing, transcoding, and distribution for kiosk/device management scenarios with multi-tenant group-based architecture.

## Core Architecture

### Backend Stack
- **PocketBase (v0.28.4)** - Primary database and API layer with custom Go extensions
- **Go (1.23.0)** - Custom backend for media processing, job queue, and transcoding
- **FFmpeg** - Video/audio transcoding and metadata extraction
- **Deno** - TypeScript job execution runtime
- **Docker** - Containerized deployment

### Key Components
1. **Media Processing Pipeline** (`go/cmd/pocketbase/medias.go`) - Auto-extracts metadata from uploaded files
2. **Asynchronous Job System** (`go/cmd/pocketbase/jobs.go`) - Concurrent job execution with progress tracking
3. **Video Transcoding** (`go/cmd/pocketbase/transcode.go`) - Multi-format/quality video processing
4. **Authentication & Rules** (`go/cmd/init-rules/main.go`) - Group-based security with role hierarchy

## Development Commands

### Local Development
```bash
# Start backend development server
cd /home/kaci/Documents/m4k/server/go
./dev.sh

# Build and run production containers
cd /home/kaci/Documents/m4k/server
./docker_build.sh

# Initialize database schema (collections and rules)
./init_schema.sh

# Create admin user
./create_admin.sh
```

### Container Management
- **PocketBase runs on port 8090**
- Admin UI: http://localhost:8090/_/
- API: http://localhost:8090/api/
- Health check: http://localhost:8090/api/health

### Testing Connectivity
```bash
# Test PocketBase connectivity
curl http://localhost:8090/api/health

# Check running containers
docker ps | grep pocketbase
```

## Data Architecture

### Core Collections
- **users** - PocketBase authentication
- **groups** - Multi-tenant organizations  
- **members** - User-group relationships with roles (Viewer:10, Editor:20, Admin:30)
- **medias** - Files with auto-extracted metadata
- **contents** - Dynamic content (forms, tables, playlists)
- **devices** - Connected kiosks/displays
- **jobs** - Asynchronous task queue
- **transcodes** - Video processing results

### Security Model
All resources are scoped to groups with hierarchical permissions. Device access is controlled through user ownership or group membership.

## Media Processing Flow

1. **Upload Detection** - PocketBase hooks trigger on media create/update
2. **Metadata Extraction** - FFprobe analysis for videos, image dimensions for photos
3. **Job Queue** - Asynchronous processing with progress tracking via WebSocket
4. **Transcoding** - On-demand video conversion to multiple formats/qualities

## Job System

TypeScript jobs in `/jobs/` directory execute via Deno runtime:
- **Concurrency**: Max 3 parallel jobs with semaphore control
- **Communication**: Tab-separated protocol for progress/results  
- **Isolation**: Process-level separation with comprehensive error handling
- **Utilities**: Base classes in `/jobs/utils/` for logging and API access

## Common Development Patterns

### Adding New Media Processing
Extend `medias.go` with additional metadata extraction or processing hooks.

### Creating Background Jobs  
Add TypeScript files to `/jobs/` directory following the base utility pattern.

### API Extensions
Add routes in `serve.go` or create new bind functions in the main application.

### Database Changes
Create migrations in `pb_migrations/` directory.

### Environment Configuration
- Admin credentials: `ADMIN_EMAIL` / `ADMIN_PASSWORD` in `.env`
- Production mode: `DENO_ENV=production`

## Architecture Notes

- **Group Isolation**: All data access is scoped to user's groups
- **Role-Based Access**: Hierarchical permissions with dynamic PocketBase rules
- **Real-time Updates**: WebSocket integration for live progress tracking
- **Container Limits**: 4GB RAM, 0.8 CPU for production deployment
- **Transcoding Profiles**: SD/HD/FHD/UHD with format support (H264/H265/VP8/VP9)

## File Structure Context

```
server/
├── go/cmd/pocketbase/     # Main application with custom PocketBase extensions
├── jobs/                  # TypeScript background job scripts
├── pb_data/              # PocketBase database and storage
├── pb_hooks/             # JavaScript hooks for PocketBase
├── pb_migrations/        # Database schema migrations
└── create_admin.sh       # Admin user creation script
```

When working with this codebase, always consider group-based security, async job implications, and the media processing pipeline.