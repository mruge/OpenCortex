'use client'

import { useState } from 'react'
import { ServiceCapability, Operation } from '@/app/capabilities/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Play, Settings, Save, RotateCcw, CheckCircle, XCircle, Clock } from 'lucide-react'

interface CapabilityTriggerProps {
  capability: ServiceCapability
  operation: Operation
}

interface ExecutionResult {
  success: boolean
  data?: any
  error?: string
  duration: number
  timestamp: Date
}

export function CapabilityTrigger({ capability, operation }: CapabilityTriggerProps) {
  const [requestData, setRequestData] = useState(() => 
    JSON.stringify(operation.inputExample, null, 2)
  )
  const [isExecuting, setIsExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null)
  const [executionHistory, setExecutionHistory] = useState<ExecutionResult[]>([])
  const [showAdvanced, setShowAdvanced] = useState(false)

  const handleExecute = async () => {
    if (!capability.isOnline) {
      setExecutionResult({
        success: false,
        error: 'Service is offline',
        duration: 0,
        timestamp: new Date()
      })
      return
    }

    setIsExecuting(true)
    const startTime = Date.now()

    try {
      const payload = JSON.parse(requestData)
      
      // This would be replaced with actual API call
      // const response = await fetch(`/api/${capability.component}`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(payload)
      // })
      
      // Simulate API delay and response
      await new Promise(resolve => setTimeout(resolve, Math.random() * 2000 + 500))
      
      const mockSuccess = Math.random() > 0.2 // 80% success rate
      const duration = Date.now() - startTime
      
      const result: ExecutionResult = {
        success: mockSuccess,
        duration,
        timestamp: new Date(),
        ...(mockSuccess 
          ? { data: generateMockResponse(operation, payload) }
          : { error: generateMockError(operation) }
        )
      }
      
      setExecutionResult(result)
      setExecutionHistory(prev => [result, ...prev.slice(0, 9)]) // Keep last 10 executions
      
    } catch (error) {
      const result: ExecutionResult = {
        success: false,
        error: error instanceof Error ? error.message : 'Invalid JSON format',
        duration: Date.now() - startTime,
        timestamp: new Date()
      }
      
      setExecutionResult(result)
      setExecutionHistory(prev => [result, ...prev.slice(0, 9)])
    } finally {
      setIsExecuting(false)
    }
  }

  const generateMockResponse = (operation: Operation, request: any) => {
    const baseResponse = { ...operation.outputExample }
    baseResponse.correlation_id = request.correlation_id || 'test-execution'
    baseResponse.timestamp = new Date().toISOString()
    baseResponse.execution_time = `${Math.random() * 1000 + 100}ms`
    return baseResponse
  }

  const generateMockError = (operation: Operation) => {
    const errors = [
      'Connection timeout',
      'Invalid input parameters',
      'Service temporarily unavailable',
      'Rate limit exceeded',
      'Authentication failed'
    ]
    return errors[Math.floor(Math.random() * errors.length)]
  }

  const resetToDefault = () => {
    setRequestData(JSON.stringify(operation.inputExample, null, 2))
    setExecutionResult(null)
  }

  const saveAsTemplate = () => {
    // This would save the current request as a reusable template
    console.log('Saving template...', requestData)
  }

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Play className="h-5 w-5" />
          <span>Test Execution</span>
        </CardTitle>
        <CardDescription>
          Test the {operation.name} operation with custom parameters
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Status indicator */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${capability.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
            <span className="text-sm text-gray-600">
              Service {capability.isOnline ? 'Online' : 'Offline'}
            </span>
          </div>
          
          {executionResult && (
            <div className="flex items-center space-x-1 text-sm">
              {executionResult.success ? (
                <CheckCircle className="h-4 w-4 text-green-500" />
              ) : (
                <XCircle className="h-4 w-4 text-red-500" />
              )}
              <span className="text-gray-600">
                {formatDuration(executionResult.duration)}
              </span>
            </div>
          )}
        </div>

        {/* Request Editor */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-700">
              Request Payload
            </label>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="p-1 hover:bg-gray-100 rounded text-gray-500"
              >
                <Settings className="h-4 w-4" />
              </button>
              <button
                onClick={saveAsTemplate}
                className="p-1 hover:bg-gray-100 rounded text-gray-500"
                title="Save as template"
              >
                <Save className="h-4 w-4" />
              </button>
              <button
                onClick={resetToDefault}
                className="p-1 hover:bg-gray-100 rounded text-gray-500"
                title="Reset to default"
              >
                <RotateCcw className="h-4 w-4" />
              </button>
            </div>
          </div>
          
          <textarea
            value={requestData}
            onChange={(e) => setRequestData(e.target.value)}
            className="w-full h-48 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono text-sm"
            placeholder="Enter JSON request payload..."
          />
        </div>

        {/* Advanced Options */}
        {showAdvanced && (
          <div className="border border-gray-200 rounded-lg p-3 bg-gray-50">
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Timeout (seconds)
                </label>
                <input
                  type="number"
                  defaultValue={30}
                  min={1}
                  max={300}
                  className="w-20 px-2 py-1 border border-gray-300 rounded text-sm"
                />
              </div>
              <div>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">Include debug information</span>
                </label>
              </div>
              <div>
                <label className="flex items-center space-x-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">Retry on failure</span>
                </label>
              </div>
            </div>
          </div>
        )}

        {/* Execute Button */}
        <button
          onClick={handleExecute}
          disabled={isExecuting || !capability.isOnline}
          className="flex items-center justify-center space-x-2 w-full py-3 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isExecuting ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Executing...</span>
            </>
          ) : (
            <>
              <Play className="h-4 w-4" />
              <span>Execute Operation</span>
            </>
          )}
        </button>

        {/* Execution Result */}
        {executionResult && (
          <div className={`border rounded-lg p-4 ${
            executionResult.success 
              ? 'border-green-200 bg-green-50' 
              : 'border-red-200 bg-red-50'
          }`}>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center space-x-2">
                {executionResult.success ? (
                  <CheckCircle className="h-5 w-5 text-green-500" />
                ) : (
                  <XCircle className="h-5 w-5 text-red-500" />
                )}
                <span className={`font-medium ${
                  executionResult.success ? 'text-green-900' : 'text-red-900'
                }`}>
                  {executionResult.success ? 'Success' : 'Failed'}
                </span>
              </div>
              <span className="text-sm text-gray-600">
                {formatDuration(executionResult.duration)}
              </span>
            </div>
            
            {executionResult.error ? (
              <div className="text-sm text-red-800">
                {executionResult.error}
              </div>
            ) : (
              <div className="bg-gray-900 rounded p-3 overflow-x-auto">
                <pre className="text-sm text-gray-100 font-mono">
                  <code>{JSON.stringify(executionResult.data, null, 2)}</code>
                </pre>
              </div>
            )}
          </div>
        )}

        {/* Execution History */}
        {executionHistory.length > 0 && (
          <div>
            <h4 className="font-medium text-gray-900 mb-2">Recent Executions</h4>
            <div className="space-y-2 max-h-32 overflow-y-auto">
              {executionHistory.map((result, index) => (
                <div key={index} className="flex items-center justify-between text-sm p-2 bg-gray-50 rounded">
                  <div className="flex items-center space-x-2">
                    {result.success ? (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-500" />
                    )}
                    <span className="text-gray-700">
                      {result.timestamp.toLocaleTimeString()}
                    </span>
                  </div>
                  <span className="text-gray-600">
                    {formatDuration(result.duration)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}