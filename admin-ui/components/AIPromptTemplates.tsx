'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { FileText, ChevronDown, ChevronRight, Play } from 'lucide-react'

interface PromptTemplate {
  id: string
  name: string
  description: string
  category: string
  prompt: string
  systemPrompt?: string
}

interface AIPromptTemplatesProps {
  onSelectTemplate: (prompt: string, systemPrompt?: string) => void
}

const PROMPT_TEMPLATES: PromptTemplate[] = [
  {
    id: 'data-analysis',
    name: 'Data Analysis',
    description: 'Analyze patterns and trends in the current dataset',
    category: 'Analysis',
    prompt: 'Analyze the current data in the system. Look for interesting patterns, trends, and anomalies. Provide insights about relationships between entities and suggest areas for deeper investigation.',
    systemPrompt: 'You are a data analyst with expertise in graph databases and relationship analysis. Focus on actionable insights and clear explanations.'
  },
  {
    id: 'system-health',
    name: 'System Health Check',
    description: 'Evaluate the overall health and performance of the system',
    category: 'Operations',
    prompt: 'Perform a comprehensive health check of the Smart Data Abstractor system. Analyze the status of all services, identify any potential issues, and recommend optimizations.',
    systemPrompt: 'You are a system administrator monitoring a distributed microservices architecture. Focus on reliability, performance, and operational concerns.'
  },
  {
    id: 'workflow-optimization',
    name: 'Workflow Optimization',
    description: 'Suggest improvements for data processing workflows',
    category: 'Optimization',
    prompt: 'Review the current data processing workflows and suggest optimizations. Consider performance improvements, resource utilization, and workflow efficiency.',
    systemPrompt: 'You are a workflow optimization expert specializing in data pipelines and distributed processing systems.'
  },
  {
    id: 'security-audit',
    name: 'Security Audit',
    description: 'Evaluate security aspects of the system',
    category: 'Security',
    prompt: 'Conduct a security audit of the system. Examine data access patterns, identify potential vulnerabilities, and recommend security enhancements.',
    systemPrompt: 'You are a cybersecurity expert specializing in distributed systems and data security. Focus on practical security measures and risk assessment.'
  },
  {
    id: 'capability-summary',
    name: 'Capability Summary',
    description: 'Summarize all available system capabilities',
    category: 'Documentation',
    prompt: 'Provide a comprehensive summary of all capabilities available in the Smart Data Abstractor system. Include each service\'s operations, their purposes, and how they work together.',
    systemPrompt: 'You are a technical documentation specialist. Create clear, comprehensive summaries that are useful for both technical and non-technical audiences.'
  },
  {
    id: 'performance-report',
    name: 'Performance Report',
    description: 'Generate a performance analysis report',
    category: 'Reporting',
    prompt: 'Generate a detailed performance report for the system. Include metrics on response times, throughput, resource utilization, and any performance bottlenecks.',
    systemPrompt: 'You are a performance engineer creating executive-level reports. Include both technical details and business impact assessments.'
  },
  {
    id: 'data-quality',
    name: 'Data Quality Assessment',
    description: 'Assess the quality and consistency of data',
    category: 'Quality',
    prompt: 'Assess the quality of data in the system. Look for inconsistencies, missing values, data validation issues, and overall data integrity. Suggest improvements.',
    systemPrompt: 'You are a data quality specialist focused on maintaining high standards for data integrity and consistency in enterprise systems.'
  },
  {
    id: 'troubleshooting',
    name: 'System Troubleshooting',
    description: 'Help diagnose and resolve system issues',
    category: 'Support',
    prompt: 'Help me troubleshoot potential issues in the Smart Data Abstractor system. Analyze recent activity, error patterns, and system behavior to identify problems and solutions.',
    systemPrompt: 'You are a senior support engineer with deep knowledge of the system architecture. Provide step-by-step troubleshooting guidance.'
  }
]

export function AIPromptTemplates({ onSelectTemplate }: AIPromptTemplatesProps) {
  const [expandedCategory, setExpandedCategory] = useState<string | null>('Analysis')
  
  const categories = [...new Set(PROMPT_TEMPLATES.map(t => t.category))]
  
  const toggleCategory = (category: string) => {
    setExpandedCategory(expandedCategory === category ? null : category)
  }
  
  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'Analysis':
        return '📊'
      case 'Operations':
        return '⚙️'
      case 'Optimization':
        return '🚀'
      case 'Security':
        return '🔒'
      case 'Documentation':
        return '📚'
      case 'Reporting':
        return '📈'
      case 'Quality':
        return '✅'
      case 'Support':
        return '🔧'
      default:
        return '📄'
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <FileText className="h-5 w-5" />
          <span>Prompt Templates</span>
        </CardTitle>
        <CardDescription>
          Pre-built prompts for common tasks
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        {categories.map(category => (
          <div key={category}>
            <button
              onClick={() => toggleCategory(category)}
              className="flex items-center justify-between w-full p-2 text-left hover:bg-gray-50 rounded-md"
            >
              <div className="flex items-center space-x-2">
                <span>{getCategoryIcon(category)}</span>
                <span className="text-sm font-medium text-gray-900">{category}</span>
                <span className="text-xs text-gray-500">
                  ({PROMPT_TEMPLATES.filter(t => t.category === category).length})
                </span>
              </div>
              {expandedCategory === category ? (
                <ChevronDown className="h-4 w-4 text-gray-500" />
              ) : (
                <ChevronRight className="h-4 w-4 text-gray-500" />
              )}
            </button>
            
            {expandedCategory === category && (
              <div className="ml-6 space-y-2 mt-2">
                {PROMPT_TEMPLATES
                  .filter(template => template.category === category)
                  .map(template => (
                    <div
                      key={template.id}
                      className="border border-gray-200 rounded-md p-3 hover:bg-gray-50"
                    >
                      <div className="flex items-center justify-between mb-1">
                        <div className="font-medium text-sm text-gray-900">
                          {template.name}
                        </div>
                        <button
                          onClick={() => onSelectTemplate(template.prompt, template.systemPrompt)}
                          className="p-1 hover:bg-gray-200 rounded"
                          title="Use this template"
                        >
                          <Play className="h-3 w-3 text-gray-600" />
                        </button>
                      </div>
                      <div className="text-xs text-gray-600 mb-2">
                        {template.description}
                      </div>
                      <div className="text-xs text-gray-500 bg-gray-100 p-2 rounded font-mono">
                        {template.prompt.slice(0, 100)}...
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}