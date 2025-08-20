'use client'

import { useState } from 'react'
import { Message } from '@/app/messages/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Search, Filter, X } from 'lucide-react'

interface MessageFiltersProps {
  filters: {
    channels: string[]
    services: string[]
    types: string[]
    searchTerm: string
  }
  onFiltersChange: (filters: any) => void
  messages: Message[]
}

export function MessageFilters({ filters, onFiltersChange, messages }: MessageFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(true)

  // Extract unique values from messages
  const uniqueChannels = [...new Set(messages.map(m => m.channel))]
  const uniqueServices = [...new Set(messages.map(m => m.service))]
  const uniqueTypes = [...new Set(messages.map(m => m.type))]

  const updateFilter = (key: string, value: any) => {
    onFiltersChange({
      ...filters,
      [key]: value
    })
  }

  const toggleArrayFilter = (key: string, value: string) => {
    const currentArray = filters[key as keyof typeof filters] as string[]
    const newArray = currentArray.includes(value)
      ? currentArray.filter(item => item !== value)
      : [...currentArray, value]
    
    updateFilter(key, newArray)
  }

  const clearAllFilters = () => {
    onFiltersChange({
      channels: [],
      services: [],
      types: [],
      searchTerm: ''
    })
  }

  const hasActiveFilters = filters.channels.length > 0 || 
                          filters.services.length > 0 || 
                          filters.types.length > 0 || 
                          filters.searchTerm.length > 0

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Filter className="h-5 w-5" />
            <span>Filters</span>
          </div>
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="text-sm text-primary-600 hover:text-primary-700"
          >
            {isExpanded ? 'Collapse' : 'Expand'}
          </button>
        </CardTitle>
        <CardDescription>
          Filter messages by channel, service, or type
        </CardDescription>
      </CardHeader>
      
      {isExpanded && (
        <CardContent className="space-y-4">
          {/* Search */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Search Content
            </label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
              <input
                type="text"
                value={filters.searchTerm}
                onChange={(e) => updateFilter('searchTerm', e.target.value)}
                placeholder="Search message data..."
                className="pl-10 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* Channels */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Channels ({filters.channels.length} selected)
            </label>
            <div className="space-y-1 max-h-32 overflow-y-auto">
              {uniqueChannels.map(channel => (
                <label key={channel} className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={filters.channels.includes(channel)}
                    onChange={() => toggleArrayFilter('channels', channel)}
                    className="rounded"
                  />
                  <span className="text-sm text-gray-700">{channel}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Services */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Services ({filters.services.length} selected)
            </label>
            <div className="space-y-1">
              {uniqueServices.map(service => (
                <label key={service} className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={filters.services.includes(service)}
                    onChange={() => toggleArrayFilter('services', service)}
                    className="rounded"
                  />
                  <span className="text-sm text-gray-700">{service}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Types */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Message Types ({filters.types.length} selected)
            </label>
            <div className="space-y-1">
              {uniqueTypes.map(type => (
                <label key={type} className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={filters.types.includes(type)}
                    onChange={() => toggleArrayFilter('types', type)}
                    className="rounded"
                  />
                  <span className="text-sm text-gray-700 capitalize">{type}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Clear Filters */}
          {hasActiveFilters && (
            <button
              onClick={clearAllFilters}
              className="flex items-center space-x-1 w-full px-3 py-2 text-sm text-red-600 border border-red-300 rounded-md hover:bg-red-50"
            >
              <X className="h-4 w-4" />
              <span>Clear All Filters</span>
            </button>
          )}
        </CardContent>
      )}
    </Card>
  )
}