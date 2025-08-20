'use client'

import { useEffect, useRef, useState } from 'react'
import { GraphData, GraphNode, GraphEdge } from '@/app/graph/page'

interface GraphVisualizationProps {
  data: GraphData
  onNodeSelect: (node: GraphNode | null) => void
  onEdgeSelect: (edge: GraphEdge | null) => void
}

// Simple force-directed graph implementation
export function GraphVisualization({ data, onNodeSelect, onEdgeSelect }: GraphVisualizationProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [dragNodeId, setDragNodeId] = useState<string | null>(null)

  // Initialize node positions if not set
  useEffect(() => {
    data.nodes.forEach((node, index) => {
      if (node.x === undefined || node.y === undefined) {
        const angle = (index / data.nodes.length) * 2 * Math.PI
        const radius = Math.min(150, 50 + data.nodes.length * 10)
        node.x = 200 + Math.cos(angle) * radius
        node.y = 200 + Math.sin(angle) * radius
      }
    })
  }, [data])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const draw = () => {
      // Clear canvas
      ctx.clearRect(0, 0, canvas.width, canvas.height)

      // Draw edges
      data.edges.forEach(edge => {
        const sourceNode = data.nodes.find(n => n.id === edge.source)
        const targetNode = data.nodes.find(n => n.id === edge.target)
        
        if (!sourceNode || !targetNode) return

        ctx.beginPath()
        ctx.moveTo(sourceNode.x!, sourceNode.y!)
        ctx.lineTo(targetNode.x!, targetNode.y!)
        ctx.strokeStyle = '#9ca3af'
        ctx.lineWidth = 2
        ctx.stroke()

        // Draw edge label
        const midX = (sourceNode.x! + targetNode.x!) / 2
        const midY = (sourceNode.y! + targetNode.y!) / 2
        
        ctx.fillStyle = '#6b7280'
        ctx.font = '10px sans-serif'
        ctx.textAlign = 'center'
        ctx.fillText(edge.type, midX, midY - 5)
      })

      // Draw nodes
      data.nodes.forEach(node => {
        const isSelected = selectedNodeId === node.id
        const isHovered = hoveredNodeId === node.id
        
        // Determine node color based on labels
        let nodeColor = '#3b82f6' // default blue
        if (node.labels.includes('User') || node.labels.includes('Person')) {
          nodeColor = '#10b981' // green
        } else if (node.labels.includes('Project')) {
          nodeColor = '#f59e0b' // yellow
        } else if (node.labels.includes('Dataset')) {
          nodeColor = '#8b5cf6' // purple
        } else if (node.labels.includes('Analysis')) {
          nodeColor = '#ef4444' // red
        }

        // Draw node circle
        ctx.beginPath()
        ctx.arc(node.x!, node.y!, isSelected || isHovered ? 25 : 20, 0, 2 * Math.PI)
        ctx.fillStyle = nodeColor
        ctx.fill()
        
        if (isSelected) {
          ctx.strokeStyle = '#1d4ed8'
          ctx.lineWidth = 3
          ctx.stroke()
        }

        // Draw node label
        ctx.fillStyle = 'white'
        ctx.font = 'bold 10px sans-serif'
        ctx.textAlign = 'center'
        ctx.fillText(node.labels[0] || 'Node', node.x!, node.y! + 3)

        // Draw extended label below node
        ctx.fillStyle = '#374151'
        ctx.font = '12px sans-serif'
        const displayLabel = node.properties.name || node.label
        const truncatedLabel = displayLabel.length > 15 ? displayLabel.substring(0, 15) + '...' : displayLabel
        ctx.fillText(truncatedLabel, node.x!, node.y! + 35)
      })
    }

    draw()
  }, [data, selectedNodeId, hoveredNodeId])

  const getNodeAtPosition = (x: number, y: number): GraphNode | null => {
    return data.nodes.find(node => {
      const dx = node.x! - x
      const dy = node.y! - y
      return Math.sqrt(dx * dx + dy * dy) <= 25
    }) || null
  }

  const handleMouseMove = (event: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = event.clientX - rect.left
    const y = event.clientY - rect.top

    if (isDragging && dragNodeId) {
      const node = data.nodes.find(n => n.id === dragNodeId)
      if (node) {
        node.x = x
        node.y = y
      }
      return
    }

    const hoveredNode = getNodeAtPosition(x, y)
    setHoveredNodeId(hoveredNode?.id || null)
    canvas.style.cursor = hoveredNode ? 'pointer' : 'default'
  }

  const handleMouseDown = (event: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = event.clientX - rect.left
    const y = event.clientY - rect.top

    const clickedNode = getNodeAtPosition(x, y)
    
    if (clickedNode) {
      setSelectedNodeId(clickedNode.id)
      setDragNodeId(clickedNode.id)
      setIsDragging(true)
      onNodeSelect(clickedNode)
    } else {
      setSelectedNodeId(null)
      onNodeSelect(null)
    }
  }

  const handleMouseUp = () => {
    setIsDragging(false)
    setDragNodeId(null)
  }

  return (
    <div className="w-full h-full relative">
      <canvas
        ref={canvasRef}
        width={800}
        height={400}
        className="w-full h-full border-0"
        onMouseMove={handleMouseMove}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      />
      
      {/* Legend */}
      <div className="absolute top-4 right-4 bg-white bg-opacity-90 p-3 rounded-lg shadow-sm">
        <div className="text-sm font-medium text-gray-900 mb-2">Node Types</div>
        <div className="space-y-1 text-xs">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            <span>User/Person</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
            <span>Project</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-purple-500 rounded-full"></div>
            <span>Dataset</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-red-500 rounded-full"></div>
            <span>Analysis</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span>Other</span>
          </div>
        </div>
      </div>

      {/* Instructions */}
      <div className="absolute bottom-4 left-4 bg-white bg-opacity-90 p-3 rounded-lg shadow-sm">
        <div className="text-xs text-gray-600">
          <div>• Click nodes to select and view details</div>
          <div>• Drag nodes to reposition them</div>
          <div>• Hover to highlight connections</div>
        </div>
      </div>
    </div>
  )
}