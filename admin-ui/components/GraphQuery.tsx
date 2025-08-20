'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Play, History, BookOpen, Copy } from 'lucide-react'

interface GraphQueryProps {
  onExecuteQuery: (query: string) => void
  queryHistory: string[]
  loading: boolean
}

const SAMPLE_QUERIES = [
  {
    name: 'All Nodes',
    query: 'MATCH (n) RETURN n LIMIT 50',
    description: 'Return all nodes in the graph'
  },
  {
    name: 'Users and Projects',
    query: 'MATCH (u:User)-[r]-(p:Project) RETURN u, r, p',
    description: 'Find relationships between users and projects'
  },
  {
    name: 'Data Flow',
    query: 'MATCH path = (d:Dataset)-[*1..3]-(a:Analysis) RETURN path',
    description: 'Trace data flow from datasets to analyses'
  },
  {
    name: 'Connected Components',
    query: 'MATCH (n)-[r]-(m) RETURN n, r, m LIMIT 100',
    description: 'Show all connected nodes and relationships'
  }
]

export function GraphQuery({ onExecuteQuery, queryHistory, loading }: GraphQueryProps) {
  const [query, setQuery] = useState('')
  const [showHistory, setShowHistory] = useState(false)
  const [showSamples, setShowSamples] = useState(false)

  const handleExecute = () => {
    if (query.trim()) {
      onExecuteQuery(query.trim())
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      handleExecute()
    }
  }

  const loadQuery = (queryText: string) => {
    setQuery(queryText)
    setShowHistory(false)
    setShowSamples(false)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Cypher Query</CardTitle>
        <CardDescription>
          Execute Cypher queries against the graph database
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Query Input */}
        <div>
          <textarea
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="MATCH (n) RETURN n LIMIT 10"
            className="w-full h-24 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none font-mono text-sm"
          />
          <div className="text-xs text-gray-500 mt-1">
            Press Ctrl+Enter to execute
          </div>
        </div>

        {/* Execute Button */}
        <button
          onClick={handleExecute}
          disabled={loading || !query.trim()}
          className="flex items-center justify-center space-x-2 w-full py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Play className={`h-4 w-4 ${loading ? 'animate-pulse' : ''}`} />
          <span>{loading ? 'Executing...' : 'Execute Query'}</span>
        </button>

        {/* Quick Actions */}
        <div className="flex space-x-2">
          <button
            onClick={() => setShowSamples(!showSamples)}
            className="flex items-center space-x-1 px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            <BookOpen className="h-4 w-4" />
            <span>Samples</span>
          </button>
          
          <button
            onClick={() => setShowHistory(!showHistory)}
            className="flex items-center space-x-1 px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          >
            <History className="h-4 w-4" />
            <span>History</span>
          </button>
        </div>

        {/* Sample Queries */}
        {showSamples && (
          <div className="border border-gray-200 rounded-md p-3 space-y-2">
            <div className="text-sm font-medium text-gray-700">Sample Queries</div>
            {SAMPLE_QUERIES.map((sample, index) => (
              <div key={index} className="border-b border-gray-100 last:border-b-0 pb-2 last:pb-0">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm font-medium text-gray-900">{sample.name}</div>
                    <div className="text-xs text-gray-600">{sample.description}</div>
                  </div>
                  <button
                    onClick={() => loadQuery(sample.query)}
                    className="p-1 hover:bg-gray-100 rounded"
                    title="Load query"
                  >
                    <Copy className="h-4 w-4 text-gray-500" />
                  </button>
                </div>
                <div className="mt-1">
                  <code className="text-xs bg-gray-100 px-2 py-1 rounded font-mono">
                    {sample.query}
                  </code>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Query History */}
        {showHistory && (
          <div className="border border-gray-200 rounded-md p-3 space-y-2">
            <div className="text-sm font-medium text-gray-700">Query History</div>
            {queryHistory.length === 0 ? (
              <div className="text-sm text-gray-500">No queries executed yet</div>
            ) : (
              <div className="space-y-1 max-h-32 overflow-y-auto">
                {queryHistory.map((historyQuery, index) => (
                  <div key={index} className="flex items-center justify-between group">
                    <code className="text-xs bg-gray-100 px-2 py-1 rounded font-mono flex-1 mr-2 truncate">
                      {historyQuery}
                    </code>
                    <button
                      onClick={() => loadQuery(historyQuery)}
                      className="opacity-0 group-hover:opacity-100 p-1 hover:bg-gray-100 rounded"
                      title="Load query"
                    >
                      <Copy className="h-3 w-3 text-gray-500" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Query Tips */}
        <div className="text-xs text-gray-500 space-y-1">
          <div className="font-medium">Tips:</div>
          <div>• Use LIMIT to control result size</div>
          <div>• Combine MATCH with WHERE for filtering</div>
          <div>• Use RETURN to specify output format</div>
        </div>
      </CardContent>
    </Card>
  )
}