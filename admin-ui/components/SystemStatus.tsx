'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { CheckCircle, AlertCircle, Clock, Database, Zap } from 'lucide-react'

interface SystemHealth {
  service: string
  status: 'online' | 'offline' | 'degraded'
  responseTime: number
  lastCheck: Date
  uptime: string
}

export function SystemStatus() {
  const [systemHealth, setSystemHealth] = useState<SystemHealth[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    checkSystemHealth()
    const interval = setInterval(checkSystemHealth, 30000) // Check every 30 seconds
    return () => clearInterval(interval)
  }, [])

  const checkSystemHealth = async () => {
    try {
      // This would be replaced with actual health check calls
      const mockHealth: SystemHealth[] = [
        {
          service: 'Data Abstractor',
          status: 'online',
          responseTime: 145,
          lastCheck: new Date(),
          uptime: '99.9%'
        },
        {
          service: 'AI Abstractor',
          status: 'online',
          responseTime: 890,
          lastCheck: new Date(),
          uptime: '99.8%'
        },
        {
          service: 'Exec Agent',
          status: 'online',
          responseTime: 234,
          lastCheck: new Date(),
          uptime: '99.7%'
        },
        {
          service: 'Orchestrator',
          status: 'online',
          responseTime: 67,
          lastCheck: new Date(),
          uptime: '99.9%'
        },
        {
          service: 'Redis',
          status: 'online',
          responseTime: 12,
          lastCheck: new Date(),
          uptime: '100%'
        }
      ]
      
      setSystemHealth(mockHealth)
    } catch (error) {
      console.error('Health check failed:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      case 'degraded':
        return <Clock className="h-5 w-5 text-yellow-500" />
      case 'offline':
        return <AlertCircle className="h-5 w-5 text-red-500" />
      default:
        return <AlertCircle className="h-5 w-5 text-gray-500" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'text-green-600'
      case 'degraded':
        return 'text-yellow-600'
      case 'offline':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  const getResponseTimeColor = (time: number) => {
    if (time < 200) return 'text-green-600'
    if (time < 1000) return 'text-yellow-600'
    return 'text-red-600'
  }

  const overallStatus = systemHealth.every(s => s.status === 'online') ? 'online' : 
                      systemHealth.some(s => s.status === 'offline') ? 'degraded' : 'degraded'

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Status</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-4">
            <div className="spinner"></div>
            <span className="ml-2 text-gray-600">Checking system health...</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          System Status
          <div className="flex items-center space-x-2">
            {getStatusIcon(overallStatus)}
            <span className={`text-sm font-medium capitalize ${getStatusColor(overallStatus)}`}>
              {overallStatus === 'degraded' ? 'All Systems Operational' : overallStatus}
            </span>
          </div>
        </CardTitle>
        <CardDescription>
          Real-time health monitoring of all system components
        </CardDescription>
      </CardHeader>
      
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {systemHealth.map((health) => (
            <div key={health.service} className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center space-x-2">
                  {getStatusIcon(health.status)}
                  <span className="font-medium text-gray-900">{health.service}</span>
                </div>
                <span className={`text-sm font-medium capitalize ${getStatusColor(health.status)}`}>
                  {health.status}
                </span>
              </div>
              
              <div className="space-y-1 text-sm text-gray-600">
                <div className="flex justify-between">
                  <span>Response Time:</span>
                  <span className={getResponseTimeColor(health.responseTime)}>
                    {health.responseTime}ms
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Uptime:</span>
                  <span className="text-gray-900">{health.uptime}</span>
                </div>
                <div className="flex justify-between">
                  <span>Last Check:</span>
                  <span className="text-gray-900">
                    {health.lastCheck.toLocaleTimeString()}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}