import { Check, ChevronDown } from 'lucide-react'
import { cn } from '../lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'

interface Option {
  label: string
  value: string | number
}

interface SelectProps {
  value: string | number
  onChange: (value: any) => void
  options: Option[]
  disabled?: boolean
  className?: string
}

export default function Select({ value, onChange, options, disabled, className }: SelectProps) {
  const selected = options.find((option) => String(option.value) === String(value))

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger
        disabled={disabled}
        className={cn(
          'flex h-10 w-full items-center justify-between gap-2 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm transition-colors placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/35 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-gray-100 dark:placeholder:text-gray-500',
          className,
        )}
      >
        <span className="truncate">{selected?.label ?? '请选择'}</span>
        <ChevronDown className="h-4 w-4 flex-shrink-0 opacity-50" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[var(--radix-dropdown-menu-trigger-width)]">
        {options.map((option) => {
          const active = String(option.value) === String(value)
          return (
            <DropdownMenuItem
              key={option.value}
              onSelect={() => onChange(option.value)}
              className={cn(active && 'bg-gray-100 dark:bg-white/[0.08] text-blue-600 dark:text-blue-400')}
            >
              <span className="flex h-4 w-4 items-center justify-center">
                {active && <Check className="h-4 w-4" />}
              </span>
              <span>{option.label}</span>
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
