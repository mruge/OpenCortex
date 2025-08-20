'use client'

import { useState, useEffect } from 'react'
import { MessageBusMonitor } from '@/components/MessageBusMonitor'
import { MessageFilters } from '@/components/MessageFilters'
import { MessageStats } from '@/components/MessageStats'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'

export interface Message {
  id: string
  channel: string
  timestamp: Date
  service: string
  type: 'request' | 'response' | 'announcement' | 'error'
  data: any
  size: number
}

export default function MessagesPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [filters, setFilters] = useState({
    channels: [] as string[],
    services: [] as string[],
    types: [] as string[],
    searchTerm: ''
  })
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    // This would be replaced with actual WebSocket connection to Redis
    const mockMessages = [
      {
        id: '1',
        channel: 'data-requests',
        timestamp: new Date(),
        service: 'orchestrator',
        type: 'request' as const,
        data: { operation: 'traverse', query: { cypher: 'MATCH (n) RETURN n LIMIT 10' } },
        size: 156
      },
      {
        id: '2',
        channel: 'data-responses',
        timestamp: new Date(Date.now() - 1000),
        service: 'data-abstractor',
        type: 'response' as const,
        data: { success: true, nodes: 10, relationships: 5 },
        size: 1024
      },
      {
        id: '3',
        channel: 'service_capability_announcements',
        timestamp: new Date(Date.now() - 2000),
        service: 'exec-agent',
        type: 'announcement' as const,
        data: { component: 'exec-agent', operations: 15, trigger: 'startup' },
        size: 2048
      }
    ]
    
    setMessages(mockMessages)
    setIsConnected(true)
  }, [])

  const filteredMessages = messages.filter(message => {
    if (filters.channels.length > 0 && !filters.channels.includes(message.channel)) {
      return false
    }
    if (filters.services.length > 0 && !filters.services.includes(message.service)) {
      return false
    }
    if (filters.types.length > 0 && !filters.types.includes(message.type)) {
      return false
    }
    if (filters.searchTerm && !JSON.stringify(message.data).toLowerCase().includes(filters.searchTerm.toLowerCase())) {
      return false
    }
    return true
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Message Bus Monitor</h1>
        <p className="text-gray-600 mt-2">
          Real-time monitoring of Redis message channels
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                Message Stream
                <div className="flex items-center space-x-2">
                  <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
                  <span className="text-sm text-gray-600">
                    {isConnected ? 'Connected' : 'Disconnected'}
                  </span>
                </div>
              </CardTitle>
              <CardDescription>
                Live messages from all Redis channels
              </CardDescription>
            </CardHeader>
            <CardContent>
              <MessageBusMonitor 
                messages={filteredMessages} 
                onMessageReceived={(message) => setMessages(prev => [message, ...prev.slice(0, 999)])}
              />
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <MessageFilters 
            filters={filters}
            onFiltersChange={setFilters}
            messages={messages}
          />
          
          <MessageStats messages={messages} />
        </div>
      </div>
    </div>
  )
}