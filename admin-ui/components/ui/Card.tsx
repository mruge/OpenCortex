import { ReactNode } from 'react'
import { clsx } from 'clsx'

interface CardProps {
  children: ReactNode
  className?: string
}

interface CardHeaderProps {
  children: ReactNode
  className?: string
}

interface CardTitleProps {
  children: ReactNode
  className?: string
}

interface CardDescriptionProps {
  children: ReactNode
  className?: string
}

interface CardContentProps {
  children: ReactNode
  className?: string
}

export function Card({ children, className }: CardProps) {
  return (
    <div className={clsx(
      "bg-white rounded-lg border border-gray-200 shadow-sm",
      className
    )}>
      {children}
    </div>
  )
}

export function CardHeader({ children, className }: CardHeaderProps) {
  return (
    <div className={clsx("p-6 pb-4", className)}>
      {children}
    </div>
  )
}

export function CardTitle({ children, className }: CardTitleProps) {
  return (
    <h3 className={clsx("text-lg font-semibold text-gray-900", className)}>
      {children}
    </h3>
  )
}

export function CardDescription({ children, className }: CardDescriptionProps) {
  return (
    <p className={clsx("text-sm text-gray-600 mt-1", className)}>
      {children}
    </p>
  )
}

export function CardContent({ children, className }: CardContentProps) {
  return (
    <div className={clsx("px-6 pb-6", className)}>
      {children}
    </div>
  )
}