'use client'

import { useState } from 'react'
import { AIModel } from '@/app/ai/page'
import { Send, Paperclip, Mic, Settings2 } from 'lucide-react'

interface AIQueryInterfaceProps {
  onSendMessage: (content: string, systemPrompt?: string) => void
  loading: boolean
  selectedModel: AIModel | null
}

export function AIQueryInterface({ onSendMessage, loading, selectedModel }: AIQueryInterfaceProps) {
  const [message, setMessage] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [attachments, setAttachments] = useState<File[]>([])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (message.trim()) {
      onSendMessage(message.trim(), systemPrompt.trim() || undefined)
      setMessage('')
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setAttachments(prev => [...prev, ...Array.from(e.target.files!)])
    }
  }

  const removeAttachment = (index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index))
  }

  const quickPrompts = [
    "Analyze the current system performance",
    "What patterns do you see in the data?",
    "Generate a summary report of recent activity",
    "Suggest optimizations for the data pipeline",
    "Explain the relationships in the graph data"
  ]

  return (
    <div className="space-y-4">
      {/* Advanced Settings */}
      {showAdvanced && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                System Prompt (Optional)
              </label>
              <textarea
                value={systemPrompt}
                onChange={(e) => setSystemPrompt(e.target.value)}
                placeholder="Provide context or instructions for the AI..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
                rows={2}
              />
            </div>
            
            {selectedModel && (
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-600">Max Tokens:</span>
                  <span className="ml-2 font-medium">{selectedModel.maxTokens.toLocaleString()}</span>
                </div>
                <div>
                  <span className="text-gray-600">Context Window:</span>
                  <span className="ml-2 font-medium">{selectedModel.contextWindow.toLocaleString()}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Attachments */}
      {attachments.length > 0 && (
        <div className="border border-gray-200 rounded-lg p-3">
          <div className="text-sm font-medium text-gray-700 mb-2">Attachments</div>
          <div className="space-y-1">
            {attachments.map((file, index) => (
              <div key={index} className="flex items-center justify-between bg-gray-100 px-3 py-2 rounded">
                <span className="text-sm text-gray-700">{file.name}</span>
                <button
                  onClick={() => removeAttachment(index)}
                  className="text-red-500 hover:text-red-700 text-sm"
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Main Input */}
      <form onSubmit={handleSubmit} className="space-y-3">
        <div className="relative">
          <textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyPress}
            placeholder="Ask a question or request analysis..."
            className="w-full px-4 py-3 pr-20 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
            rows={3}
            disabled={loading}
          />
          
          <div className="absolute bottom-3 right-3 flex items-center space-x-2">
            <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="p-1 text-gray-400 hover:text-gray-600 rounded"
              title="Advanced settings"
            >
              <Settings2 className="h-4 w-4" />
            </button>
            
            <label className="p-1 text-gray-400 hover:text-gray-600 rounded cursor-pointer" title="Attach file">
              <Paperclip className="h-4 w-4" />
              <input
                type="file"
                multiple
                onChange={handleFileUpload}
                className="hidden"
                accept=".txt,.json,.csv,.md"
              />
            </label>
            
            <button
              type="button"
              className="p-1 text-gray-400 hover:text-gray-600 rounded"
              title="Voice input (coming soon)"
              disabled
            >
              <Mic className="h-4 w-4" />
            </button>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div className="text-xs text-gray-500">
            Press Shift+Enter for new line, Enter to send
          </div>
          
          <button
            type="submit"
            disabled={loading || !message.trim() || !selectedModel}
            className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className={`h-4 w-4 ${loading ? 'animate-pulse' : ''}`} />
            <span>{loading ? 'Sending...' : 'Send'}</span>
          </button>
        </div>
      </form>

      {/* Quick Prompts */}
      <div>
        <div className="text-sm font-medium text-gray-700 mb-2">Quick Prompts</div>
        <div className="flex flex-wrap gap-2">
          {quickPrompts.map((prompt, index) => (
            <button
              key={index}
              onClick={() => setMessage(prompt)}
              className="px-3 py-1 text-xs bg-gray-100 text-gray-700 rounded-full hover:bg-gray-200 transition-colors"
            >
              {prompt}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}