# Otto

An AI-powered car buying assistant that helps users negotiate and communicate with multiple car sellers simultaneously through dedicated agent-based chat threads.

## Technology Stack

**Frontend:** Next.js 14+ (App Router), TypeScript, Tailwind CSS  
**Backend:** Go 1.21+, Chi router, PostgreSQL, GORM  
**AI:** Anthropic Claude SDK

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Go 1.21+
- PostgreSQL 15+
- Anthropic API Key

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd car-buyer
   ```

2. **Set up the database**
   ```bash
   createdb carbuyer
   ```

3. **Backend setup**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # Edit .env with your configuration
   go run cmd/server/main.go
   ```
   Backend runs on `http://localhost:8080`

4. **Frontend setup**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Frontend runs on `http://localhost:3000`

## Documentation

- [QUICKSTART.md](DOCS/QUICKSTART.md) - Production deployment guide
- [DEPLOYMENT.md](DOCS/DEPLOYMENT.md) - Detailed deployment documentation
- [SPEC.md](DOCS/SPEC.md) - Technical specification and API docs

## License

Copyright (c) 2025. All rights reserved.

This software and associated documentation files (the "Software") are proprietary. Unauthorized copying, modification, distribution, or use of this Software, via any medium, is strictly prohibited.
