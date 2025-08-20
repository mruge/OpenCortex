'use client'

import { useState, useEffect } from 'react'
import { GraphVisualization } from '@/components/GraphVisualization'
import { GraphControls } from '@/components/GraphControls'
import { GraphQuery } from '@/components/GraphQuery'
import { GraphStats } from '@/components/GraphStats'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'

export interface GraphNode {
  id: string
  label: string
  labels: string[]
  properties: Record<string, any>
  x?: number
  y?: number
}

export interface GraphEdge {
  id: string
  source: string
  target: string
  type: string
  properties: Record<string, any>
}

export interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

export default function GraphPage() {
  const [graphData, setGraphData] = useState<GraphData>({ nodes: [], edges: [] })
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null)
  const [selectedEdge, setSelectedEdge] = useState<GraphEdge | null>(null)
  const [loading, setLoading] = useState(false)
  const [queryHistory, setQueryHistory] = useState<string[]>([])

  useEffect(() => {
    // Load initial graph data
    loadSampleData()
  }, [])

  const loadSampleData = () => {
    // Sample graph data to demonstrate the interface
    const sampleData: GraphData = {
      nodes: [
        {
          id: '1',
          label: 'User:Alice',
          labels: ['User', 'Person'],
          properties: { name: 'Alice', age: 30, email: 'alice@example.com' }
        },
        {
          id: '2',
          label: 'User:Bob',
          labels: ['User', 'Person'],
          properties: { name: 'Bob', age: 25, email: 'bob@example.com' }
        },
        {
          id: '3',
          label: 'Project:DataPipeline',
          labels: ['Project'],
          properties: { name: 'Data Pipeline', status: 'active', created: '2024-01-15' }
        },
        {
          id: '4',
          label: 'Dataset:UserBehavior',
          labels: ['Dataset'],
          properties: { name: 'User Behavior Data', size: '1.2GB', format: 'JSON' }
        },
        {
          id: '5',
          label: 'Analysis:ChurnPrediction',
          labels: ['Analysis'],
          properties: { name: 'Churn Prediction', accuracy: 0.89, model: 'RandomForest' }
        }
      ],
      edges: [
        {
          id: 'e1',
          source: '1',
          target: '3',
          type: 'OWNS',
          properties: { since: '2024-01-15', role: 'creator' }
        },
        {
          id: 'e2',
          source: '2',
          target: '3',
          type: 'COLLABORATES_ON',
          properties: { since: '2024-01-20', role: 'contributor' }
        },
        {
          id: 'e3',
          source: '3',
          target: '4',
          type: 'USES',
          properties: { access_level: 'read_write' }
        },
        {
          id: 'e4',
          source: '4',
          target: '5',
          type: 'FEEDS_INTO',
          properties: { last_updated: '2024-01-25' }
        },
        {
          id: 'e5',
          source: '1',
          target: '5',
          type: 'CREATED',
          properties: { created_at: '2024-01-22' }
        }
      ]
    }
    setGraphData(sampleData)
  }

  const executeQuery = async (query: string) => {
    setLoading(true)
    setQueryHistory(prev => [query, ...prev.slice(0, 9)]) // Keep last 10 queries
    
    try {
      // This would be replaced with actual API call to data-abstractor
      // const response = await fetch('/api/data/traverse', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ cypher: query })
      // })
      
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      // For demo, just reload sample data
      loadSampleData()
      
    } catch (error) {
      console.error('Query execution failed:', error)
    } finally {
      setLoading(false)
    }
  }

  const refreshGraph = () => {
    executeQuery('MATCH (n)-[r]-(m) RETURN n, r, m LIMIT 50')
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Graph Visualization</h1>
        <p className="text-gray-600 mt-2">
          Explore and query the complete data graph from the data abstractor
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3">
          <Card>
            <CardHeader>
              <CardTitle>Graph Network</CardTitle>
              <CardDescription>
                Interactive visualization of nodes and relationships
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <GraphControls 
                  onRefresh={refreshGraph}
                  loading={loading}
                  nodeCount={graphData.nodes.length}
                  edgeCount={graphData.edges.length}
                />
                
                <div className="h-96 border border-gray-200 rounded-lg">
                  <GraphVisualization
                    data={graphData}
                    onNodeSelect={setSelectedNode}
                    onEdgeSelect={setSelectedEdge}
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <GraphQuery 
            onExecuteQuery={executeQuery}
            queryHistory={queryHistory}
            loading={loading}
          />
          
          <GraphStats 
            data={graphData}
            selectedNode={selectedNode}
            selectedEdge={selectedEdge}
          />
        </div>
      </div>
    </div>
  )
}