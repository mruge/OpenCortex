'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts'
import { TrendingUp, MessageSquare, Zap, Database } from 'lucide-react'

interface SystemStats {
  totalMessages: number
  messagesPerMinute: number
  totalOperations: number
  activeWorkflows: number
  systemUptime: string
  messageData: Array<{ time: string; count: number }>
  operationData: Array<{ service: string; count: number }>
  responseTimeData: Array<{ service: string; time: number }>
}

export function StatsGrid() {
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadStats()
    const interval = setInterval(loadStats, 10000) // Update every 10 seconds
    return () => clearInterval(interval)
  }, [])

  const loadStats = async () => {
    try {
      // This would be replaced with actual API calls
      const mockStats: SystemStats = {
        totalMessages: 15420,
        messagesPerMinute: 47,
        totalOperations: 28,
        activeWorkflows: 3,
        systemUptime: '99.9%',
        messageData: [
          { time: '10:00', count: 45 },
          { time: '10:05', count: 52 },
          { time: '10:10', count: 38 },
          { time: '10:15', count: 61 },
          { time: '10:20', count: 47 },
          { time: '10:25', count: 55 },
          { time: '10:30', count: 43 }
        ],
        operationData: [
          { service: 'data-abstractor', count: 8 },
          { service: 'ai-abstractor', count: 6 },
          { service: 'exec-agent', count: 9 },
          { service: 'orchestrator', count: 5 }
        ],
        responseTimeData: [
          { service: 'data-abstractor', time: 145 },
          { service: 'ai-abstractor', time: 890 },
          { service: 'exec-agent', time: 234 },
          { service: 'orchestrator', time: 67 }
        ]
      }
      
      setStats(mockStats)
    } catch (error) {
      console.error('Failed to load stats:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading || !stats) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map(i => (
          <Card key={i}>
            <CardContent className="p-6">
              <div className="animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-1/2 mb-2"></div>
                <div className="h-8 bg-gray-200 rounded w-3/4"></div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444']

  return (
    <div className="space-y-6">
      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Messages</p>
                <p className="text-2xl font-bold text-gray-900">{stats.totalMessages.toLocaleString()}</p>
              </div>
              <MessageSquare className="h-8 w-8 text-blue-600" />
            </div>
            <p className="text-xs text-gray-500 mt-2">
              <span className="text-green-600">↗ {stats.messagesPerMinute}/min</span> current rate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Operations</p>
                <p className="text-2xl font-bold text-gray-900">{stats.totalOperations}</p>
              </div>
              <Zap className="h-8 w-8 text-yellow-600" />
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Across all services
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Workflows</p>
                <p className="text-2xl font-bold text-gray-900">{stats.activeWorkflows}</p>
              </div>
              <TrendingUp className="h-8 w-8 text-green-600" />
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Currently executing
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">System Uptime</p>
                <p className="text-2xl font-bold text-gray-900">{stats.systemUptime}</p>
              </div>
              <Database className="h-8 w-8 text-purple-600" />
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Last 30 days
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Message Activity</CardTitle>
            <CardDescription>
              Messages per 5-minute interval
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={stats.messageData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Line 
                    type="monotone" 
                    dataKey="count" 
                    stroke="#3b82f6" 
                    strokeWidth={2}
                    dot={{ fill: '#3b82f6' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Operations by Service</CardTitle>
            <CardDescription>
              Available operations per service
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={stats.operationData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="service" 
                    tick={{ fontSize: 12 }}
                    angle={-45}
                    textAnchor="end"
                    height={60}
                  />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="count" fill="#10b981" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Response Times</CardTitle>
            <CardDescription>
              Average response time by service
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={stats.responseTimeData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="service" 
                    tick={{ fontSize: 12 }}
                    angle={-45}
                    textAnchor="end"
                    height={60}
                  />
                  <YAxis />
                  <Tooltip formatter={(value) => [`${value}ms`, 'Response Time']} />
                  <Bar dataKey="time" fill="#f59e0b" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Service Distribution</CardTitle>
            <CardDescription>
              Operations distributed across services
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={stats.operationData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ service, percent }) => `${service} ${(percent * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="count"
                  >
                    {stats.operationData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}