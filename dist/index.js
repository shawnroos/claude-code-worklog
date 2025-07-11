#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const index_js_1 = require("@modelcontextprotocol/sdk/server/index.js");
const stdio_js_1 = require("@modelcontextprotocol/sdk/server/stdio.js");
const types_js_1 = require("@modelcontextprotocol/sdk/types.js");
const index_js_2 = require("./tools/index.js");
class WorkTrackingMCPServer {
    server;
    workTrackingTools;
    constructor() {
        this.workTrackingTools = new index_js_2.WorkTrackingTools();
        this.server = new index_js_1.Server({
            name: 'claude-work-tracker',
            version: '1.0.0'
        }, {
            capabilities: {
                tools: {}
            }
        });
        this.setupHandlers();
    }
    setupHandlers() {
        // List available tools
        this.server.setRequestHandler(types_js_1.ListToolsRequestSchema, async () => {
            const tools = this.workTrackingTools.getTools();
            return {
                tools
            };
        });
        // Handle tool calls
        this.server.setRequestHandler(types_js_1.CallToolRequestSchema, async (request) => {
            const { name, arguments: args } = request.params;
            try {
                const result = await this.workTrackingTools.handleToolCall(name, args || {});
                if (result.success) {
                    return {
                        content: [
                            {
                                type: 'text',
                                text: JSON.stringify(result.data, null, 2)
                            }
                        ]
                    };
                }
                else {
                    return {
                        content: [
                            {
                                type: 'text',
                                text: `Error: ${result.error}`
                            }
                        ],
                        isError: true
                    };
                }
            }
            catch (error) {
                return {
                    content: [
                        {
                            type: 'text',
                            text: `Unexpected error: ${error instanceof Error ? error.message : String(error)}`
                        }
                    ],
                    isError: true
                };
            }
        });
    }
    async run() {
        const transport = new stdio_js_1.StdioServerTransport();
        await this.server.connect(transport);
        console.error('Claude Work Tracker MCP Server running on stdio');
    }
}
// Handle uncaught errors
process.on('uncaughtException', (error) => {
    console.error('Uncaught exception:', error);
    process.exit(1);
});
process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled rejection at:', promise, 'reason:', reason);
    process.exit(1);
});
// Start server
const server = new WorkTrackingMCPServer();
server.run().catch(error => {
    console.error('Failed to start server:', error);
    process.exit(1);
});
//# sourceMappingURL=index.js.map