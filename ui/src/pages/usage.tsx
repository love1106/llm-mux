import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'
import { managementApi } from '@/lib/api'
import { toast } from '@/components/ui/toast'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  Legend,
} from 'recharts'
import { Download, RefreshCw, TrendingUp, Minus } from 'lucide-react'
import { format } from 'date-fns'

type TimeRange = '1h' | '24h' | '7d' | '30d'

const COLORS = ['#8884d8', '#82ca9d', '#ffc658', '#ff8042', '#8dd1e1', '#d084d8', '#84d8c9']

const VALID_TABS = ['timeline', 'providers', 'accounts', 'ips'] as const
const VALID_RANGES: TimeRange[] = ['1h', '24h', '7d', '30d']

export function UsagePage() {
  const [searchParams, setSearchParams] = useSearchParams()

  const tab = VALID_TABS.includes(searchParams.get('tab') as (typeof VALID_TABS)[number])
    ? (searchParams.get('tab') as string)
    : 'timeline'

  const timeRange: TimeRange = VALID_RANGES.includes(searchParams.get('range') as TimeRange)
    ? (searchParams.get('range') as TimeRange)
    : '7d'

  const setTab = (value: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      if (value === 'timeline') next.delete('tab')
      else next.set('tab', value)
      return next
    })
  }

  const setTimeRange = (value: TimeRange) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      if (value === '7d') next.delete('range')
      else next.set('range', value)
      return next
    })
  }

  const getDays = () => {
    switch (timeRange) {
      case '1h':
        return 1
      case '24h':
        return 1
      case '7d':
        return 7
      case '30d':
        return 30
      default:
        return 7
    }
  }

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['usage', timeRange],
    queryFn: () => managementApi.getUsage({ days: getDays() }),
    refetchInterval: 30000,
  })

  const stats = data?.data?.data
  const byProvider = stats?.by_provider || {}
  const byAccount = stats?.by_account || {}
  const byIP = stats?.by_ip || {}
  const timeline = stats?.timeline?.by_day || []

  const providerData = Object.entries(byProvider).map(([name, data]) => ({
    name: name.charAt(0).toUpperCase() + name.slice(1),
    requests: data.requests || 0,
    tokens: data.tokens?.total || 0,
  }))

  const accountData = Object.entries(byAccount).map(([name, data]) => ({
    name: name.length > 20 ? name.slice(0, 20) + '...' : name,
    fullName: name,
    requests: data.requests || 0,
    tokens: data.tokens?.total || 0,
  }))

  const ipData = Object.entries(byIP)
    .map(([ip, data]) => ({
      ip,
      requests: data.requests || 0,
      tokens: data.tokens?.total || 0,
      input: data.tokens?.input || 0,
      output: data.tokens?.output || 0,
      models: data.models || [],
      lastSeen: data.last_seen_at || '',
    }))
    .sort((a, b) => b.tokens - a.tokens)

  const totalTokens = stats?.summary?.tokens?.total || 0
  const maxAccountTokens = Math.max(...accountData.map((a) => a.tokens), 1)
  const maxIPTokens = Math.max(...ipData.map((i) => i.tokens), 1)

  const handleExport = (format: 'csv' | 'json') => {
    if (!stats) return

    const filename = `usage-${new Date().toISOString().slice(0, 10)}.${format}`

    if (format === 'json') {
      const blob = new Blob([JSON.stringify(stats, null, 2)], { type: 'application/json' })
      downloadBlob(blob, filename)
    } else {
      const rows = [
        ['Metric', 'Value'],
        ['Total Requests', stats.summary?.total_requests?.toString() || '0'],
        ['Success Count', stats.summary?.success_count?.toString() || '0'],
        ['Failure Count', stats.summary?.failure_count?.toString() || '0'],
        ['Total Tokens', stats.summary?.tokens?.total?.toString() || '0'],
        ['Input Tokens', stats.summary?.tokens?.input?.toString() || '0'],
        ['Output Tokens', stats.summary?.tokens?.output?.toString() || '0'],
        [],
        ['Provider', 'Requests', 'Tokens'],
        ...Object.entries(byProvider).map(([name, data]) => [
          name,
          data.requests?.toString() || '0',
          data.tokens?.total?.toString() || '0',
        ]),
        [],
        ['IP Address', 'Requests', 'Input Tokens', 'Output Tokens', 'Total Tokens'],
        ...Object.entries(byIP).map(([ip, data]) => [
          ip,
          (data.requests || 0).toString(),
          (data.tokens?.input || 0).toString(),
          (data.tokens?.output || 0).toString(),
          (data.tokens?.total || 0).toString(),
        ]),
      ]
      const csv = rows.map((row) => row.join(',')).join('\n')
      const blob = new Blob([csv], { type: 'text/csv' })
      downloadBlob(blob, filename)
    }
    toast.success(`Exported as ${format.toUpperCase()}`)
  }

  const downloadBlob = (blob: Blob, filename: string) => {
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    window.URL.revokeObjectURL(url)
  }

  const successRate = stats?.summary?.total_requests
    ? ((stats.summary.success_count / stats.summary.total_requests) * 100).toFixed(1)
    : '100'

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Usage Statistics</h2>
          <p className="text-muted-foreground">Request and token usage over time</p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={timeRange} onValueChange={(v) => setTimeRange(v as TimeRange)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1h">Last Hour</SelectItem>
              <SelectItem value="24h">Last 24h</SelectItem>
              <SelectItem value="7d">Last 7 Days</SelectItem>
              <SelectItem value="30d">Last 30 Days</SelectItem>
            </SelectContent>
          </Select>

          <Button variant="outline" size="icon" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4" />
          </Button>

          <Button variant="outline" onClick={() => handleExport('csv')}>
            <Download className="h-4 w-4 mr-2" />
            CSV
          </Button>

          <Button variant="outline" onClick={() => handleExport('json')}>
            <Download className="h-4 w-4 mr-2" />
            JSON
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-8 text-muted-foreground">Loading usage data...</div>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Total Requests</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.summary?.total_requests?.toLocaleString() || 0}</div>
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <TrendingUp className="h-3 w-3 text-green-500" />
                  <span>Last {timeRange}</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Total Tokens</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{totalTokens.toLocaleString()}</div>
                <div className="text-xs text-muted-foreground">
                  In: {stats?.summary?.tokens?.input?.toLocaleString() || 0} / Out:{' '}
                  {stats?.summary?.tokens?.output?.toLocaleString() || 0}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Success Rate</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-500">{successRate}%</div>
                <div className="text-xs text-muted-foreground">
                  {stats?.summary?.failure_count || 0} failures
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Avg Tokens/Request</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats?.summary?.total_requests
                    ? Math.round(totalTokens / stats.summary.total_requests).toLocaleString()
                    : 0}
                </div>
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Minus className="h-3 w-3" />
                  <span>Average</span>
                </div>
              </CardContent>
            </Card>
          </div>

          <Tabs value={tab} onValueChange={setTab} className="space-y-4">
            <TabsList>
              <TabsTrigger value="timeline">Timeline</TabsTrigger>
              <TabsTrigger value="providers">By Provider</TabsTrigger>
              <TabsTrigger value="accounts">By Account</TabsTrigger>
              <TabsTrigger value="ips">By IP</TabsTrigger>
            </TabsList>

            <TabsContent value="timeline">
              <Card>
                <CardHeader>
                  <CardTitle>Request & Token Timeline</CardTitle>
                </CardHeader>
                <CardContent>
                  {timeline.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No timeline data available</div>
                  ) : (
                    <ResponsiveContainer width="100%" height={300}>
                      <LineChart data={timeline}>
                        <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                        <XAxis
                          dataKey="day"
                          tickFormatter={(value) => {
                            try {
                              return format(new Date(value), 'MM/dd')
                            } catch {
                              return value
                            }
                          }}
                          className="text-xs"
                        />
                        <YAxis yAxisId="left" className="text-xs" />
                        <YAxis yAxisId="right" orientation="right" className="text-xs" />
                        <Tooltip
                          contentStyle={{
                            backgroundColor: 'hsl(var(--card))',
                            border: '1px solid hsl(var(--border))',
                            borderRadius: '6px',
                          }}
                        />
                        <Legend />
                        <Line
                          yAxisId="left"
                          type="monotone"
                          dataKey="requests"
                          stroke="#8884d8"
                          strokeWidth={2}
                          dot={false}
                          name="Requests"
                        />
                        <Line
                          yAxisId="right"
                          type="monotone"
                          dataKey="tokens"
                          stroke="#82ca9d"
                          strokeWidth={2}
                          dot={false}
                          name="Tokens"
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="providers">
              <div className="grid gap-4 md:grid-cols-2">
                <Card>
                  <CardHeader>
                    <CardTitle>Requests by Provider</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {providerData.length === 0 ? (
                      <div className="text-center py-8 text-muted-foreground">No provider data</div>
                    ) : (
                      <ResponsiveContainer width="100%" height={250}>
                        <PieChart>
                          <Pie
                            data={providerData}
                            dataKey="requests"
                            nameKey="name"
                            cx="50%"
                            cy="50%"
                            outerRadius={80}
                            label={({ name, percent }) => `${name} ${((percent || 0) * 100).toFixed(0)}%`}
                          >
                            {providerData.map((_, index) => (
                              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                            ))}
                          </Pie>
                          <Tooltip />
                        </PieChart>
                      </ResponsiveContainer>
                    )}
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Tokens by Provider</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {providerData.length === 0 ? (
                      <div className="text-center py-8 text-muted-foreground">No provider data</div>
                    ) : (
                      <ResponsiveContainer width="100%" height={250}>
                        <BarChart data={providerData} layout="vertical">
                          <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                          <XAxis type="number" />
                          <YAxis dataKey="name" type="category" width={80} />
                          <Tooltip />
                          <Bar dataKey="tokens" fill="#82ca9d" name="Tokens" />
                        </BarChart>
                      </ResponsiveContainer>
                    )}
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            <TabsContent value="accounts">
              <Card>
                <CardHeader>
                  <CardTitle>Usage by Account</CardTitle>
                </CardHeader>
                <CardContent>
                  {accountData.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No account data available</div>
                  ) : (
                    <div className="space-y-4">
                      {accountData.map((account) => (
                        <div key={account.fullName} className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-sm truncate max-w-[200px]" title={account.fullName}>
                              {account.name}
                            </span>
                            <div className="text-right text-sm">
                              <span className="text-muted-foreground">{account.requests.toLocaleString()} req</span>
                              <span className="mx-2">·</span>
                              <span>{account.tokens.toLocaleString()} tokens</span>
                            </div>
                          </div>
                          <Progress value={(account.tokens / maxAccountTokens) * 100} />
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="ips">
              <Card>
                <CardHeader>
                  <CardTitle>Usage by Client IP</CardTitle>
                </CardHeader>
                <CardContent>
                  {ipData.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No IP data available</div>
                  ) : (
                    <div className="space-y-6">
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b">
                              <th className="text-left py-3 px-2 font-medium text-muted-foreground">IP Address</th>
                              <th className="text-right py-3 px-2 font-medium text-muted-foreground">Requests</th>
                              <th className="text-right py-3 px-2 font-medium text-muted-foreground">Input Tokens</th>
                              <th className="text-right py-3 px-2 font-medium text-muted-foreground">Output Tokens</th>
                              <th className="text-right py-3 px-2 font-medium text-muted-foreground">Total Tokens</th>
                              <th className="text-left py-3 px-2 font-medium text-muted-foreground">Models</th>
                              <th className="text-right py-3 px-2 font-medium text-muted-foreground">Last Seen</th>
                            </tr>
                          </thead>
                          <tbody>
                            {ipData.map((row) => (
                              <tr key={row.ip} className="border-b last:border-0">
                                <td className="py-3 px-2 font-medium">{row.ip}</td>
                                <td className="py-3 px-2 text-right text-muted-foreground">{row.requests.toLocaleString()}</td>
                                <td className="py-3 px-2 text-right text-muted-foreground">{row.input.toLocaleString()}</td>
                                <td className="py-3 px-2 text-right text-muted-foreground">{row.output.toLocaleString()}</td>
                                <td className="py-3 px-2 text-right font-medium">{row.tokens.toLocaleString()}</td>
                                <td className="py-3 px-2 text-muted-foreground">{row.models.join(', ')}</td>
                                <td className="py-3 px-2 text-right text-muted-foreground">
                                  {row.lastSeen ? format(new Date(row.lastSeen), 'MMM dd, HH:mm') : '—'}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                      <div className="space-y-4">
                        {ipData.map((row) => (
                          <div key={row.ip} className="space-y-2">
                            <div className="flex items-center justify-between">
                              <span className="font-medium text-sm">{row.ip}</span>
                              <div className="text-right text-sm">
                                <span className="text-muted-foreground">{row.requests.toLocaleString()} req</span>
                                <span className="mx-2">·</span>
                                <span>{row.tokens.toLocaleString()} tokens</span>
                              </div>
                            </div>
                            <Progress value={(row.tokens / maxIPTokens) * 100} />
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </>
      )}
    </div>
  )
}
