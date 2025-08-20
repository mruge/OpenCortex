'use client'

import { useState, useEffect } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Activity, MessageCircle, Zap, AlertTriangle, CheckCircle } from 'lucide-react'

interface ActivityItem {
  id: string
  type: 'message' | 'capability' | 'workflow' | 'error' | 'system'
  title: string
  description: string
  timestamp: Date
  service?: string
  severity: 'info' | 'warning' | 'error' | 'success'
}

interface ActivityFeedProps {
  limit?: number
}

export function ActivityFeed({ limit = 20 }: ActivityFeedProps) {
  const [activities, setActivities] = useState<ActivityItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadActivities()
    
    // Simulate real-time activity updates
    const interval = setInterval(() => {
      if (Math.random() < 0.3) { // 30% chance of new activity
        addRandomActivity()
      }
    }, 5000)
    
    return () => clearInterval(interval)
  }, [])

  const loadActivities = () => {
    // Mock initial activities
    const mockActivities: ActivityItem[] = [
      {
        id: '1',
        type: 'capability',
        title: 'Service Capability Updated',
        description: 'exec-agent announced 15 new worker image capabilities',
        timestamp: new Date(Date.now() - 120000), // 2 minutes ago
        service: 'exec-agent',
        severity: 'success'
      },
      {
        id: '2',
        type: 'workflow',
        title: 'Workflow Completed',
        description: 'Data analysis workflow executed successfully in 2m45s',
        timestamp: new Date(Date.now() - 300000), // 5 minutes ago
        service: 'orchestrator',
        severity: 'success'
      },
      {
        id: '3',
        type: 'message',
        title: 'High Message Volume',
        description: 'Processing 150+ messages per minute on data-requests channel',
        timestamp: new Date(Date.now() - 480000), // 8 minutes ago
        service: 'data-abstractor',
        severity: 'info'
      },
      {
        id: '4',
        type: 'error',
        title: 'AI Request Timeout',
        description: 'Claude API request timed out after 30 seconds',
        timestamp: new Date(Date.now() - 720000), // 12 minutes ago
        service: 'ai-abstractor',
        severity: 'warning'
      },
      {
        id: '5',
        type: 'system',
        title: 'Service Restarted',
        description: 'data-abstractor service restarted and is now online',
        timestamp: new Date(Date.now() - 900000), // 15 minutes ago
        service: 'data-abstractor',
        severity: 'info'
      }
    ]
    
    setActivities(mockActivities)
    setLoading(false)
  }

  const addRandomActivity = () => {
    const randomActivities = [
      {
        type: 'message' as const,
        title: 'Message Spike Detected',
        description: 'Unusual activity on workflow-requests channel',
        service: 'orchestrator',
        severity: 'info' as const
      },
      {
        type: 'capability' as const,
        title: 'New Worker Image',
        description: 'tensorflow/tensorflow:latest-gpu capability discovered',
        service: 'exec-agent',
        severity: 'success' as const
      },
      {
        type: 'workflow' as const,
        title: 'Workflow Started',
        description: 'AI-generated data processing workflow initiated',
        service: 'orchestrator',
        severity: 'info' as const
      },
      {
        type: 'error' as const,
        title: 'Connection Error',
        description: 'Neo4j connection temporarily unavailable',
        service: 'data-abstractor',
        severity: 'warning' as const
      }
    ]
    
    const template = randomActivities[Math.floor(Math.random() * randomActivities.length)]
    const newActivity: ActivityItem = {
      id: Date.now().toString(),
      timestamp: new Date(),
      ...template
    }
    
    setActivities(prev => [newActivity, ...prev.slice(0, limit - 1)])
  }

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'message':
        return <MessageCircle className="h-4 w-4" />
      case 'capability':
        return <Zap className="h-4 w-4" />
      case 'workflow':
        return <Activity className="h-4 w-4" />
      case 'error':
        return <AlertTriangle className="h-4 w-4" />
      case 'system':
        return <CheckCircle className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'success':
        return 'text-green-600 bg-green-100'
      case 'warning':
        return 'text-yellow-600 bg-yellow-100'
      case 'error':
        return 'text-red-600 bg-red-100'
      case 'info':
      default:
        return 'text-blue-600 bg-blue-100'
    }
  }

  const getServiceColor = (service?: string) => {
    if (!service) return 'bg-gray-100 text-gray-800'
    
    switch (service) {
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

  if (loading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map(i => (
          <div key={i} className="animate-pulse">
            <div className="flex space-x-3">
              <div className="w-8 h-8 bg-gray-200 rounded-full"></div>
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                <div className="h-3 bg-gray-200 rounded w-1/2"></div>
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {activities.length === 0 ? (
        <div className="text-center text-gray-500 py-4">
          No recent activity
        </div>
      ) : (
        activities.map((activity) => (
          <div key={activity.id} className="flex items-start space-x-3">
            <div className={`p-2 rounded-full ${getSeverityColor(activity.severity)}`}>
              {getActivityIcon(activity.type)}
            </div>
            
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {activity.title}
                </p>
                <p className="text-xs text-gray-500">
                  {formatDistanceToNow(activity.timestamp, { addSuffix: true })}
                </p>
              </div>
              
              <p className="text-sm text-gray-600">
                {activity.description}
              </p>
              
              {activity.service && (
                <span className={`inline-block mt-1 px-2 py-1 text-xs font-medium rounded-full ${getServiceColor(activity.service)}`}>
                  {activity.service}
                </span>
              )}
            </div>
          </div>
        ))
      )}
    </div>
  )
}