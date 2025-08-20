'use client'

import { useState, useEffect } from 'react'
import { AIQueryInterface } from '@/components/AIQueryInterface'
import { AIConversationHistory } from '@/components/AIConversationHistory'
import { AIModelSelector } from '@/components/AIModelSelector'
import { AIPromptTemplates } from '@/components/AIPromptTemplates'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'

export interface AIMessage {
  id: string
  type: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  model?: string
  tokens?: number
  duration?: number
}

export interface AIModel {
  id: string
  name: string
  provider: 'anthropic' | 'openai'
  description: string
  maxTokens: number
  contextWindow: number
  available: boolean
}

export default function AIPage() {
  const [messages, setMessages] = useState<AIMessage[]>([])
  const [selectedModel, setSelectedModel] = useState<AIModel | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [models, setModels] = useState<AIModel[]>([])

  useEffect(() => {
    // Load available AI models
    const availableModels: AIModel[] = [
      {
        id: 'claude-3-sonnet',
        name: 'Claude 3 Sonnet',
        provider: 'anthropic',
        description: 'Balanced performance for most tasks',
        maxTokens: 4096,
        contextWindow: 200000,
        available: true
      },
      {
        id: 'claude-3-haiku',
        name: 'Claude 3 Haiku',
        provider: 'anthropic', 
        description: 'Fast and efficient for simple tasks',
        maxTokens: 4096,
        contextWindow: 200000,
        available: true
      },
      {
        id: 'gpt-4',
        name: 'GPT-4',
        provider: 'openai',
        description: 'Advanced reasoning and complex tasks',
        maxTokens: 4096,
        contextWindow: 8192,
        available: true
      },
      {
        id: 'gpt-3.5-turbo',
        name: 'GPT-3.5 Turbo',
        provider: 'openai',
        description: 'Fast and cost-effective',
        maxTokens: 4096,
        contextWindow: 4096,
        available: true
      }
    ]
    
    setModels(availableModels)
    setSelectedModel(availableModels[0])

    // Add welcome message
    setMessages([
      {
        id: '1',
        type: 'system',
        content: 'Welcome to the AI Query Interface. You can ask questions about your data, request analysis, or generate content. Select a model and start chatting!',
        timestamp: new Date()
      }
    ])
  }, [])

  const handleSendMessage = async (content: string, systemPrompt?: string) => {
    if (!selectedModel) return

    const userMessage: AIMessage = {
      id: Date.now().toString(),
      type: 'user',
      content,
      timestamp: new Date(),
      model: selectedModel.name
    }

    setMessages(prev => [...prev, userMessage])
    setIsLoading(true)

    try {
      // This would be replaced with actual API call to ai-abstractor
      // const response = await fetch('/api/ai/generate', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({
      //     provider: selectedModel.provider,
      //     model: selectedModel.id,
      //     prompt: content,
      //     system_message: systemPrompt,
      //     max_tokens: selectedModel.maxTokens
      //   })
      // })

      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 1500))

      // Mock AI response
      const aiResponse: AIMessage = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: generateMockResponse(content, selectedModel),
        timestamp: new Date(),
        model: selectedModel.name,
        tokens: Math.floor(Math.random() * 500) + 100,
        duration: Math.floor(Math.random() * 3000) + 1000
      }

      setMessages(prev => [...prev, aiResponse])
    } catch (error) {
      console.error('AI request failed:', error)
      const errorMessage: AIMessage = {
        id: (Date.now() + 1).toString(),
        type: 'system',
        content: 'Sorry, there was an error processing your request. Please try again.',
        timestamp: new Date()
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  const generateMockResponse = (prompt: string, model: AIModel): string => {
    const responses = [
      `Based on the data abstractor's graph database, I can help you analyze that information. Here's what I found:

The query you're asking about would involve traversing the relationship network to identify patterns. Given the current data structure, I recommend:

1. **Data Analysis**: Use graph traversal to find connected entities
2. **Pattern Recognition**: Look for clusters and relationship types
3. **Insights Generation**: Extract meaningful trends from the connections

Would you like me to generate a specific Cypher query to explore this further?`,

      `I understand you're looking for insights from your data. Using ${model.name}, I can provide analysis across different dimensions:

**Key Findings:**
- Your data shows interesting patterns in user behavior
- There are 3 distinct clusters of activity
- Temporal analysis reveals peak usage periods

**Recommendations:**
- Consider implementing real-time monitoring for these patterns
- The exec-agent could process batch analysis jobs
- Graph relationships suggest optimization opportunities

What specific aspect would you like me to dive deeper into?`,

      `Great question! Let me analyze this using the smart data abstractor system:

**Current System State:**
- Data Abstractor: Processing graph relationships
- AI Abstractor: Available for complex analysis
- Exec Agent: Ready for container-based processing
- Orchestrator: Coordinating workflow execution

**Analysis:**
Based on the capabilities announced by each service, I can see several approaches to solve this. The most efficient would be to:

1. Use the data abstractor for initial graph traversal
2. Process results through the exec agent if computation is needed
3. Generate insights and visualizations

Would you like me to create a workflow for this analysis?`
    ]
    
    return responses[Math.floor(Math.random() * responses.length)]
  }

  const clearConversation = () => {
    setMessages([
      {
        id: '1',
        type: 'system',
        content: 'Conversation cleared. How can I help you today?',
        timestamp: new Date()
      }
    ])
  }

  const exportConversation = () => {
    const data = {
      timestamp: new Date().toISOString(),
      model: selectedModel?.name,
      messages: messages.filter(m => m.type !== 'system')
    }
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `ai-conversation-${new Date().toISOString().slice(0, 19)}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">AI Query Interface</h1>
        <p className="text-gray-600 mt-2">
          Ask questions, request analysis, or generate content using AI models
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                AI Conversation
                <div className="flex items-center space-x-2">
                  {selectedModel && (
                    <span className="text-sm text-gray-600">
                      Using {selectedModel.name}
                    </span>
                  )}
                  <div className={`w-2 h-2 rounded-full ${selectedModel?.available ? 'bg-green-500' : 'bg-red-500'}`}></div>
                </div>
              </CardTitle>
              <CardDescription>
                Interactive AI assistant for data analysis and insights
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <AIConversationHistory 
                  messages={messages}
                  loading={isLoading}
                  onClear={clearConversation}
                  onExport={exportConversation}
                />
                
                <AIQueryInterface
                  onSendMessage={handleSendMessage}
                  loading={isLoading}
                  selectedModel={selectedModel}
                />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <AIModelSelector
            models={models}
            selectedModel={selectedModel}
            onModelSelect={setSelectedModel}
          />
          
          <AIPromptTemplates
            onSelectTemplate={handleSendMessage}
          />

          {/* AI Stats */}
          <Card>
            <CardHeader>
              <CardTitle>Session Stats</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between">
                <span className="text-sm text-gray-600">Messages</span>
                <span className="text-sm font-medium">{messages.filter(m => m.type !== 'system').length}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-gray-600">Total Tokens</span>
                <span className="text-sm font-medium">
                  {messages.reduce((sum, m) => sum + (m.tokens || 0), 0)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-gray-600">Avg Response Time</span>
                <span className="text-sm font-medium">
                  {Math.round(
                    messages.filter(m => m.duration).reduce((sum, m) => sum + (m.duration || 0), 0) / 
                    Math.max(1, messages.filter(m => m.duration).length)
                  )}ms
                </span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}