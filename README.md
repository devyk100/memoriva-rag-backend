# Memoriva RAG Backend

A high-performance Go backend for intelligent flashcard study session generation using RAG (Retrieval-Augmented Generation) with DeepSeek/OpenAI integration.

## Features

- **Intelligent Card Selection**: Uses LLM analysis to select optimal cards based on user prompts
- **Weakness Detection**: Prioritizes cards with poor SRS performance (high again/hard ratios)
- **Semantic Search**: Finds cards relevant to user study goals
- **Smart Repetition**: Repeats very weak cards multiple times in study sessions
- **Cost-Effective**: Optimized for DeepSeek API (cheaper than OpenAI)
- **High Performance**: Go-based for low latency and memory usage

## Architecture

```
memoriva-backend/
├── main.go              # Application entry point
├── config/              # Configuration management
├── models/              # Database and API models
├── services/            # Core business logic
│   ├── database.go      # Database operations
│   ├── llm.go          # LLM integration (DeepSeek/OpenAI)
│   ├── embedding.go    # Vector embeddings
│   └── rag.go          # RAG processing logic
├── handlers/            # HTTP request handlers
└── utils/              # Helper utilities
```

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL database (existing Memoriva database)
- DeepSeek API key (recommended) or OpenAI API key

## Quick Start

### 1. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit with your configuration
nano .env
```

Required environment variables:
```env
DATABASE_URL=postgresql://user:password@host:port/memoriva?sslmode=disable
DEEPSEEK_API_KEY=your_deepseek_api_key_here
OPENAI_API_KEY=your_openai_api_key_here  # Optional fallback
PORT=8080
```

### 2. Development (Local)

```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go
```

The server will start on `http://localhost:8080`

### 3. Production (Docker)

```bash
# Build and run with Docker Compose
sudo docker-compose up --build

# Or build Docker image manually
sudo docker build -t memoriva-rag .
sudo docker run -p 8080:8080 --env-file .env memoriva-rag
```

## API Endpoints

### Health Check
```
GET /health
```

### Study Session Processing
```
POST /api/study-sessions/process
Content-Type: application/json

{
  "sessionId": "uuid-of-study-session"
}
```

### Study Session Status
```
GET /api/study-sessions/{id}/status
```

## Integration with Frontend

The backend integrates with your Next.js frontend through the study session system:

1. **Frontend creates study session** in database with PENDING status
2. **Frontend calls** `POST /api/study-sessions/process` with session ID
3. **Backend processes** cards using RAG analysis
4. **Backend updates** session status to READY when complete
5. **Frontend polls** status endpoint until READY

## RAG Processing Flow

1. **Fetch deck data** from PostgreSQL (cards + SRS metadata)
2. **Analyze weakness patterns** using review counts (easy/hard/again ratios)
3. **LLM prompt engineering** with user's study prompt + card metadata
4. **Intelligent card selection** with possible repetition for weak cards
5. **Create ordered study collection** in database
6. **Update session status** to READY

## LLM Integration

### DeepSeek (Primary - Cost Effective)
- **Cost**: ~$0.14/1M input tokens, $0.28/1M output tokens
- **Quality**: Competitive with GPT-4 for reasoning tasks
- **Speed**: Good response times for RAG applications

### OpenAI (Fallback)
- **Embeddings**: text-embedding-3-small ($0.02/1M tokens)
- **Chat**: GPT-3.5-turbo for fallback scenarios

## Performance Optimizations

- **Concurrent Processing**: Multiple study sessions in parallel using goroutines
- **Efficient Database Queries**: Batch operations with GORM
- **Smart API Usage**: Minimize LLM calls while maximizing quality
- **Fallback Logic**: Intelligent card selection even without LLM

## Database Schema

The backend works with your existing Prisma schema:
- `StudySession` - Study session metadata
- `StudySessionCard` - Ordered card collections
- `SRSCardMetadata` - Review performance data
- `Flashcard` - Card content

## Deployment

### Railway/Render (Recommended)
1. Connect GitHub repository
2. Set environment variables
3. Deploy automatically

### Docker Deployment
```bash
# Production build
sudo docker build -t memoriva-rag .

# Run with environment
sudo docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="your_db_url" \
  -e DEEPSEEK_API_KEY="your_key" \
  memoriva-rag
```

## Development

### Adding New Features
1. **Models**: Add to `models/models.go`
2. **Database**: Add methods to `services/database.go`
3. **Business Logic**: Add to appropriate service
4. **API**: Add handlers to `handlers/`
5. **Routes**: Register in `main.go`

### Testing
```bash
# Run tests
go test ./...

# Test specific package
go test ./services
```

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Logs
```bash
# Docker logs
sudo docker logs memoriva-rag

# Local logs
go run main.go
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check DATABASE_URL format
   - Ensure PostgreSQL is accessible
   - Verify credentials

2. **LLM API Errors**
   - Check API keys are valid
   - Verify API quotas/limits
   - Check network connectivity

3. **Build Errors**
   - Run `go mod tidy`
   - Check Go version (1.21+)
   - Verify all dependencies

### Debug Mode
```bash
# Enable debug logging
export GIN_MODE=debug
go run main.go
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Submit pull request

## License

MIT License - see LICENSE file for details
