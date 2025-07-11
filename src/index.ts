#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { 
  CallToolRequestSchema, 
  ListToolsRequestSchema,
  Tool
} from '@modelcontextprotocol/sdk/types.js'
import { WorkTrackingTools } from './tools/index.js'

class WorkTrackingMCPServer {
  private server: Server
  private workTrackingTools: WorkTrackingTools

  constructor() {
    this.workTrackingTools = new WorkTrackingTools()
    this.server = new Server(
      {
        name: 'claude-work-tracker',
        version: '1.0.0'
      },
      {
        capabilities: {
          tools: {}
        }
      }
    )

    this.setupHandlers()
  }

  private setupHandlers(): void {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      const tools = this.workTrackingTools.getTools()
      return {
        tools
      }
    })

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params
      
      try {
        const result = await this.workTrackingTools.handleToolCall(name, args || {})
        
        if (result.success) {
          return {
            content: [
              {
                type: 'text',
                text: JSON.stringify(result.data, null, 2)
              }
            ]
          }
        } else {
          return {
            content: [
              {
                type: 'text',
                text: `Error: ${result.error}`
              }
            ],
            isError: true
          }
        }
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Unexpected error: ${error instanceof Error ? error.message : String(error)}`
            }
          ],
          isError: true
        }
      }
    })
  }

  public async run(): Promise<void> {
    const transport = new StdioServerTransport()
    await this.server.connect(transport)
    
    console.error('Claude Work Tracker MCP Server running on stdio')
  }
}

// Handle uncaught errors
process.on('uncaughtException', (error) => {
  console.error('Uncaught exception:', error)
  process.exit(1)
})

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason)
  process.exit(1)
})

// Start server
const server = new WorkTrackingMCPServer()
server.run().catch(error => {
  console.error('Failed to start server:', error)
  process.exit(1)
})