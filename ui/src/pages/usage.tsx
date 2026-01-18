import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { managementApi } from '@/lib/api'

export function UsagePage() {
  const { data, isLoading } = useQuery({
    queryKey: ['usage'],
    queryFn: () => managementApi.getUsage({ days: 30 }),
  })

  const stats = data?.data?.data
  const byProvider = stats?.by_provider || {}
  const byAccount = stats?.by_account || {}

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Usage Statistics</h2>
        <p className="text-muted-foreground">Request and token usage over time</p>
      </div>

      {isLoading ? (
        <div className="text-center py-8 text-muted-foreground">Loading usage data...</div>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.summary?.total_requests?.toLocaleString() || 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Input Tokens</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.summary?.tokens?.input?.toLocaleString() || 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Output Tokens</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.summary?.tokens?.output?.toLocaleString() || 0}</div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Usage by Provider</CardTitle>
            </CardHeader>
            <CardContent>
              {Object.keys(byProvider).length === 0 ? (
                <p className="text-muted-foreground">No provider data available</p>
              ) : (
                <div className="space-y-4">
                  {Object.entries(byProvider).map(([provider, data]) => (
                    <div key={provider} className="flex items-center justify-between border-b pb-2">
                      <span className="font-medium capitalize">{provider}</span>
                      <div className="text-right">
                        <p className="text-sm">{data.requests?.toLocaleString() || 0} requests</p>
                        <p className="text-xs text-muted-foreground">{data.tokens?.total?.toLocaleString() || 0} tokens</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Usage by Account</CardTitle>
            </CardHeader>
            <CardContent>
              {Object.keys(byAccount).length === 0 ? (
                <p className="text-muted-foreground">No account data available</p>
              ) : (
                <div className="space-y-4">
                  {Object.entries(byAccount).map(([account, data]) => (
                    <div key={account} className="flex items-center justify-between border-b pb-2">
                      <span className="font-medium truncate max-w-[200px]">{account}</span>
                      <div className="text-right">
                        <p className="text-sm">{data.requests?.toLocaleString() || 0} requests</p>
                        <p className="text-xs text-muted-foreground">{data.tokens?.total?.toLocaleString() || 0} tokens</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
