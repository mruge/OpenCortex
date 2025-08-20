'use client'

import { ServiceCapability, Operation } from '@/app/capabilities/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Copy, ExternalLink, Info, Shield, Clock } from 'lucide-react'
import { useState } from 'react'

interface CapabilityDetailsProps {
  capability: ServiceCapability
  operation: Operation
}

export function CapabilityDetails({ capability, operation }: CapabilityDetailsProps) {
  const [copiedSection, setCopiedSection] = useState<string | null>(null)

  const copyToClipboard = (text: string, section: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSection(section)
    setTimeout(() => setCopiedSection(null), 2000)
  }

  const formatJSON = (obj: any) => {
    return JSON.stringify(obj, null, 2)
  }

  const getServiceDescription = (component: string) => {
    switch (component) {
      case 'data-abstractor':
        return 'Graph database operations and vector search capabilities'
      case 'ai-abstractor':
        return 'AI model inference and content generation services'
      case 'exec-agent':
        return 'Container execution and worker image orchestration'
      case 'orchestrator':
        return 'Workflow management and task coordination'
      default:
        return 'System service component'
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Info className="h-5 w-5" />
          <span>Operation Details</span>
        </CardTitle>
        <CardDescription>
          {capability.component} • {operation.name}
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-6">
        {/* Service Information */}
        <div>
          <h4 className="font-medium text-gray-900 mb-2">Service Information</h4>
          <div className="bg-gray-50 rounded-lg p-3 space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Component:</span>
              <span className="text-sm font-medium text-gray-900">{capability.component}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Description:</span>
              <span className="text-sm text-gray-900">{getServiceDescription(capability.component)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Status:</span>
              <div className="flex items-center space-x-1">
                <div className={`w-2 h-2 rounded-full ${capability.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <span className="text-sm text-gray-900">{capability.isOnline ? 'Online' : 'Offline'}</span>
              </div>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Last Updated:</span>
              <span className="text-sm text-gray-900">{capability.timestamp}</span>
            </div>
          </div>
        </div>

        {/* Operation Information */}
        <div>
          <h4 className="font-medium text-gray-900 mb-2">Operation Information</h4>
          <div className="bg-gray-50 rounded-lg p-3 space-y-2">
            <div>
              <span className="text-sm text-gray-600">Name:</span>
              <span className="ml-2 text-sm font-medium text-gray-900">{operation.name}</span>
            </div>
            <div>
              <span className="text-sm text-gray-600">Description:</span>
              <div className="mt-1 text-sm text-gray-900">{operation.description}</div>
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-1">
                <Clock className="h-4 w-4 text-gray-500" />
                <span className="text-sm text-gray-600">Duration:</span>
                <span className="text-sm text-gray-900">{operation.estimatedDuration}</span>
              </div>
              <div className="flex items-center space-x-1">
                <Shield className={`h-4 w-4 ${operation.retrySafe ? 'text-green-500' : 'text-yellow-500'}`} />
                <span className="text-sm text-gray-900">
                  {operation.retrySafe ? 'Retry Safe' : 'Not Retry Safe'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Message Patterns */}
        <div>
          <h4 className="font-medium text-gray-900 mb-2">Communication Channels</h4>
          <div className="bg-gray-50 rounded-lg p-3 space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Request Channel:</span>
              <span className="text-sm font-mono text-gray-900">
                {capability.capabilities.message_patterns.request_channel}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Response Channel:</span>
              <span className="text-sm font-mono text-gray-900">
                {capability.capabilities.message_patterns.response_channel}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Correlation Field:</span>
              <span className="text-sm font-mono text-gray-900">
                {capability.capabilities.message_patterns.correlation_field}
              </span>
            </div>
          </div>
        </div>

        {/* Input Example */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <h4 className="font-medium text-gray-900">Input Example</h4>
            <button
              onClick={() => copyToClipboard(formatJSON(operation.inputExample), 'input')}
              className="flex items-center space-x-1 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded"
            >
              <Copy className="h-3 w-3" />
              <span>{copiedSection === 'input' ? 'Copied!' : 'Copy'}</span>
            </button>
          </div>
          <div className="bg-gray-900 rounded-lg p-3 overflow-x-auto">
            <pre className="text-sm text-gray-100 font-mono">
              <code>{formatJSON(operation.inputExample)}</code>
            </pre>
          </div>
        </div>

        {/* Output Example */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <h4 className="font-medium text-gray-900">Output Example</h4>
            <button
              onClick={() => copyToClipboard(formatJSON(operation.outputExample), 'output')}
              className="flex items-center space-x-1 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded"
            >
              <Copy className="h-3 w-3" />
              <span>{copiedSection === 'output' ? 'Copied!' : 'Copy'}</span>
            </button>
          </div>
          <div className="bg-gray-900 rounded-lg p-3 overflow-x-auto">
            <pre className="text-sm text-gray-100 font-mono">
              <code>{formatJSON(operation.outputExample)}</code>
            </pre>
          </div>
        </div>

        {/* Documentation Link */}
        <div className="pt-4 border-t border-gray-200">
          <button className="flex items-center space-x-2 text-sm text-primary-600 hover:text-primary-700">
            <ExternalLink className="h-4 w-4" />
            <span>View API Documentation</span>
          </button>
        </div>
      </CardContent>
    </Card>
  )
}