import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'
import { managementApi, type AuthFile } from '@/lib/api'
import { toast } from '@/components/ui/toast'
import {
  Trash2,
  RefreshCw,
  Plus,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  Loader2,
  ExternalLink,
  Zap,
} from 'lucide-react'

const CLAUDE_DAILY_LIMIT = 45_000_000

const statusConfig: Record<string, { icon: React.ReactNode; color: string }> = {
  active: { icon: <CheckCircle className="h-4 w-4" />, color: 'text-green-500' },
  disabled: { icon: <XCircle className="h-4 w-4" />, color: 'text-gray-500' },
  error: { icon: <XCircle className="h-4 w-4" />, color: 'text-red-500' },
  cooling: { icon: <Clock className="h-4 w-4" />, color: 'text-yellow-500' },
  unavailable: { icon: <AlertCircle className="h-4 w-4" />, color: 'text-orange-500' },
}

const providerColors: Record<string, string> = {
  claude: 'bg-purple-500',
  openai: 'bg-green-500',
  google: 'bg-blue-500',
  gemini: 'bg-blue-400',
  anthropic: 'bg-purple-600',
}

const PROVIDERS = ['claude', 'gemini', 'openai']

export function AccountsPage() {
  const queryClient = useQueryClient()
  const [oauthDialogOpen, setOauthDialogOpen] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState('')
  const [oauthState, setOauthState] = useState<string | null>(null)
  const [oauthStatus, setOauthStatus] = useState<'idle' | 'pending' | 'success' | 'error'>('idle')
  const [authUrl, setAuthUrl] = useState<string | null>(null)

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['auth-files'],
    queryFn: () => managementApi.getAuthFiles(),
    refetchInterval: 10000,
  })

  const { data: usageData } = useQuery({
    queryKey: ['usage', 'day'],
    queryFn: () => managementApi.getUsage({ days: 1 }),
    refetchInterval: 30000,
  })

  const deleteMutation = useMutation({
    mutationFn: (name: string) => managementApi.deleteAuthFile(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth-files'] })
      toast.success('Account deleted')
    },
    onError: () => toast.error('Failed to delete account'),
  })

  const resetOauthState = () => {
    setOauthState(null)
    setOauthStatus('idle')
    setAuthUrl(null)
    setSelectedProvider('')
  }

  const startOAuth = async () => {
    if (!selectedProvider) return
    try {
      const response = await managementApi.oauthStart(selectedProvider)
      if (response.data.auth_url) {
        setAuthUrl(response.data.auth_url)
        setOauthState(response.data.state)
        setOauthStatus('pending')
        window.open(response.data.auth_url, '_blank')
      }
    } catch {
      toast.error('Failed to start OAuth flow')
      setOauthStatus('error')
    }
  }

  useEffect(() => {
    if (!oauthState || oauthStatus !== 'pending') return

    const interval = setInterval(async () => {
      try {
        const response = await managementApi.oauthStatus(oauthState)
        if (response.data.status === 'completed') {
          setOauthStatus('success')
          queryClient.invalidateQueries({ queryKey: ['auth-files'] })
          clearInterval(interval)
          setTimeout(() => {
            setOauthDialogOpen(false)
            resetOauthState()
          }, 2000)
        } else if (response.data.status === 'failed') {
          setOauthStatus('error')
          clearInterval(interval)
        }
      } catch {
        // Still pending
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [oauthState, oauthStatus, queryClient, resetOauthState])

  const handleDialogClose = () => {
    if (oauthState && oauthStatus === 'pending') {
      managementApi.oauthCancel(oauthState).catch(() => {})
    }
    setOauthDialogOpen(false)
    resetOauthState()
  }

  const accounts: AuthFile[] = data?.data?.data?.files || []
  const activeCount = accounts.filter((a) => a.status === 'active').length
  const providers = [...new Set(accounts.map((a) => a.provider || a.type))]

  const getAccountUsage = (account: AuthFile) => {
    const byAccount = usageData?.data?.data?.by_account || {}
    const key = Object.keys(byAccount).find((k) => k.includes(account.name) || k.includes(account.id))
    return key ? byAccount[key] : null
  }

  const formatTokens = (tokens: number) => {
    if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M`
    if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}K`
    return tokens.toString()
  }

  const getResetTime = (account: AuthFile) => {
    if (account.quota_state?.real_quota?.window_reset_at) {
      return new Date(account.quota_state.real_quota.window_reset_at)
    }
    const now = new Date()
    const nextMidnight = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() + 1))
    return nextMidnight
  }

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
          <Button size="sm" onClick={() => setOauthDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Account
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Accounts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{accounts.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Active Accounts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">{activeCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Providers</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-1 flex-wrap">
              {providers.map((p) => (
                <Badge key={p} variant="secondary" className="capitalize">
                  {p}
                </Badge>
              ))}
              {providers.length === 0 && <span className="text-muted-foreground text-sm">None</span>}
            </div>
          </CardContent>
        </Card>
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
          {accounts.map((account) => {
            const status = statusConfig[account.status] || statusConfig.unavailable
            const providerColor = providerColors[account.provider?.toLowerCase() || ''] || 'bg-gray-500'

            return (
              <Card key={account.id || account.name}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className={`w-3 h-3 rounded-full ${providerColor}`} />
                      <span className={status.color}>{status.icon}</span>
                      <CardTitle className="text-lg">{account.label || account.name}</CardTitle>
                      <Badge variant={account.status === 'active' ? 'success' : 'secondary'} className="capitalize">
                        {account.status}
                      </Badge>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        if (confirm('Delete this account?')) {
                          deleteMutation.mutate(account.name)
                        }
                      }}
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
                      <p className="font-medium capitalize">{account.provider || account.type}</p>
                    </div>
                    {account.email && (
                      <div>
                        <span className="text-muted-foreground">Email:</span>
                        <p className="font-medium truncate">{account.email}</p>
                      </div>
                    )}
                    {(account.account_type || account.subscription_type) && (
                      <div>
                        <span className="text-muted-foreground">Plan:</span>
                        <p className="font-medium capitalize">{account.subscription_type || account.account_type}</p>
                      </div>
                    )}
                    {account.last_refresh && (
                      <div>
                        <span className="text-muted-foreground">Last Refresh:</span>
                        <p className="font-medium">{new Date(account.last_refresh).toLocaleString()}</p>
                      </div>
                    )}
                  </div>
                  {account.status_message && (
                    <p className="mt-2 text-sm text-muted-foreground">{account.status_message}</p>
                  )}
                  {account.provider?.toLowerCase() === 'claude' && (() => {
                    const usage = getAccountUsage(account)
                    const tokensUsed = usage?.tokens?.total || account.quota_state?.total_tokens_used || 0
                    const limit = CLAUDE_DAILY_LIMIT
                    const percentage = Math.min((tokensUsed / limit) * 100, 100)
                    const resetTime = getResetTime(account)
                    const resetHours = Math.max(0, Math.ceil((resetTime.getTime() - Date.now()) / (1000 * 60 * 60)))

                    return (
                      <div className="mt-4 pt-4 border-t">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm text-muted-foreground flex items-center gap-1">
                            <Zap className="h-3 w-3" /> Daily Token Usage
                          </span>
                          <span className="text-sm font-medium">
                            {formatTokens(tokensUsed)} / {formatTokens(limit)} ({percentage.toFixed(1)}%)
                          </span>
                        </div>
                        <Progress value={percentage} className="h-2" />
                        <p className="text-xs text-muted-foreground mt-1">
                          Resets in {resetHours}h ({resetTime.toLocaleTimeString()})
                        </p>
                      </div>
                    )
                  })()}
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      <Dialog open={oauthDialogOpen} onOpenChange={handleDialogClose}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Account</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            {oauthStatus === 'idle' && (
              <>
                <div>
                  <label className="text-sm font-medium">Provider</label>
                  <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                    <SelectTrigger className="mt-1">
                      <SelectValue placeholder="Select a provider" />
                    </SelectTrigger>
                    <SelectContent>
                      {PROVIDERS.map((provider) => (
                        <SelectItem key={provider} value={provider}>
                          <span className="capitalize">{provider}</span>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button onClick={startOAuth} disabled={!selectedProvider} className="w-full">
                  Start OAuth Flow
                </Button>
              </>
            )}

            {oauthStatus === 'pending' && (
              <div className="space-y-4">
                <div className="bg-muted p-4 rounded-md">
                  <p className="text-sm">Please complete the authentication in your browser.</p>
                  {authUrl && (
                    <Button variant="outline" size="sm" className="mt-2" onClick={() => window.open(authUrl, '_blank')}>
                      <ExternalLink className="h-4 w-4 mr-2" />
                      Open Auth Page
                    </Button>
                  )}
                </div>
                <div className="flex items-center justify-center py-4">
                  <Loader2 className="h-8 w-8 animate-spin text-primary" />
                </div>
                <Button variant="outline" onClick={handleDialogClose} className="w-full">
                  Cancel
                </Button>
              </div>
            )}

            {oauthStatus === 'success' && (
              <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-md flex items-center gap-2">
                <CheckCircle className="h-5 w-5 text-green-500" />
                <span>Account successfully connected!</span>
              </div>
            )}

            {oauthStatus === 'error' && (
              <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-md">
                <p className="text-destructive">Authentication failed. Please try again.</p>
                <Button variant="outline" onClick={resetOauthState} className="mt-2">
                  Try Again
                </Button>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
