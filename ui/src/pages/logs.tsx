import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { managementApi } from '@/lib/api'
import { RefreshCw, Trash2, Download } from 'lucide-react'
import { toast } from '@/components/ui/toast'

export function LogsPage() {
  const [autoRefresh, setAutoRefresh] = useState(false)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['logs'],
    queryFn: () => managementApi.getLogs({ limit: 200 }),
    refetchInterval: autoRefresh ? 3000 : false,
  })

  const lines = data?.data?.data?.lines || []

  const handleClear = async () => {
    try {
      await managementApi.deleteLogs()
      refetch()
      toast.success('Logs cleared')
    } catch {
      toast.error('Failed to clear logs')
    }
  }

  const handleDownload = () => {
    const blob = new Blob([lines.join('\n')], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `llm-mux-logs-${new Date().toISOString().slice(0, 10)}.txt`
    a.click()
    URL.revokeObjectURL(url)
    toast.success('Logs downloaded')
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Logs</h2>
          <p className="text-muted-foreground">Server request logs</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={autoRefresh ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto-refresh
          </Button>
          <Button variant="outline" size="sm" onClick={handleDownload} disabled={lines.length === 0}>
            <Download className="h-4 w-4 mr-2" />
            Download
          </Button>
          <Button variant="outline" size="sm" onClick={handleClear}>
            <Trash2 className="h-4 w-4 mr-2" />
            Clear
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{lines.length} log entries</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading logs...</div>
          ) : lines.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No logs available. Enable logging in settings.
            </div>
          ) : (
            <div className="bg-muted rounded-md p-4 max-h-[600px] overflow-auto font-mono text-xs">
              {lines.map((line, i) => (
                <div key={i} className="py-0.5 hover:bg-background/50">
                  {line}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
