# Architecture

This document describes the technical architecture and design decisions of Claude Work Tracker.

## System Overview

Claude Work Tracker is a distributed system with three main components:

1. **MCP Server** - Provides API access via Model Context Protocol
2. **Work State Manager** - Handles all work item persistence and retrieval
3. **Smart Reference Engine** - AI-powered relationship detection and suggestions

```
┌─────────────────────────────────────────────────────────────┐
│                        Claude Code                           │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ /work cmds  │  │ TodoWrite    │  │ Auto-capture     │  │
│  └──────┬──────┘  └──────┬───────┘  └────────┬─────────┘  │
└─────────┼────────────────┼───────────────────┼─────────────┘
          │                │                   │
          ▼                ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                      MCP Server                              │
│  ┌────────────┐  ┌────────────────┐  ┌─────────────────┐   │
│  │   Tools    │  │    Handlers    │  │   Validation    │   │
│  │   Registry │  │  (TypeScript)  │  │   & Security    │   │
│  └──────┬─────┘  └────────┬───────┘  └────────┬────────┘   │
└─────────┼─────────────────┼───────────────────┼────────────┘
          │                 │                   │
          ▼                 ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                  Work State Manager                          │
│  ┌─────────────┐  ┌───────────────┐  ┌─────────────────┐   │
│  │   Active    │  │   Historical  │  │     Future      │   │
│  │   Storage   │  │    Archive    │  │   Work Queue    │   │
│  └─────────────┘  └───────────────┘  └─────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Smart Reference Engine                  │   │
│  │  • Similarity Analysis  • Auto-linking              │   │
│  │  • Conflict Detection   • Suggestions              │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Core Design Principles

### 1. Local-First Architecture
- Each project maintains its own `.claude-work/` directory
- No global state pollution
- Fast local file access
- Git-friendly storage format

### 2. Automatic Context Preservation
- Zero-configuration capture of work items
- Implicit git branch association
- Session-based organization
- Transparent to user workflow

### 3. Intelligent Relationship Management
- AI-powered similarity detection
- Multi-dimensional analysis
- Automatic reference generation
- Confidence-based suggestions

### 4. Progressive Disclosure
- Simple commands for basic usage
- Advanced tools available when needed
- Contextual help and suggestions
- Minimal cognitive overhead

## Component Architecture

### MCP Server (`src/index.ts`)

The entry point that implements the Model Context Protocol:

```typescript
class WorkTrackingMCPServer {
  private server: Server
  private workTrackingTools: WorkTrackingTools
  
  // Handles tool discovery
  ListToolsRequestSchema → Available tools
  
  // Handles tool execution
  CallToolRequestSchema → Tool results
}
```

**Responsibilities:**
- Protocol compliance
- Request routing
- Error handling
- Response formatting

### Work Tracking Tools (`src/tools/index.ts`)

Tool registry and handler implementation:

```typescript
class WorkTrackingTools {
  private workStateManager: WorkStateManager
  
  getTools(): Tool[]          // Tool definitions
  handleToolCall(): Response  // Tool execution
}
```

**Tool Categories:**
1. Core work management
2. Smart references
3. Historical context
4. Future work management
5. Visualization

### Work State Manager (`src/services/WorkStateManager.ts`)

Central state management and persistence:

```typescript
class WorkStateManager {
  // Storage paths
  private localWorkDir = '.claude-work/'
  private activeDir = '.claude-work/active/'
  private historyDir = '.claude-work/history/'
  private futureDir = '.claude-work/future/'
  
  // Core operations
  saveWorkItem(item: WorkItem): void
  loadActiveTodos(): WorkItem[]
  queryHistory(keyword: string): WorkItem[]
  
  // Smart features
  getContextualSuggestions(): Suggestion[]
  generateSmartReferences(itemId: string): Reference[]
}
```

**Key Features:**
- Two-tier storage (active/historical)
- Automatic metadata extraction
- Git context awareness
- Future work management

### Smart Reference Engine (`src/services/SmartReferenceEngine.ts`)

AI-powered relationship detection:

```typescript
class SmartReferenceEngine {
  // Core analysis
  calculateSemanticSimilarity(item1, item2): SimilarityScore
  generateAutomaticReferences(item): SmartReference[]
  
  // Contextual intelligence
  getContextualSuggestions(activeItems): Suggestion[]
  updateReferencesOnChange(itemId): void
}
```

**Similarity Dimensions:**
1. **Keyword Analysis** (30% weight)
   - Common term extraction
   - Stop word filtering
   - Frequency analysis

2. **Domain Alignment** (25% weight)
   - Feature domain matching
   - Technical domain correlation
   - Business area clustering

3. **Code Location** (20% weight)
   - File path analysis
   - Module detection
   - Component grouping

4. **Strategic Theme** (15% weight)
   - High-level goal alignment
   - Business value correlation
   - Initiative grouping

5. **Content Similarity** (10% weight)
   - Text comparison
   - Semantic overlap
   - Pattern matching

### Reference Mapper (`src/services/ReferenceMapper.ts`)

Visual relationship mapping:

```typescript
class ReferenceMapper {
  generateReferenceMap(): ReferenceMap
  generateFocusedMap(itemId, depth): ReferenceMap
  findReferencePath(source, target): string[]
  generateASCIIVisualization(map): string
}
```

**Visualization Features:**
- Graph-based representation
- Cluster detection
- Path finding
- ASCII output for terminal

## Data Storage

### Directory Structure

```
project/
└── .claude-work/
    ├── active/
    │   ├── current-work-context.json
    │   └── {timestamp}_{id}.json
    ├── history/
    │   └── {date}-{type}-{id}.json
    ├── future/
    │   ├── items/
    │   │   └── item-{id}.json
    │   ├── groups/
    │   │   └── {group-name}.json
    │   └── suggestions.json
    └── PENDING_TODOS.json
```

### File Formats

**Work Item:**
```json
{
  "id": "1234567890_abc123def",
  "type": "todo",
  "content": "Implement user authentication",
  "status": "in_progress",
  "context": {
    "branch": "feature-auth",
    "worktree": "main",
    "working_directory": "/path/to/project"
  },
  "session_id": "2025-07-11-auth-work",
  "timestamp": "2025-07-11T10:30:00Z",
  "metadata": {
    "priority": "high",
    "similarity_metadata": {
      "keywords": ["auth", "user", "login"],
      "feature_domain": "user-management",
      "technical_domain": "backend-api",
      "code_locations": ["src/auth/"],
      "strategic_theme": "security"
    },
    "smart_references": [{
      "target_id": "plan-auth-design",
      "similarity_score": 0.92,
      "relationship_type": "continuation",
      "confidence": 0.88,
      "auto_generated": true
    }]
  }
}
```

**Future Work Group:**
```json
{
  "id": "group-123",
  "name": "authentication-enhancements",
  "description": "Advanced auth features for phase 2",
  "items": ["item-456", "item-789"],
  "similarity_score": 0.85,
  "strategic_value": "high",
  "estimated_effort": "medium",
  "readiness_status": "blocked",
  "created_date": "2025-07-11T12:00:00Z",
  "last_updated": "2025-07-11T14:30:00Z"
}
```

## Algorithms

### Similarity Calculation

```typescript
function calculateSemanticSimilarity(item1, item2) {
  // Extract metadata
  const meta1 = extractMetadata(item1)
  const meta2 = extractMetadata(item2)
  
  // Calculate dimension scores
  const keywordScore = calculateKeywordOverlap(meta1, meta2)
  const domainScore = calculateDomainAlignment(meta1, meta2)
  const locationScore = calculateLocationSimilarity(meta1, meta2)
  const strategicScore = calculateStrategicAlignment(meta1, meta2)
  const contentScore = calculateContentSimilarity(item1, item2)
  
  // Weighted combination
  return {
    total_score: (
      keywordScore * 0.30 +
      domainScore * 0.25 +
      locationScore * 0.20 +
      strategicScore * 0.15 +
      contentScore * 0.10
    ),
    // Individual scores for transparency
    keyword_score: keywordScore,
    domain_score: domainScore,
    location_score: locationScore,
    strategic_score: strategicScore,
    content_score: contentScore
  }
}
```

### Relationship Type Detection

```typescript
function determineRelationshipType(item1, item2, similarity) {
  const content1 = item1.content.toLowerCase()
  const content2 = item2.content.toLowerCase()
  
  // Pattern matching for relationship types
  if (hasContinuationPattern(content1, content2)) {
    return 'continuation'
  }
  
  if (hasConflictPattern(content1, content2)) {
    return 'conflict'
  }
  
  if (hasDependencyPattern(content1, content2)) {
    return 'dependency'
  }
  
  return 'related'
}
```

### Confidence Calculation

```typescript
function calculateConfidence(similarity) {
  let confidence = similarity.total_score
  
  // Boost for multi-dimensional alignment
  const alignedDimensions = countAlignedDimensions(similarity)
  confidence += alignedDimensions * 0.1
  
  // Penalty for weak individual scores
  if (similarity.keyword_score < 0.2) confidence *= 0.8
  if (similarity.domain_score === 0) confidence *= 0.9
  
  return Math.min(confidence, 1.0)
}
```

## Performance Considerations

### Optimization Strategies

1. **Lazy Loading**
   - Historical items loaded on demand
   - Metadata cached in memory
   - References generated once per session

2. **Efficient Search**
   - Keyword indexing for fast lookup
   - Date-based partitioning
   - Type-filtered queries

3. **Bounded Operations**
   - Similarity calculations limited to top N items
   - Reference depth configurable
   - Visualization size constraints

### Scalability

- **Storage**: Linear with work items
- **Search**: O(n) with optimization for recent items
- **Similarity**: O(n²) worst case, O(n) with heuristics
- **Memory**: Bounded by active context size

## Security Considerations

### Data Protection
- All data stored locally
- No network transmission
- Git-ignored storage directory
- User-controlled persistence

### Input Validation
- JSON schema validation
- Path traversal prevention
- Command injection protection
- Size limits on inputs

### Error Handling
- Graceful degradation
- Non-blocking failures
- Detailed error messages
- Recovery mechanisms

## Integration Points

### Claude Code Integration
- MCP server configuration
- Slash command handling
- TodoWrite tool integration
- Automatic work capture

### Git Integration
- Branch detection
- Worktree support
- Commit association
- Repository awareness

### File System Integration
- Cross-platform paths
- Permission handling
- Atomic writes
- Directory watching

## Future Architecture Considerations

### Planned Enhancements
1. **Semantic Embeddings**: Vector-based similarity
2. **Graph Database**: Neo4j for complex relationships
3. **Real-time Sync**: Multi-device support
4. **Plugin System**: Extensible tool architecture

### Performance Improvements
1. **Incremental Indexing**: Background processing
2. **Caching Layer**: Redis for frequent queries
3. **Parallel Processing**: Worker threads
4. **Compression**: Storage optimization

### Scalability Path
1. **Sharding**: By project/date
2. **Federation**: Cross-project queries
3. **Cloud Sync**: Optional backup
4. **API Gateway**: Rate limiting

The architecture is designed to be simple, fast, and extensible while maintaining the core principle of enhancing Claude's memory without adding complexity to the user's workflow.