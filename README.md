# Support Chatbot with RAG

A support chatbot that uses Retrieval Augmented Generation (RAG) to provide accurate answers based on uploaded manuals and policies. The system processes PDF and text documents, stores their embeddings in a Qdrant vector database, and uses OpenAI's API to generate responses with citations.

## Features

- Upload PDF and text documents
- Automatic text extraction and chunking
- Vector storage using Qdrant
- Semantic search for relevant context
- OpenAI-powered responses with citations
- RESTful API endpoints

## Prerequisites

- Go 1.21 or later
- Qdrant vector database
- OpenAI API key

## Setup

1. Clone the repository:
```bash
git clone https://github.com/bduff/SupportChatbot.git
cd SupportChatbot
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export OPENAI_API_KEY=your_openai_api_key
export QDRANT_HOST=localhost  # or your Qdrant host
export UPLOAD_DIR=uploads     # directory for uploaded files
```

4. Start Qdrant:
```bash
docker run -p 6333:6333 qdrant/qdrant
```

5. Run the application:
```bash
go run main.go
```

## API Endpoints

### Upload Document
```http
POST /upload
Content-Type: multipart/form-data

file: <file>
```

### Chat
```http
POST /chat
Content-Type: application/json

{
    "message": "Your question here"
}
```

## Configuration

The following environment variables can be configured:

- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `QDRANT_HOST`: Qdrant host address (default: localhost)
- `QDRANT_PORT`: Qdrant port (default: 6333)
- `UPLOAD_DIR`: Directory for uploaded files (default: uploads)
- `CHUNK_SIZE`: Size of text chunks in words (default: 1000)
- `CHUNK_OVERLAP`: Overlap between chunks in words (default: 200)

## License

MIT
