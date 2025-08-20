'use client'

import { useState, useEffect } from 'react'
import { CapabilityList } from '@/components/CapabilityList'
import { CapabilityTrigger } from '@/components/CapabilityTrigger'
import { CapabilityDetails } from '@/components/CapabilityDetails'
import { ServiceStatus } from '@/components/ServiceStatus'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'

export interface Operation {
  name: string
  description: string
  inputExample: any
  outputExample: any
  retrySafe: boolean
  estimatedDuration: string
}

export interface ServiceCapability {
  component: string
  timestamp: string
  trigger: string
  capabilities: {
    operations: Operation[]
    message_patterns: {
      request_channel: string
      response_channel: string
      correlation_field: string
    }
  }
  lastUpdated?: Date
  isOnline?: boolean
}

export default function CapabilitiesPage() {
  const [capabilities, setCapabilities] = useState<ServiceCapability[]>([])
  const [selectedCapability, setSelectedCapability] = useState<ServiceCapability | null>(null)
  const [selectedOperation, setSelectedOperation] = useState<Operation | null>(null)
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState({
    service: '',
    search: '',
    onlineOnly: false
  })

  useEffect(() => {
    loadCapabilities()
    
    // Set up real-time updates
    const interval = setInterval(loadCapabilities, 10000) // Refresh every 10 seconds
    return () => clearInterval(interval)
  }, [])

  const loadCapabilities = async () => {
    try {
      // This would be replaced with actual API call to get capabilities
      // const response = await fetch('/api/capabilities')
      // const data = await response.json()
      
      // Mock data for demonstration
      const mockCapabilities: ServiceCapability[] = [
        {
          component: 'data-abstractor',
          timestamp: new Date().toISOString(),
          trigger: 'startup',
          capabilities: {
            operations: [
              {
                name: 'traverse',
                description: 'Execute Cypher queries against Neo4j graph database to traverse relationships and retrieve connected nodes',
                inputExample: {
                  operation: 'traverse',
                  correlation_id: 'unique-request-id',
                  query: {
                    cypher: 'MATCH (n:Person)-[r:KNOWS]->(m:Person) RETURN n, r, m LIMIT 10'
                  }
                },
                outputExample: {
                  correlation_id: 'unique-request-id',
                  success: true,
                  nodes: 10,
                  relationships: 5,
                  execution_time: '245ms'
                },
                retrySafe: true,
                estimatedDuration: '100ms-2s'
              },
              {
                name: 'search',
                description: 'Perform vector similarity search using Qdrant vector database',
                inputExample: {
                  operation: 'search',
                  correlation_id: 'search-001',
                  query: {
                    embedding: [0.1, 0.2, 0.3],
                    limit: 10,
                    threshold: 0.8
                  }
                },
                outputExample: {
                  correlation_id: 'search-001',
                  success: true,
                  results: 8,
                  execution_time: '150ms'
                },
                retrySafe: true,
                estimatedDuration: '50ms-500ms'
              },
              {
                name: 'enrich',
                description: 'Enrich graph data with additional metadata and computed properties',
                inputExample: {
                  operation: 'enrich',
                  correlation_id: 'enrich-001',
                  data: {
                    nodes: [{ id: 'node1', properties: {} }]
                  }
                },
                outputExample: {
                  correlation_id: 'enrich-001',
                  success: true,
                  enriched_count: 15,
                  execution_time: '1.2s'
                },
                retrySafe: false,
                estimatedDuration: '500ms-5s'
              }
            ],
            message_patterns: {
              request_channel: 'data-requests',
              response_channel: 'data-responses',
              correlation_field: 'correlation_id'
            }
          },
          lastUpdated: new Date(),
          isOnline: true
        },
        {
          component: 'ai-abstractor',
          timestamp: new Date().toISOString(),
          trigger: 'startup',
          capabilities: {
            operations: [
              {
                name: 'generate_content',
                description: 'Generate text content using various AI models (Anthropic Claude, OpenAI GPT)',
                inputExample: {
                  operation: 'generate_content',
                  correlation_id: 'gen-001',
                  provider: 'anthropic',
                  prompt: 'Analyze this data and provide insights',
                  model: 'claude-3-sonnet',
                  max_tokens: 1000
                },
                outputExample: {
                  correlation_id: 'gen-001',
                  success: true,
                  content: 'Generated analysis text...',
                  tokens_used: 850,
                  execution_time: '2.3s'
                },
                retrySafe: true,
                estimatedDuration: '1s-10s'
              },
              {
                name: 'generate_embeddings',
                description: 'Generate vector embeddings for text using AI models',
                inputExample: {
                  operation: 'generate_embeddings',
                  correlation_id: 'embed-001',
                  text: 'Sample text for embedding',
                  model: 'text-embedding-ada-002'
                },
                outputExample: {
                  correlation_id: 'embed-001',
                  success: true,
                  embedding: [0.1, 0.2, 0.3],
                  dimensions: 1536,
                  execution_time: '300ms'
                },
                retrySafe: true,
                estimatedDuration: '200ms-1s'
              }
            ],
            message_patterns: {
              request_channel: 'ai-requests',
              response_channel: 'ai-responses',
              correlation_field: 'correlation_id'
            }
          },
          lastUpdated: new Date(Date.now() - 30000), // 30 seconds ago
          isOnline: true
        },
        {
          component: 'exec-agent',
          timestamp: new Date().toISOString(),
          trigger: 'periodic_refresh',
          capabilities: {
            operations: [
              {
                name: 'execute_container',
                description: 'Execute Docker containers with custom configurations, data mounting, and output collection',
                inputExample: {
                  operation: 'execute_container',
                  correlation_id: 'exec-001',
                  container: {
                    image: 'python:3.9-slim',
                    command: ['python', '/workspace/script.py']
                  },
                  input: {
                    files: [{ name: 'script.py', content: 'print("Hello World")' }]
                  }
                },
                outputExample: {
                  correlation_id: 'exec-001',
                  success: true,
                  execution_id: 'exec_123456',
                  exit_code: 0,
                  output: 'Hello World',
                  duration: '2.1s'
                },
                retrySafe: false,
                estimatedDuration: '10s-10m'
              },
              {
                name: 'image_python_data_processing',
                description: '[python:3.9-slim] Process data files with pandas and numpy',
                inputExample: {
                  operation: 'image_python_data_processing',
                  correlation_id: 'py-proc-001',
                  container: {
                    image: 'python:3.9-slim',
                    command: ['/app/run.sh']
                  }
                },
                outputExample: {
                  correlation_id: 'py-proc-001',
                  success: true,
                  processed_files: 3,
                  duration: '45s'
                },
                retrySafe: true,
                estimatedDuration: '30s-5m'
              }
            ],
            message_patterns: {
              request_channel: 'exec-requests',
              response_channel: 'exec-responses',
              correlation_field: 'correlation_id'
            }
          },
          lastUpdated: new Date(Date.now() - 120000), // 2 minutes ago
          isOnline: true
        },
        {
          component: 'orchestrator',
          timestamp: new Date().toISOString(),
          trigger: 'startup',
          capabilities: {
            operations: [
              {
                name: 'execute_workflow',
                description: 'Execute complex workflows with DAG-based task dependencies and AI integration',
                inputExample: {
                  operation: 'execute_workflow',
                  correlation_id: 'workflow-001',
                  workflow_template: 'data-analysis-basic',
                  variables: {
                    query_limit: 100,
                    analysis_prompt: 'Analyze trends'
                  }
                },
                outputExample: {
                  correlation_id: 'workflow-001',
                  success: true,
                  execution_id: 'exec_workflow_001',
                  status: 'completed',
                  duration: '2m45s',
                  task_results: {}
                },
                retrySafe: false,
                estimatedDuration: '30s-60m'
              },
              {
                name: 'generate_ai_workflow',
                description: 'Generate workflow definitions using AI based on natural language descriptions',
                inputExample: {
                  operation: 'generate_ai_workflow',
                  prompt: 'Create a data analysis pipeline',
                  domain: 'data-science',
                  complexity: 'medium'
                },
                outputExample: {
                  id: 'ai-generated-12345',
                  name: 'Data Analysis Pipeline',
                  tasks: []
                },
                retrySafe: true,
                estimatedDuration: '5-30s'
              }
            ],
            message_patterns: {
              request_channel: 'workflow-requests',
              response_channel: 'workflow-responses',
              correlation_field: 'correlation_id'
            }
          },
          lastUpdated: new Date(Date.now() - 60000), // 1 minute ago
          isOnline: true
        }
      ]
      
      setCapabilities(mockCapabilities)
    } catch (error) {
      console.error('Failed to load capabilities:', error)
    } finally {
      setLoading(false)
    }
  }

  const filteredCapabilities = capabilities.filter(cap => {
    if (filter.service && cap.component !== filter.service) return false
    if (filter.onlineOnly && !cap.isOnline) return false
    if (filter.search) {
      const searchLower = filter.search.toLowerCase()
      const matchesComponent = cap.component.toLowerCase().includes(searchLower)
      const matchesOperation = cap.capabilities.operations.some(op => 
        op.name.toLowerCase().includes(searchLower) ||
        op.description.toLowerCase().includes(searchLower)
      )
      if (!matchesComponent && !matchesOperation) return false
    }
    return true
  })

  const handleOperationSelect = (capability: ServiceCapability, operation: Operation) => {
    setSelectedCapability(capability)
    setSelectedOperation(operation)
  }

  const refreshCapabilities = () => {
    setLoading(true)
    loadCapabilities()
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">System Capabilities</h1>
        <p className="text-gray-600 mt-2">
          View and test all available service operations across the system
        </p>
      </div>

      <ServiceStatus 
        capabilities={capabilities}
        onRefresh={refreshCapabilities}
        loading={loading}
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <CapabilityList
            capabilities={filteredCapabilities}
            filter={filter}
            onFilterChange={setFilter}
            onOperationSelect={handleOperationSelect}
            loading={loading}
          />
        </div>

        <div className="space-y-6">
          {selectedCapability && selectedOperation ? (
            <>
              <CapabilityDetails
                capability={selectedCapability}
                operation={selectedOperation}
              />
              
              <CapabilityTrigger
                capability={selectedCapability}
                operation={selectedOperation}
              />
            </>
          ) : (
            <Card>
              <CardHeader>
                <CardTitle>Operation Details</CardTitle>
                <CardDescription>
                  Select an operation to view details and test it
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center text-gray-500 py-8">
                  Click on any operation to view its details and test execution
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}