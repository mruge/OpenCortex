'use client'

import { useState, useEffect, useRef } from 'react'
import { Message } from '@/app/messages/page'
import { formatDistanceToNow } from 'date-fns'
import { ChevronDown, ChevronRight, Copy, Download } from 'lucide-react'
import { clsx } from 'clsx'

interface MessageBusMonitorProps {
  messages: Message[]
  onMessageReceived: (message: Message) => void
}

interface MessageItemProps {
  message: Message
}

function MessageItem({ message }: MessageItemProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [copied, setCopied] = useState(false)

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'request':
        return 'bg-blue-100 text-blue-800'
      case 'response':
        return 'bg-green-100 text-green-800'
      case 'announcement':
        return 'bg-purple-100 text-purple-800'
      case 'error':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(JSON.stringify(message, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formatData = (data: any) => {
    try {
      return JSON.stringify(data, null, 2)
    } catch {
      return String(data)
    }
  }

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-white hover:bg-gray-50 transition-colors">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="text-gray-400 hover:text-gray-600"
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
          
          <div className="flex items-center space-x-2">
            <span className={clsx(
              'px-2 py-1 text-xs font-medium rounded-full',
              getTypeColor(message.type)
            )}>
              {message.type}
            </span>
            <span className="text-sm font-medium text-gray-900">
              {message.channel}
            </span>
          </div>
        </div>

        <div className="flex items-center space-x-4 text-sm text-gray-500">
          <span>{message.service}</span>
          <span>{message.size} bytes</span>
          <span>{formatDistanceToNow(message.timestamp, { addSuffix: true })}</span>
          <button
            onClick={copyToClipboard}
            className="p-1 hover:bg-gray-200 rounded"
            title="Copy message"
          >
            <Copy className="h-4 w-4" />
          </button>
        </div>
      </div>

      {isExpanded && (
        <div className="mt-4 border-t border-gray-200 pt-4">
          <div className="space-y-2">
            <div className="text-sm">
              <span className="font-medium text-gray-700">Timestamp:</span>
              <span className="ml-2 text-gray-600">{message.timestamp.toISOString()}</span>
            </div>
            <div className="text-sm">
              <span className="font-medium text-gray-700">Message ID:</span>
              <span className="ml-2 text-gray-600 font-mono">{message.id}</span>
            </div>
          </div>
          
          <div className="mt-4">
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium text-gray-700">Data:</span>
              {copied && (
                <span className="text-xs text-green-600">Copied to clipboard!</span>
              )}
            </div>
            <pre className="bg-gray-100 rounded-md p-3 text-sm overflow-x-auto scrollbar-thin">
              <code>{formatData(message.data)}</code>
            </pre>
          </div>
        </div>
      )}
    </div>
  )
}

export function MessageBusMonitor({ messages, onMessageReceived }: MessageBusMonitorProps) {
  const [autoScroll, setAutoScroll] = useState(true)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (autoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, autoScroll])

  useEffect(() => {
    // Simulate real-time message reception
    const interval = setInterval(() => {
      const channels = ['data-requests', 'data-responses', 'ai-requests', 'ai-responses', 'exec-requests', 'exec-responses', 'workflow-requests', 'workflow-responses', 'service_capability_announcements']
      const services = ['orchestrator', 'data-abstractor', 'ai-abstractor', 'exec-agent']
      const types = ['request', 'response', 'announcement'] as const

      const mockMessage: Message = {
        id: Math.random().toString(36).substr(2, 9),
        channel: channels[Math.floor(Math.random() * channels.length)],
        timestamp: new Date(),
        service: services[Math.floor(Math.random() * services.length)],
        type: types[Math.floor(Math.random() * types.length)],
        data: {
          operation: 'test_operation',
          correlation_id: Math.random().toString(36).substr(2, 9),
          success: Math.random() > 0.2,
          timestamp: new Date().toISOString()
        },
        size: Math.floor(Math.random() * 2048) + 100
      }

      // Randomly generate messages (10% chance every second)
      if (Math.random() < 0.1) {
        onMessageReceived(mockMessage)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [onMessageReceived])

  const handleScroll = () => {
    if (containerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = containerRef.current
      const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10
      setAutoScroll(isAtBottom)
    }
  }

  const exportMessages = () => {
    const dataStr = JSON.stringify(messages, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `messages-${new Date().toISOString().slice(0, 19)}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-600">
          {messages.length} messages
        </div>
        <div className="flex items-center space-x-2">
          <label className="flex items-center space-x-2 text-sm">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="rounded"
            />
            <span>Auto-scroll</span>
          </label>
          <button
            onClick={exportMessages}
            className="flex items-center space-x-1 px-3 py-1 bg-primary-600 text-white text-sm rounded-md hover:bg-primary-700"
          >
            <Download className="h-4 w-4" />
            <span>Export</span>
          </button>
        </div>
      </div>

      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="h-96 overflow-y-auto space-y-2 scrollbar-thin"
      >
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No messages yet. Waiting for system activity...
          </div>
        ) : (
          messages.map((message) => (
            <MessageItem key={message.id} message={message} />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
    </div>
  )
}