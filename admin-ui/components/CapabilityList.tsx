'use client'

import { useState } from 'react'
import { ServiceCapability, Operation } from '@/app/capabilities/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Search, Filter, Play, Clock, Shield, AlertCircle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { clsx } from 'clsx'

interface CapabilityListProps {
  capabilities: ServiceCapability[]
  filter: {
    service: string
    search: string
    onlineOnly: boolean
  }
  onFilterChange: (filter: any) => void
  onOperationSelect: (capability: ServiceCapability, operation: Operation) => void
  loading: boolean
}

interface OperationItemProps {
  capability: ServiceCapability
  operation: Operation
  onSelect: () => void
}

function OperationItem({ capability, operation, onSelect }: OperationItemProps) {
  const getOperationIcon = () => {
    if (operation.name.includes('image_')) return '🐳' // Docker container
    if (operation.name.includes('ai') || operation.name.includes('generate')) return '🤖'
    if (operation.name.includes('data') || operation.name.includes('traverse')) return '🔍'
    if (operation.name.includes('execute') || operation.name.includes('workflow')) return '⚙️'
    return '🔧'
  }

  const getServiceColor = (component: string) => {
    switch (component) {
      case 'data-abstractor':
        return 'bg-blue-100 text-blue-800'
      case 'ai-abstractor':
        return 'bg-purple-100 text-purple-800'
      case 'exec-agent':
        return 'bg-green-100 text-green-800'
      case 'orchestrator':
        return 'bg-orange-100 text-orange-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div
      onClick={onSelect}
      className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 cursor-pointer transition-colors"
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center space-x-3">
          <span className="text-2xl">{getOperationIcon()}</span>
          <div>
            <div className="font-medium text-gray-900">{operation.name}</div>
            <span className={clsx(
              'inline-block px-2 py-1 text-xs font-medium rounded-full',
              getServiceColor(capability.component)
            )}>
              {capability.component}
            </span>
          </div>
        </div>
        
        <button className="p-2 hover:bg-gray-200 rounded-full">
          <Play className="h-4 w-4 text-gray-600" />
        </button>
      </div>
      
      <div className="text-sm text-gray-600 mb-3">
        {operation.description}
      </div>
      
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-1">
            <Clock className="h-3 w-3" />
            <span>{operation.estimatedDuration}</span>
          </div>
          <div className="flex items-center space-x-1">
            <Shield className={`h-3 w-3 ${operation.retrySafe ? 'text-green-500' : 'text-yellow-500'}`} />
            <span>{operation.retrySafe ? 'Retry Safe' : 'Not Retry Safe'}</span>
          </div>
        </div>
        
        <div className="flex items-center space-x-1">
          <div className={`w-2 h-2 rounded-full ${capability.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span>{capability.isOnline ? 'Online' : 'Offline'}</span>
        </div>
      </div>
    </div>
  )
}

export function CapabilityList({ 
  capabilities, 
  filter, 
  onFilterChange, 
  onOperationSelect, 
  loading 
}: CapabilityListProps) {
  const [expandedServices, setExpandedServices] = useState<Set<string>>(new Set(['data-abstractor']))

  const toggleService = (service: string) => {
    const newExpanded = new Set(expandedServices)
    if (newExpanded.has(service)) {
      newExpanded.delete(service)
    } else {
      newExpanded.add(service)
    }
    setExpandedServices(newExpanded)
  }

  const uniqueServices = [...new Set(capabilities.map(c => c.component))]
  
  const totalOperations = capabilities.reduce((sum, cap) => sum + cap.capabilities.operations.length, 0)
  const onlineServices = capabilities.filter(cap => cap.isOnline).length

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Capabilities</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="spinner"></div>
            <span className="ml-2 text-gray-600">Loading capabilities...</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          System Capabilities
          <div className="text-sm font-normal text-gray-600">
            {totalOperations} operations across {onlineServices}/{capabilities.length} services
          </div>
        </CardTitle>
        <CardDescription>
          All available operations across the distributed system
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Filters */}
        <div className="flex flex-wrap gap-4">
          <div className="flex-1 min-w-48">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
              <input
                type="text"
                value={filter.search}
                onChange={(e) => onFilterChange({ ...filter, search: e.target.value })}
                placeholder="Search operations..."
                className="pl-10 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>
          </div>
          
          <select
            value={filter.service}
            onChange={(e) => onFilterChange({ ...filter, service: e.target.value })}
            className="px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          >
            <option value="">All Services</option>
            {uniqueServices.map(service => (
              <option key={service} value={service}>{service}</option>
            ))}
          </select>
          
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={filter.onlineOnly}
              onChange={(e) => onFilterChange({ ...filter, onlineOnly: e.target.checked })}
              className="rounded"
            />
            <span className="text-sm text-gray-700">Online Only</span>
          </label>
        </div>

        {/* Service Groups */}
        <div className="space-y-4">
          {capabilities.map((capability) => (
            <div key={capability.component} className="border border-gray-200 rounded-lg">
              <button
                onClick={() => toggleService(capability.component)}
                className="w-full flex items-center justify-between p-4 text-left hover:bg-gray-50"
              >
                <div className="flex items-center space-x-3">
                  <div className={`w-3 h-3 rounded-full ${capability.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
                  <div>
                    <div className="font-medium text-gray-900">{capability.component}</div>
                    <div className="text-sm text-gray-600">
                      {capability.capabilities.operations.length} operations
                      {capability.lastUpdated && (
                        <span className="ml-2">
                          • Updated {formatDistanceToNow(capability.lastUpdated, { addSuffix: true })}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="flex items-center space-x-2">
                  {!capability.isOnline && (
                    <AlertCircle className="h-4 w-4 text-red-500" />
                  )}
                  <Filter className={`h-4 w-4 transition-transform ${
                    expandedServices.has(capability.component) ? 'rotate-180' : ''
                  }`} />
                </div>
              </button>
              
              {expandedServices.has(capability.component) && (
                <div className="border-t border-gray-200 p-4 space-y-3">
                  {capability.capabilities.operations.length === 0 ? (
                    <div className="text-center text-gray-500 py-4">
                      No operations available
                    </div>
                  ) : (
                    capability.capabilities.operations.map((operation) => (
                      <OperationItem
                        key={operation.name}
                        capability={capability}
                        operation={operation}
                        onSelect={() => onOperationSelect(capability, operation)}
                      />
                    ))
                  )}
                </div>
              )}
            </div>
          ))}
        </div>

        {capabilities.length === 0 && (
          <div className="text-center text-gray-500 py-8">
            No capabilities found. Services may be offline or not announcing capabilities.
          </div>
        )}
      </CardContent>
    </Card>
  )
}