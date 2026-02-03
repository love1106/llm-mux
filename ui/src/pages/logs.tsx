import { useState, useEffect, useRef, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { managementApi } from '@/lib/api'
import { toast } from '@/components/ui/toast'
import { RefreshCw, Trash2, Download, Search, Pause, Play, AlertTriangle, Info, Bug, XCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error'

interface ParsedLog {
  timestamp: string
  level: LogLevel
  message: string
  raw: string
}

const LOG_LEVEL_CONFIG: Record<LogLevel, { icon: React.ReactNode; color: string; bg: string }> = {
  all: { icon: null, color: '', bg: '' },
  debug: { icon: <Bug className="h-3 w-3" />, color: 'text-gray-500', bg: 'bg-gray-100 dark:bg-gray-800' },
  info: { icon: <Info className="h-3 w-3" />, color: 'text-blue-500', bg: 'bg-blue-50 dark:bg-blue-900/20' },
  warn: {
    icon: <AlertTriangle className="h-3 w-3" />,
    color: 'text-yellow-500',
    bg: 'bg-yellow-50 dark:bg-yellow-900/20',
  },
  error: { icon: <XCircle className="h-3 w-3" />, color: 'text-red-500', bg: 'bg-red-50 dark:bg-red-900/20' },
}

function utcToLocal(utcTimestamp: string): string {
  if (!utcTimestamp) return ''
  const date = new Date(utcTimestamp.replace(' ', 'T') + 'Z')
  if (isNaN(date.getTime())) return utcTimestamp
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function parseLogLine(line: string): ParsedLog {
  const timestampMatch = line.match(/^\[(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\]/)
  const levelMatch = line.match(/\[(debug|info|warn|warning|error)\]/i)

  let level: LogLevel = 'info'
  if (levelMatch) {
    const levelStr = levelMatch[1].toLowerCase()
    if (levelStr === 'warning') level = 'warn'
    else if (['debug', 'info', 'warn', 'error'].includes(levelStr)) level = levelStr as LogLevel
  }

  return {
    timestamp: timestampMatch ? utcToLocal(timestampMatch[1]) : '',
    level,
    message: line,
    raw: line,
  }
}

export function LogsPage() {
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [autoScroll, setAutoScroll] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [logLevel, setLogLevel] = useState<LogLevel>('all')
  const [limit, setLimit] = useState(100)
  const containerRef = useRef<HTMLDivElement>(null)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['logs', limit],
    queryFn: () => managementApi.getLogs({ limit }),
    refetchInterval: autoRefresh ? 3000 : false,
  })

  const lines = data?.data?.data?.lines || []

  const parsedLogs = useMemo(() => lines.map(parseLogLine), [lines])

  const filteredLogs = useMemo(() => {
    return parsedLogs
      .filter((log) => {
        if (logLevel !== 'all' && log.level !== logLevel) return false
        if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase())) return false
        return true
      })
      .reverse()
  }, [parsedLogs, logLevel, searchTerm])

  const levelCounts = useMemo(() => {
    return parsedLogs.reduce(
      (acc, log) => {
        acc[log.level] = (acc[log.level] || 0) + 1
        return acc
      },
      {} as Record<string, number>
    )
  }, [parsedLogs])

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [filteredLogs, autoScroll])

  const handleClear = async () => {
    if (!confirm('Are you sure you want to clear all logs?')) return
    try {
      await managementApi.deleteLogs()
      refetch()
      toast.success('Logs cleared')
    } catch {
      toast.error('Failed to clear logs')
    }
  }

  const handleDownload = () => {
    const content = filteredLogs.map((log) => log.raw).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `llm-mux-logs-${new Date().toISOString().slice(0, 10)}.txt`
    a.click()
    URL.revokeObjectURL(url)
    toast.success('Logs downloaded')
  }

  const highlightSearch = (text: string) => {
    if (!searchTerm) return text

    const parts = text.split(new RegExp(`(${searchTerm})`, 'gi'))
    return parts.map((part, i) =>
      part.toLowerCase() === searchTerm.toLowerCase() ? (
        <mark key={i} className="bg-yellow-300 dark:bg-yellow-600 px-0.5 rounded">
          {part}
        </mark>
      ) : (
        part
      )
    )
  }

  return (
    <div className="space-y-6 h-full flex flex-col">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Logs</h2>
          <p className="text-muted-foreground">Server request logs</p>
        </div>
        <div className="flex gap-2">
          <Button variant={autoRefresh ? 'default' : 'outline'} size="sm" onClick={() => setAutoRefresh(!autoRefresh)}>
            {autoRefresh ? <Pause className="h-4 w-4 mr-2" /> : <Play className="h-4 w-4 mr-2" />}
            {autoRefresh ? 'Pause' : 'Resume'}
          </Button>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className={cn('h-4 w-4 mr-2', autoRefresh && 'animate-spin')} />
            Refresh
          </Button>
          <Button variant="outline" size="sm" onClick={handleDownload} disabled={filteredLogs.length === 0}>
            <Download className="h-4 w-4 mr-2" />
            Download
          </Button>
          <Button variant="destructive" size="sm" onClick={handleClear}>
            <Trash2 className="h-4 w-4 mr-2" />
            Clear
          </Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
            <div className="relative md:col-span-2">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search logs..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-9"
              />
            </div>

            <Select value={logLevel} onValueChange={(v) => setLogLevel(v as LogLevel)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Levels</SelectItem>
                <SelectItem value="debug">Debug</SelectItem>
                <SelectItem value="info">Info</SelectItem>
                <SelectItem value="warn">Warning</SelectItem>
                <SelectItem value="error">Error</SelectItem>
              </SelectContent>
            </Select>

            <Select value={limit.toString()} onValueChange={(v) => setLimit(Number(v))}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="100">Last 100</SelectItem>
                <SelectItem value="500">Last 500</SelectItem>
                <SelectItem value="1000">Last 1000</SelectItem>
                <SelectItem value="2000">Last 2000</SelectItem>
              </SelectContent>
            </Select>

            <div className="flex items-center gap-2">
              <Switch id="auto-scroll" checked={autoScroll} onCheckedChange={setAutoScroll} />
              <label htmlFor="auto-scroll" className="text-sm">
                Auto-scroll
              </label>
            </div>
          </div>

          <div className="flex gap-2 mt-3">
            <Badge variant="outline">Total: {parsedLogs.length}</Badge>
            {levelCounts.error > 0 && (
              <Badge variant="destructive">Errors: {levelCounts.error}</Badge>
            )}
            {levelCounts.warn > 0 && (
              <Badge variant="warning" className="bg-yellow-500 text-white">
                Warnings: {levelCounts.warn}
              </Badge>
            )}
            {searchTerm && <Badge variant="secondary">Filtered: {filteredLogs.length}</Badge>}
          </div>
        </CardContent>
      </Card>

      <Card className="flex-1 overflow-hidden flex flex-col min-h-0">
        <CardHeader className="pb-2 flex-shrink-0">
          <CardTitle className="text-sm font-medium">{filteredLogs.length} log entries</CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 overflow-hidden min-h-0">
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading logs...</div>
          ) : filteredLogs.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {searchTerm || logLevel !== 'all' ? 'No logs match your filters' : 'No logs available'}
            </div>
          ) : (
            <div
              ref={containerRef}
              className="h-full overflow-auto font-mono text-xs divide-y divide-border"
            >
              {filteredLogs.map((log, i) => {
                const config = LOG_LEVEL_CONFIG[log.level] || LOG_LEVEL_CONFIG.info
                return (
                  <div key={i} className={cn('px-4 py-2 hover:bg-muted/50', config.bg)}>
                    <div className="flex items-start gap-2">
                      {log.timestamp && (
                        <span className="text-muted-foreground whitespace-nowrap">{log.timestamp}</span>
                      )}
                      <span className={cn('flex items-center gap-1 uppercase font-semibold', config.color)}>
                        {config.icon}
                        {log.level}
                      </span>
                      <span className="flex-1 break-all">{highlightSearch(log.message)}</span>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
