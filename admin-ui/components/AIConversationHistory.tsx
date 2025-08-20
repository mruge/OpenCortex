'use client'

import { useEffect, useRef } from 'react'
import { AIMessage } from '@/app/ai/page'
import { formatDistanceToNow } from 'date-fns'
import { Copy, Download, Trash2, User, Bot, AlertCircle } from 'lucide-react'
import { clsx } from 'clsx'

interface AIConversationHistoryProps {
  messages: AIMessage[]
  loading: boolean
  onClear: () => void
  onExport: () => void
}

interface MessageItemProps {
  message: AIMessage
}

function MessageItem({ message }: MessageItemProps) {
  const copyToClipboard = () => {
    navigator.clipboard.writeText(message.content)
  }

  const getMessageIcon = () => {
    switch (message.type) {
      case 'user':
        return <User className="h-4 w-4" />
      case 'assistant':
        return <Bot className="h-4 w-4" />
      case 'system':
        return <AlertCircle className="h-4 w-4" />
      default:
        return null
    }
  }

  const getMessageStyle = () => {
    switch (message.type) {
      case 'user':
        return 'bg-primary-50 border-primary-200'
      case 'assistant':
        return 'bg-gray-50 border-gray-200'
      case 'system':
        return 'bg-yellow-50 border-yellow-200'
      default:
        return 'bg-white border-gray-200'
    }
  }

  const formatContent = (content: string) => {
    // Simple markdown-like formatting
    const lines = content.split('\n')
    return lines.map((line, index) => {
      if (line.startsWith('**') && line.endsWith('**')) {
        return (
          <div key={index} className="font-bold text-gray-900 mt-2 mb-1">
            {line.slice(2, -2)}
          </div>
        )
      }
      if (line.startsWith('- ')) {
        return (
          <div key={index} className="ml-4 text-gray-700">
            • {line.slice(2)}
          </div>
        )
      }
      if (line.match(/^\d+\./)) {
        return (
          <div key={index} className="ml-4 text-gray-700">
            {line}
          </div>
        )
      }
      if (line.trim() === '') {
        return <div key={index} className="h-2"></div>
      }
      return (
        <div key={index} className="text-gray-700">
          {line}
        </div>
      )
    })
  }

  return (
    <div className={clsx('border rounded-lg p-4 transition-colors', getMessageStyle())}>
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center space-x-2">
          <div className={clsx(
            'p-1 rounded-full',
            message.type === 'user' ? 'bg-primary-200 text-primary-700' :
            message.type === 'assistant' ? 'bg-gray-200 text-gray-700' :
            'bg-yellow-200 text-yellow-700'
          )}>
            {getMessageIcon()}
          </div>
          <span className="text-sm font-medium text-gray-900 capitalize">
            {message.type === 'assistant' ? 'AI Assistant' : message.type}
          </span>
          {message.model && (
            <span className="text-xs text-gray-500">({message.model})</span>
          )}
        </div>
        
        <div className="flex items-center space-x-2 text-xs text-gray-500">
          <span>{formatDistanceToNow(message.timestamp, { addSuffix: true })}</span>
          <button
            onClick={copyToClipboard}
            className="p-1 hover:bg-gray-200 rounded"
            title="Copy message"
          >
            <Copy className="h-3 w-3" />
          </button>
        </div>
      </div>
      
      <div className="prose prose-sm max-w-none">
        {formatContent(message.content)}
      </div>
      
      {(message.tokens || message.duration) && (
        <div className="flex items-center space-x-4 mt-3 pt-3 border-t border-gray-200 text-xs text-gray-500">
          {message.tokens && (
            <span>{message.tokens} tokens</span>
          )}
          {message.duration && (
            <span>{(message.duration / 1000).toFixed(1)}s</span>
          )}
        </div>
      )}
    </div>
  )
}

export function AIConversationHistory({ messages, loading, onClear, onExport }: AIConversationHistoryProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-600">
          {messages.filter(m => m.type !== 'system').length} messages
        </div>
        
        <div className="flex items-center space-x-2">
          <button
            onClick={onExport}
            className="flex items-center space-x-1 px-3 py-1 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
          >
            <Download className="h-4 w-4" />
            <span>Export</span>
          </button>
          
          <button
            onClick={onClear}
            className="flex items-center space-x-1 px-3 py-1 text-sm bg-red-600 text-white rounded-md hover:bg-red-700"
          >
            <Trash2 className="h-4 w-4" />
            <span>Clear</span>
          </button>
        </div>
      </div>

      <div className="h-96 overflow-y-auto space-y-3 scrollbar-thin">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No messages yet. Start a conversation with the AI assistant.
          </div>
        ) : (
          messages.map((message) => (
            <MessageItem key={message.id} message={message} />
          ))
        )}
        
        {loading && (
          <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
            <div className="flex items-center space-x-2">
              <div className="p-1 rounded-full bg-gray-200 text-gray-700">
                <Bot className="h-4 w-4" />
              </div>
              <span className="text-sm font-medium text-gray-900">AI Assistant</span>
            </div>
            <div className="mt-2 flex items-center space-x-2">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
              <span className="text-sm text-gray-500">Thinking...</span>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>
    </div>
  )
}