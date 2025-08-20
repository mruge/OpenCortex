'use client'

import { useMemo } from 'react'
import { Message } from '@/app/messages/page'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

interface MessageStatsProps {
  messages: Message[]
}

export function MessageStats({ messages }: MessageStatsProps) {
  const stats = useMemo(() => {
    // Channel distribution
    const channelCounts = messages.reduce((acc, msg) => {
      acc[msg.channel] = (acc[msg.channel] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    const channelData = Object.entries(channelCounts).map(([channel, count]) => ({
      channel: channel.replace(/^(data-|ai-|exec-|workflow-)/, ''),
      count
    }))

    // Type distribution
    const typeCounts = messages.reduce((acc, msg) => {
      acc[msg.type] = (acc[msg.type] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    const typeData = Object.entries(typeCounts).map(([type, count]) => ({
      type,
      count
    }))

    // Service distribution
    const serviceCounts = messages.reduce((acc, msg) => {
      acc[msg.service] = (acc[msg.service] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    // Time-based stats (last hour)
    const hourAgo = new Date(Date.now() - 60 * 60 * 1000)
    const recentMessages = messages.filter(msg => msg.timestamp > hourAgo)
    
    // Size stats
    const totalSize = messages.reduce((sum, msg) => sum + msg.size, 0)
    const avgSize = messages.length > 0 ? Math.round(totalSize / messages.length) : 0

    return {
      channelData,
      typeData,
      serviceCounts,
      recentCount: recentMessages.length,
      totalSize,
      avgSize,
      totalMessages: messages.length
    }
  }, [messages])

  const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4']

  return (
    <div className="space-y-6">
      {/* Overview Stats */}
      <Card>
        <CardHeader>
          <CardTitle>Overview</CardTitle>
          <CardDescription>Message statistics</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{stats.totalMessages}</div>
              <div className="text-sm text-gray-600">Total Messages</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{stats.recentCount}</div>
              <div className="text-sm text-gray-600">Last Hour</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{(stats.totalSize / 1024).toFixed(1)}KB</div>
              <div className="text-sm text-gray-600">Total Size</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{stats.avgSize}B</div>
              <div className="text-sm text-gray-600">Avg Size</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Channel Distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Channel Activity</CardTitle>
          <CardDescription>Messages per channel</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={stats.channelData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="channel" 
                  tick={{ fontSize: 12 }}
                  angle={-45}
                  textAnchor="end"
                  height={60}
                />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="#3b82f6" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Message Type Distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Message Types</CardTitle>
          <CardDescription>Distribution by type</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={stats.typeData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ type, percent }) => `${type} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={60}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {stats.typeData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Service Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Service Activity</CardTitle>
          <CardDescription>Messages per service</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Object.entries(stats.serviceCounts).map(([service, count]) => (
              <div key={service} className="flex items-center justify-between">
                <span className="text-sm text-gray-700">{service}</span>
                <div className="flex items-center space-x-2">
                  <div className="w-16 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-primary-600 h-2 rounded-full"
                      style={{ 
                        width: `${Math.min(100, (count / Math.max(...Object.values(stats.serviceCounts))) * 100)}%` 
                      }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-8 text-right">{count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}