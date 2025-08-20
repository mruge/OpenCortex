'use client'

import { RefreshCw, Download, ZoomIn, ZoomOut, RotateCcw, Settings } from 'lucide-react'

interface GraphControlsProps {
  onRefresh: () => void
  loading: boolean
  nodeCount: number
  edgeCount: number
}

export function GraphControls({ onRefresh, loading, nodeCount, edgeCount }: GraphControlsProps) {
  const exportGraph = () => {
    // This would export the current graph as JSON or image
    console.log('Exporting graph...')
  }

  const resetLayout = () => {
    // This would reset the graph layout to default positions
    console.log('Resetting layout...')
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-4">
        <div className="text-sm text-gray-600">
          <span className="font-medium">{nodeCount}</span> nodes, 
          <span className="font-medium ml-1">{edgeCount}</span> relationships
        </div>
      </div>

      <div className="flex items-center space-x-2">
        <button
          onClick={onRefresh}
          disabled={loading}
          className="flex items-center space-x-1 px-3 py-2 text-sm bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>

        <button
          onClick={resetLayout}
          className="flex items-center space-x-1 px-3 py-2 text-sm bg-gray-600 text-white rounded-md hover:bg-gray-700"
        >
          <RotateCcw className="h-4 w-4" />
          <span>Reset Layout</span>
        </button>

        <button
          onClick={exportGraph}
          className="flex items-center space-x-1 px-3 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
        >
          <Download className="h-4 w-4" />
          <span>Export</span>
        </button>

        <div className="flex items-center space-x-1 border border-gray-300 rounded-md">
          <button className="p-2 hover:bg-gray-100 rounded-l-md">
            <ZoomIn className="h-4 w-4 text-gray-600" />
          </button>
          <button className="p-2 hover:bg-gray-100 rounded-r-md">
            <ZoomOut className="h-4 w-4 text-gray-600" />
          </button>
        </div>

        <button className="p-2 border border-gray-300 rounded-md hover:bg-gray-100">
          <Settings className="h-4 w-4 text-gray-600" />
        </button>
      </div>
    </div>
  )
}