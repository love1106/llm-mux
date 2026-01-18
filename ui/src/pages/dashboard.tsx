import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { managementApi } from '@/lib/api'
import { Activity, Users, Zap, AlertCircle } from 'lucide-react'

export function DashboardPage() {
  const { data: usage } = useQuery({
    queryKey: ['usage'],
    queryFn: () => managementApi.getUsage({ days: 7 }),
  })

  const { data: authFiles } = useQuery({
    queryKey: ['auth-files'],
    queryFn: () => managementApi.getAuthFiles(),
  })

  const stats = usage?.data?.data?.summary
  const accounts = authFiles?.data?.data?.files || []
  const activeAccounts = accounts.filter((a) => a.status === 'active').length

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <p className="text-muted-foreground">Overview of your llm-mux gateway</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_requests?.toLocaleString() || 0}</div>
            <p className="text-xs text-muted-foreground">Last 7 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Accounts</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeAccounts}</div>
            <p className="text-xs text-muted-foreground">{accounts.length} total</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Tokens</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.tokens?.total?.toLocaleString() || 0}</div>
            <p className="text-xs text-muted-foreground">Input + Output</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
            <AlertCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.total_requests
                ? ((stats.success_count / stats.total_requests) * 100).toFixed(1)
                : 100}
              %
            </div>
            <p className="text-xs text-muted-foreground">{stats?.failure_count || 0} failures</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
