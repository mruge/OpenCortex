# Smart Data Abstractor Admin Interface

A comprehensive NextJS-based admin interface for monitoring and managing the Smart Data Abstractor distributed system.

## Features

### 🏠 Dashboard
- System health monitoring
- Real-time metrics and statistics
- Activity feed with live updates
- Quick action controls

### 📨 Message Bus Monitor
- Real-time Redis message stream visualization
- Message filtering by channel, service, and type
- Message content inspection with JSON formatting
- Export capabilities for debugging

### 🕸️ Graph Visualization
- Interactive Neo4j graph exploration
- Cypher query interface with syntax examples
- Node and relationship inspection
- Graph statistics and analytics

### 🤖 AI Query Interface
- Multi-model AI chat interface (Claude, GPT-4)
- Prompt templates for common tasks
- Conversation history and export
- Model selection and configuration

### ⚡ Capabilities Management
- Real-time service capability discovery
- Operation testing and validation
- Worker image capability scanning
- Service status monitoring

## Technology Stack

- **Frontend**: Next.js 14 with TypeScript
- **Styling**: Tailwind CSS
- **Charts**: Recharts for data visualization
- **Real-time**: Socket.IO for live updates
- **Icons**: Lucide React
- **Date**: date-fns for time formatting

## Installation

1. **Clone and navigate to the admin interface:**
   ```bash
   cd /home/matthew/git/smart_data_abstractor/admin-ui
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.local.example .env.local
   ```
   
   Edit `.env.local` with your service URLs:
   ```env
   DATA_ABSTRACTOR_URL=http://localhost:8080
   AI_ABSTRACTOR_URL=http://localhost:8081
   EXEC_AGENT_URL=http://localhost:8082
   ORCHESTRATOR_URL=http://localhost:8083
   REDIS_URL=redis://localhost:6379
   ```

4. **Start the development server:**
   ```bash
   npm run dev
   ```

5. **Open in browser:**
   ```
   http://localhost:3000
   ```

## Environment Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DATA_ABSTRACTOR_URL` | Data abstractor service URL | `http://localhost:8080` |
| `AI_ABSTRACTOR_URL` | AI abstractor service URL | `http://localhost:8081` |
| `EXEC_AGENT_URL` | Exec agent service URL | `http://localhost:8082` |
| `ORCHESTRATOR_URL` | Orchestrator service URL | `http://localhost:8083` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `PORT` | Admin interface port | `3000` |

## Features Overview

### Dashboard
The main dashboard provides:
- **System Status**: Health checks for all services
- **Activity Feed**: Recent system events and notifications
- **Statistics**: Message counts, operation metrics, uptime data
- **Quick Actions**: Common administrative tasks

### Message Bus Monitor
Real-time monitoring of Redis pub/sub channels:
- **Live Stream**: See messages as they flow through the system
- **Filtering**: Filter by channel, service, message type, or content
- **Inspection**: View full message payloads with JSON formatting
- **Export**: Download message logs for analysis

### Graph Visualization
Interactive exploration of the Neo4j graph database:
- **Visual Graph**: Interactive node and relationship visualization
- **Cypher Queries**: Execute custom graph queries
- **Query Templates**: Pre-built queries for common operations
- **Statistics**: Graph metrics and relationship analysis

### AI Query Interface
Chat with AI models for system analysis:
- **Multi-Model Support**: Claude 3, GPT-4, and other models
- **Prompt Templates**: Pre-built prompts for system analysis
- **Conversation History**: Track and export chat sessions
- **System Context**: AI has knowledge of your system's capabilities

### Capabilities Management
Monitor and test all system capabilities:
- **Service Discovery**: Automatically discover service capabilities
- **Operation Testing**: Test individual operations with custom inputs
- **Worker Images**: Monitor Docker image capabilities
- **Live Status**: Real-time service health and capability updates

## API Integration

The admin interface connects to your services through:

1. **Direct HTTP APIs** for service-specific operations
2. **Redis pub/sub** for real-time message monitoring
3. **WebSocket connections** for live updates
4. **REST endpoints** for capability discovery

## Development

### Project Structure
```
admin-ui/
├── app/                    # Next.js 13+ app directory
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Dashboard page
│   ├── messages/          # Message monitoring
│   ├── graph/             # Graph visualization
│   ├── ai/                # AI chat interface
│   └── capabilities/      # Capabilities management
├── components/            # Reusable components
│   ├── ui/               # Basic UI components
│   ├── MessageBusMonitor.tsx
│   ├── GraphVisualization.tsx
│   ├── AIQueryInterface.tsx
│   └── CapabilityList.tsx
├── public/               # Static assets
└── styles/              # Global styles
```

### Adding New Features

1. **New Page**: Create in `app/` directory with `page.tsx`
2. **New Component**: Add to `components/` with TypeScript
3. **Navigation**: Update `components/Sidebar.tsx`
4. **API Integration**: Add service calls in component hooks

### Customization

The interface is designed to be easily customizable:

- **Styling**: Modify `tailwind.config.js` for theme changes
- **Components**: All components are modular and reusable
- **Data Sources**: Easy to connect to different APIs or data sources
- **Charts**: Recharts configuration for custom visualizations

## Production Deployment

### Docker Build
```bash
# Build the application
npm run build

# Start production server
npm run start
```

### Docker Container
```bash
# Build Docker image
docker build -t smart-data-abstractor-admin .

# Run container
docker run -p 3000:3000 \
  -e DATA_ABSTRACTOR_URL=http://data-abstractor:8080 \
  -e AI_ABSTRACTOR_URL=http://ai-abstractor:8081 \
  -e EXEC_AGENT_URL=http://exec-agent:8082 \
  -e ORCHESTRATOR_URL=http://orchestrator:8083 \
  -e REDIS_URL=redis://redis:6379 \
  smart-data-abstractor-admin
```

### Environment Variables for Production
- Use environment-specific URLs
- Configure proper Redis connection strings
- Set up authentication if required
- Configure CORS for API access

## Troubleshooting

### Common Issues

1. **Services Not Appearing**: Check service URLs in environment variables
2. **Messages Not Showing**: Verify Redis connection and pub/sub channels
3. **Graph Not Loading**: Confirm Neo4j database connection
4. **AI Not Responding**: Check AI service API keys and endpoints

### Debug Mode
Enable debug logging:
```bash
NODE_ENV=development npm run dev
```

### Health Checks
The interface includes built-in health checks for:
- Service connectivity
- Redis pub/sub functionality
- API endpoint availability

## Contributing

1. Follow TypeScript best practices
2. Use Tailwind CSS for styling
3. Ensure components are responsive
4. Add proper error handling
5. Include loading states for async operations

## License

This admin interface is part of the Smart Data Abstractor project.