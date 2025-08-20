'use client'

import { useMemo } from 'react'
import { GraphData, GraphNode, GraphEdge } from '@/app/graph/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Eye, Database, Network, Tag } from 'lucide-react'

interface GraphStatsProps {
  data: GraphData
  selectedNode: GraphNode | null
  selectedEdge: GraphEdge | null
}

export function GraphStats({ data, selectedNode, selectedEdge }: GraphStatsProps) {
  const stats = useMemo(() => {
    // Node label distribution
    const labelCounts = data.nodes.reduce((acc, node) => {
      node.labels.forEach(label => {
        acc[label] = (acc[label] || 0) + 1
      })
      return acc
    }, {} as Record<string, number>)

    // Relationship type distribution
    const relationshipCounts = data.edges.reduce((acc, edge) => {
      acc[edge.type] = (acc[edge.type] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    // Node degree calculation (connections per node)
    const nodeDegrees = data.nodes.reduce((acc, node) => {
      const degree = data.edges.filter(edge => 
        edge.source === node.id || edge.target === node.id
      ).length
      acc[node.id] = degree
      return acc
    }, {} as Record<string, number>)

    const avgDegree = Object.values(nodeDegrees).reduce((sum, degree) => sum + degree, 0) / data.nodes.length || 0

    return {
      labelCounts,
      relationshipCounts,
      nodeDegrees,
      avgDegree: Math.round(avgDegree * 100) / 100
    }
  }, [data])

  const formatPropertyValue = (value: any): string => {
    if (typeof value === 'object') {
      return JSON.stringify(value, null, 2)
    }
    return String(value)
  }

  return (
    <div className="space-y-6">
      {/* Selection Details */}
      {(selectedNode || selectedEdge) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Eye className="h-5 w-5" />
              <span>Selection Details</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {selectedNode && (
              <div className="space-y-3">
                <div>
                  <div className="text-sm font-medium text-gray-700">Node ID</div>
                  <div className="text-sm font-mono text-gray-900">{selectedNode.id}</div>
                </div>
                
                <div>
                  <div className="text-sm font-medium text-gray-700">Labels</div>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {selectedNode.labels.map(label => (
                      <span key={label} className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                        {label}
                      </span>
                    ))}
                  </div>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700">Properties</div>
                  <div className="mt-1 space-y-1">
                    {Object.entries(selectedNode.properties).map(([key, value]) => (
                      <div key={key} className="text-xs">
                        <span className="font-medium text-gray-600">{key}:</span>
                        <span className="ml-2 text-gray-900">{formatPropertyValue(value)}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700">Connections</div>
                  <div className="text-sm text-gray-900">{stats.nodeDegrees[selectedNode.id] || 0} relationships</div>
                </div>
              </div>
            )}

            {selectedEdge && (
              <div className="space-y-3">
                <div>
                  <div className="text-sm font-medium text-gray-700">Edge ID</div>
                  <div className="text-sm font-mono text-gray-900">{selectedEdge.id}</div>
                </div>
                
                <div>
                  <div className="text-sm font-medium text-gray-700">Type</div>
                  <div className="text-sm text-gray-900">{selectedEdge.type}</div>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700">Direction</div>
                  <div className="text-sm text-gray-900">
                    {selectedEdge.source} → {selectedEdge.target}
                  </div>
                </div>

                <div>
                  <div className="text-sm font-medium text-gray-700">Properties</div>
                  <div className="mt-1 space-y-1">
                    {Object.entries(selectedEdge.properties).map(([key, value]) => (
                      <div key={key} className="text-xs">
                        <span className="font-medium text-gray-600">{key}:</span>
                        <span className="ml-2 text-gray-900">{formatPropertyValue(value)}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Graph Statistics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Database className="h-5 w-5" />
            <span>Graph Statistics</span>
          </CardTitle>
          <CardDescription>
            Overall graph metrics and distribution
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{data.nodes.length}</div>
              <div className="text-sm text-gray-600">Nodes</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{data.edges.length}</div>
              <div className="text-sm text-gray-600">Relationships</div>
            </div>
          </div>

          <div className="text-center">
            <div className="text-lg font-bold text-gray-900">{stats.avgDegree}</div>
            <div className="text-sm text-gray-600">Average Connections</div>
          </div>
        </CardContent>
      </Card>

      {/* Node Labels */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Tag className="h-5 w-5" />
            <span>Node Labels</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Object.entries(stats.labelCounts).map(([label, count]) => (
              <div key={label} className="flex items-center justify-between">
                <span className="text-sm text-gray-700">{label}</span>
                <div className="flex items-center space-x-2">
                  <div className="w-16 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full"
                      style={{ 
                        width: `${Math.min(100, (count / Math.max(...Object.values(stats.labelCounts))) * 100)}%` 
                      }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-6 text-right">{count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Relationship Types */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Network className="h-5 w-5" />
            <span>Relationships</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Object.entries(stats.relationshipCounts).map(([type, count]) => (
              <div key={type} className="flex items-center justify-between">
                <span className="text-sm text-gray-700">{type}</span>
                <div className="flex items-center space-x-2">
                  <div className="w-16 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-600 h-2 rounded-full"
                      style={{ 
                        width: `${Math.min(100, (count / Math.max(...Object.values(stats.relationshipCounts))) * 100)}%` 
                      }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-6 text-right">{count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}