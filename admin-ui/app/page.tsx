import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { SystemStatus } from '@/components/SystemStatus'
import { ActivityFeed } from '@/components/ActivityFeed'
import { StatsGrid } from '@/components/StatsGrid'

export default function HomePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Smart Data Abstractor Admin</h1>
        <p className="text-gray-600 mt-2">
          Monitor and manage your distributed data processing system
        </p>
      </div>

      <SystemStatus />
      
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <StatsGrid />
        </div>
        <div>
          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>
                Latest system events and messages
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ActivityFeed limit={10} />
            </CardContent>
          </Card>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>
              Common administrative tasks
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <button className="w-full bg-primary-600 text-white px-4 py-2 rounded-md hover:bg-primary-700 transition-colors">
              Refresh All Services
            </button>
            <button className="w-full bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition-colors">
              Test System Health
            </button>
            <button className="w-full bg-orange-600 text-white px-4 py-2 rounded-md hover:bg-orange-700 transition-colors">
              View System Logs
            </button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Overview</CardTitle>
            <CardDescription>
              Current system configuration and status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-600">Data Abstractor</span>
                <span className="flex items-center">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  Online
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">AI Abstractor</span>
                <span className="flex items-center">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  Online
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Exec Agent</span>
                <span className="flex items-center">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  Online
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Orchestrator</span>
                <span className="flex items-center">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  Online
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}