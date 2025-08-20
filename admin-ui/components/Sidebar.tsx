'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { clsx } from 'clsx'
import { 
  Home, 
  MessageCircle, 
  Network, 
  Brain, 
  Settings, 
  Activity,
  Database,
  Zap
} from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'Message Bus', href: '/messages', icon: MessageCircle },
  { name: 'Graph View', href: '/graph', icon: Network },
  { name: 'AI Query', href: '/ai', icon: Brain },
  { name: 'Capabilities', href: '/capabilities', icon: Zap },
  { name: 'Data Explorer', href: '/data', icon: Database },
  { name: 'System Logs', href: '/logs', icon: Activity },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="w-64 bg-white shadow-sm border-r border-gray-200">
      <div className="flex items-center justify-center h-16 px-4 border-b border-gray-200">
        <h1 className="text-xl font-bold text-gray-900">SDA Admin</h1>
      </div>
      
      <nav className="mt-6 px-3">
        <div className="space-y-1">
          {navigation.map((item) => {
            const isActive = pathname === item.href
            return (
              <Link
                key={item.name}
                href={item.href}
                className={clsx(
                  'group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors',
                  isActive
                    ? 'bg-primary-100 text-primary-700'
                    : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                )}
              >
                <item.icon
                  className={clsx(
                    'mr-3 h-5 w-5',
                    isActive
                      ? 'text-primary-500'
                      : 'text-gray-400 group-hover:text-gray-500'
                  )}
                />
                {item.name}
              </Link>
            )
          })}
        </div>
      </nav>

      <div className="absolute bottom-0 w-64 p-4 border-t border-gray-200">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center">
              <span className="text-white text-sm font-medium">A</span>
            </div>
          </div>
          <div className="ml-3">
            <p className="text-sm font-medium text-gray-700">Admin User</p>
            <p className="text-xs text-gray-500">System Administrator</p>
          </div>
        </div>
      </div>
    </div>
  )
}