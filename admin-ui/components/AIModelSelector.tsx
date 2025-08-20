'use client'

import { AIModel } from '@/app/ai/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Brain, Check, Clock, Zap } from 'lucide-react'
import { clsx } from 'clsx'

interface AIModelSelectorProps {
  models: AIModel[]
  selectedModel: AIModel | null
  onModelSelect: (model: AIModel) => void
}

export function AIModelSelector({ models, selectedModel, onModelSelect }: AIModelSelectorProps) {
  const getProviderIcon = (provider: string) => {
    switch (provider) {
      case 'anthropic':
        return '🤖'
      case 'openai':
        return '🧠'
      default:
        return '🔮'
    }
  }

  const getModelFeatures = (model: AIModel) => {
    const features = []
    
    if (model.name.includes('Haiku') || model.name.includes('3.5')) {
      features.push({ icon: Zap, label: 'Fast', color: 'text-green-600' })
    }
    
    if (model.name.includes('Sonnet') || model.name.includes('GPT-4')) {
      features.push({ icon: Brain, label: 'Balanced', color: 'text-blue-600' })
    }
    
    if (model.contextWindow > 100000) {
      features.push({ icon: Clock, label: 'Long Context', color: 'text-purple-600' })
    }
    
    return features
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>AI Model</CardTitle>
        <CardDescription>
          Select the AI model for your queries
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {models.map((model) => (
          <div
            key={model.id}
            onClick={() => onModelSelect(model)}
            className={clsx(
              'p-3 rounded-lg border cursor-pointer transition-all',
              selectedModel?.id === model.id
                ? 'border-primary-500 bg-primary-50'
                : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50',
              !model.available && 'opacity-50 cursor-not-allowed'
            )}
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center space-x-2">
                <span className="text-lg">{getProviderIcon(model.provider)}</span>
                <div>
                  <div className="font-medium text-sm text-gray-900">{model.name}</div>
                  <div className="text-xs text-gray-600 capitalize">{model.provider}</div>
                </div>
              </div>
              
              {selectedModel?.id === model.id && (
                <Check className="h-4 w-4 text-primary-600" />
              )}
            </div>
            
            <div className="text-xs text-gray-600 mb-2">
              {model.description}
            </div>
            
            <div className="flex items-center space-x-3 text-xs text-gray-500">
              <span>{(model.maxTokens / 1000).toFixed(0)}K tokens</span>
              <span>{(model.contextWindow / 1000).toFixed(0)}K context</span>
              <div className={`w-2 h-2 rounded-full ${model.available ? 'bg-green-500' : 'bg-red-500'}`}></div>
            </div>
            
            {getModelFeatures(model).length > 0 && (
              <div className="flex items-center space-x-2 mt-2">
                {getModelFeatures(model).map((feature, index) => (
                  <div key={index} className="flex items-center space-x-1">
                    <feature.icon className={`h-3 w-3 ${feature.color}`} />
                    <span className={`text-xs ${feature.color}`}>{feature.label}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}