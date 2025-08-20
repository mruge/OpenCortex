'use client'

import { ServiceCapability } from '@/app/capabilities/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { RefreshCw, AlertTriangle, CheckCircle, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'

interface ServiceStatusProps {
  capabilities: ServiceCapability[]
  onRefresh: () => void
  loading: boolean
}

export function ServiceStatus({ capabilities, onRefresh, loading }: ServiceStatusProps) {
  const onlineServices = capabilities.filter(cap => cap.isOnline).length
  const totalOperations = capabilities.reduce((sum, cap) => sum + cap.capabilities.operations.length, 0)
  
  const getServiceHealth = (capability: ServiceCapability) => {
    if (!capability.isOnline) return 'offline'
    if (!capability.lastUpdated) return 'unknown'
    
    const timeSinceUpdate = Date.now() - capability.lastUpdated.getTime()
    if (timeSinceUpdate > 5 * 60 * 1000) return 'stale' // 5 minutes
    return 'healthy'
  }

  const getHealthColor = (health: string) => {
    switch (health) {
      case 'healthy':
        return 'text-green-600'
      case 'stale':
        return 'text-yellow-600'
      case 'offline':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  const getHealthIcon = (health: string) => {
    switch (health) {
      case 'healthy':
        return <CheckCircle className="h-4 w-4" />
      case 'stale':
        return <Clock className="h-4 w-4" />
      case 'offline':
        return <AlertTriangle className="h-4 w-4" />
      default:
        return <AlertTriangle className="h-4 w-4" />
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          Service Status Overview
          <button
            onClick={onRefresh}
            disabled={loading}
            className="flex items-center space-x-1 px-3 py-1 bg-primary-600 text-white text-sm rounded-md hover:bg-primary-700 disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </button>
        </CardTitle>
        <CardDescription>
          Real-time status of all system services and their capabilities
        </CardDescription>
      </CardHeader>
      
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900">{capabilities.length}</div>
            <div className="text-sm text-gray-600">Total Services</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{onlineServices}</div>
            <div className="text-sm text-gray-600">Online</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600">{capabilities.length - onlineServices}</div>
            <div className="text-sm text-gray-600">Offline</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{totalOperations}</div>
            <div className="text-sm text-gray-600">Operations</div>
          </div>
        </div>

        <div className="space-y-3">
          {capabilities.map((capability) => {
            const health = getServiceHealth(capability)
            return (
              <div key={capability.component} className="flex items-center justify-between p-3 border border-gray-200 rounded-lg">
                <div className="flex items-center space-x-3">
                  <div className={getHealthColor(health)}>
                    {getHealthIcon(health)}
                  </div>
                  <div>
                    <div className="font-medium text-gray-900">{capability.component}</div>
                    <div className="text-sm text-gray-600">
                      {capability.capabilities.operations.length} operations available
                    </div>
                  </div>
                </div>
                
                <div className="text-right">
                  <div className={`text-sm font-medium capitalize ${getHealthColor(health)}`}>
                    {health}
                  </div>
                  {capability.lastUpdated && (
                    <div className="text-xs text-gray-500">
                      {formatDistanceToNow(capability.lastUpdated, { addSuffix: true })}
                    </div>
                  )}
                </div>
              </div>
            )
          })}
        </div>

        {capabilities.length === 0 && (
          <div className="text-center text-gray-500 py-8">
            No services detected. Check system configuration and service health.
          </div>
        )}
      </CardContent>
    </Card>
  )
}