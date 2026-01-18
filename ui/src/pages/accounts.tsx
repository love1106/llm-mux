import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { managementApi, type AuthFile } from '@/lib/api'
import { Trash2, RefreshCw, Plus, CheckCircle, XCircle, Clock } from 'lucide-react'

const statusIcons = {
  active: <CheckCircle className="h-4 w-4 text-green-500" />,
  disabled: <XCircle className="h-4 w-4 text-gray-500" />,
  error: <XCircle className="h-4 w-4 text-red-500" />,
  cooling: <Clock className="h-4 w-4 text-yellow-500" />,
  unavailable: <XCircle className="h-4 w-4 text-orange-500" />,
}

export function AccountsPage() {
  const queryClient = useQueryClient()

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['auth-files'],
    queryFn: () => managementApi.getAuthFiles(),
  })

  const deleteMutation = useMutation({
    mutationFn: (name: string) => managementApi.deleteAuthFile(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['auth-files'] }),
  })

  const startOAuth = async (provider: string) => {
    try {
      const response = await managementApi.oauthStart(provider)
      if (response.data.auth_url) {
        window.open(response.data.auth_url, '_blank')
      }
    } catch (error) {
      console.error('OAuth start failed:', error)
    }
  }

  const accounts: AuthFile[] = data?.data?.data?.files || []

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Accounts</h2>
          <p className="text-muted-foreground">Manage OAuth accounts and providers</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button size="sm" onClick={() => startOAuth('claude')}>
            <Plus className="h-4 w-4 mr-2" />
            Add Account
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-8 text-muted-foreground">Loading accounts...</div>
      ) : accounts.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            No accounts configured. Click "Add Account" to get started.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {accounts.map((account) => (
            <Card key={account.id || account.name}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {statusIcons[account.status] || statusIcons.unavailable}
                    <CardTitle className="text-lg">{account.label || account.name}</CardTitle>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => deleteMutation.mutate(account.name)}
                    disabled={deleteMutation.isPending}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Provider:</span>
                    <p className="font-medium">{account.provider || account.type}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Status:</span>
                    <p className="font-medium capitalize">{account.status}</p>
                  </div>
                  {account.email && (
                    <div>
                      <span className="text-muted-foreground">Email:</span>
                      <p className="font-medium truncate">{account.email}</p>
                    </div>
                  )}
                  {account.account_type && (
                    <div>
                      <span className="text-muted-foreground">Type:</span>
                      <p className="font-medium capitalize">{account.account_type}</p>
                    </div>
                  )}
                </div>
                {account.status_message && (
                  <p className="mt-2 text-sm text-muted-foreground">{account.status_message}</p>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
