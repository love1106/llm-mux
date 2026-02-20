import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
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
  FileJson,
  Copy,
  Check,
  Upload,
  Play,
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

const OAUTH_PROVIDERS = [
  { id: 'claude', name: 'Claude', description: 'Anthropic Claude Pro/Max', icon: '🟣', flowType: 'oauth', supportsManual: true },
  { id: 'gemini', name: 'Gemini', description: 'Google Gemini CLI', icon: '🔵', flowType: 'oauth', supportsManual: true },
  { id: 'antigravity', name: 'Antigravity', description: 'Google Cloud AI Companion', icon: '🌐', flowType: 'oauth', supportsManual: true },
  { id: 'copilot', name: 'GitHub Copilot', description: 'GitHub Copilot subscription', icon: '🐙', flowType: 'device', supportsManual: false },
  { id: 'qwen', name: 'Qwen', description: 'Alibaba Qwen', icon: '🟠', flowType: 'device', supportsManual: false },
  { id: 'codex', name: 'Codex', description: 'OpenAI Codex CLI', icon: '🟢', flowType: 'oauth', supportsManual: true },
  { id: 'iflow', name: 'iFlow', description: 'iFlow AI', icon: '⚡', flowType: 'oauth', supportsManual: false },
]

const MANUAL_OAUTH_PROVIDERS = OAUTH_PROVIDERS.filter(p => p.supportsManual)

export function AccountsPage() {
  const queryClient = useQueryClient()
  const [oauthDialogOpen, setOauthDialogOpen] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState('')
  const [oauthState, setOauthState] = useState<string | null>(null)
  const [oauthStatus, setOauthStatus] = useState<'idle' | 'pending' | 'success' | 'error'>('idle')
  const [authUrl, setAuthUrl] = useState<string | null>(null)
  const [deviceCode, setDeviceCode] = useState<{ userCode: string; verificationUrl: string } | null>(null)
  const [jsonDialogOpen, setJsonDialogOpen] = useState(false)
  const [jsonContent, setJsonContent] = useState<string>('')
  const [jsonFileName, setJsonFileName] = useState<string>('')
  const [copied, setCopied] = useState(false)
  const [addMode, setAddMode] = useState<'oauth' | 'manual' | 'import'>('oauth')
  const [importJson, setImportJson] = useState('')
  const [isImporting, setIsImporting] = useState(false)
  const [manualAuthUrl, setManualAuthUrl] = useState<string | null>(null)
  const [manualState, setManualState] = useState<string | null>(null)
  const [manualCode, setManualCode] = useState('')
  const [isSubmittingCode, setIsSubmittingCode] = useState(false)
  const [urlCopied, setUrlCopied] = useState(false)
  const [testDialogOpen, setTestDialogOpen] = useState(false)
  const [testResponse, setTestResponse] = useState('')
  const [isTesting, setIsTesting] = useState(false)
  const [testMessage, setTestMessage] = useState('Say hello!')
  const [testingAccount, setTestingAccount] = useState<AuthFile | null>(null)

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

  const refreshMutation = useMutation({
    mutationFn: (id: string) => managementApi.refreshAuthFile(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth-files'] })
      toast.success('Token refresh triggered')
    },
    onError: () => toast.error('Failed to refresh token'),
  })

  const resetOauthState = () => {
    setOauthState(null)
    setOauthStatus('idle')
    setAuthUrl(null)
    setDeviceCode(null)
    setSelectedProvider('')
    setAddMode('oauth')
    setImportJson('')
    setManualAuthUrl(null)
    setManualState(null)
    setManualCode('')
    setUrlCopied(false)
  }

  const handleImportJson = async () => {
    if (!importJson.trim()) {
      toast.error('Please paste JSON content')
      return
    }
    try {
      JSON.parse(importJson)
    } catch {
      toast.error('Invalid JSON format')
      return
    }
    setIsImporting(true)
    try {
      await managementApi.importRawJSON(importJson)
      toast.success('Account imported successfully')
      queryClient.invalidateQueries({ queryKey: ['auth-files'] })
      setOauthDialogOpen(false)
      resetOauthState()
    } catch (err: any) {
      toast.error(err?.response?.data?.error?.message || 'Failed to import account')
    } finally {
      setIsImporting(false)
    }
  }

  const viewJson = async (name: string) => {
    try {
      const response = await managementApi.getAuthFileContent(name)
      setJsonContent(JSON.stringify(response.data, null, 2))
      setJsonFileName(name)
      setJsonDialogOpen(true)
      setCopied(false)
    } catch {
      toast.error('Failed to load auth file')
    }
  }

  const copyJson = async () => {
    try {
      await navigator.clipboard.writeText(jsonContent)
      setCopied(true)
      toast.success('Copied to clipboard')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Failed to copy')
    }
  }

  const getTestModel = (provider: string): string => {
    const p = provider?.toLowerCase() || ''
    if (p.includes('claude') || p.includes('anthropic')) return 'claude-sonnet-4-20250514'
    if (p.includes('copilot') || p.includes('github')) return 'gpt-4o'
    if (p.includes('gemini') || p.includes('antigravity')) return 'gemini-2.5-flash'
    if (p.includes('qwen')) return 'qwen-plus'
    if (p.includes('codex')) return 'o4-mini'
    return 'claude-sonnet-4-20250514'
  }

  const openTestDialog = (account: AuthFile) => {
    setTestingAccount(account)
    setTestResponse('')
    setTestDialogOpen(true)
  }

  const runTest = async () => {
    if (!testingAccount) return
    const provider = testingAccount.provider || testingAccount.type || ''
    const model = getTestModel(provider)
    
    setTestResponse('')
    setIsTesting(true)

    try {
      const response = await fetch('/v1/messages?skip-auth=true', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'anthropic-version': '2023-06-01',
        },
        body: JSON.stringify({
          model,
          messages: [{ role: 'user', content: testMessage }],
          max_tokens: 1024,
          stream: true,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        setTestResponse(`Error ${response.status}: ${errorText}`)
        toast.error('Account test failed')
        return
      }

      const reader = response.body?.getReader()
      if (!reader) {
        setTestResponse('Error: No response body')
        toast.error('Account test failed')
        return
      }

      const decoder = new TextDecoder()
      let fullText = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6).trim()
            if (data === '[DONE]' || !data) continue
            try {
              const parsed = JSON.parse(data)
              const text = parsed.delta?.text || parsed.content_block?.text || parsed.content?.[0]?.text || ''
              if (text) {
                fullText += text
                setTestResponse(fullText)
              }
            } catch {
            }
          }
        }
      }

      toast.success('Account test passed')
    } catch (err) {
      setTestResponse(`Error: ${(err as Error).message}`)
      toast.error('Account test failed')
    } finally {
      setIsTesting(false)
    }
  }

  const startOAuth = async () => {
    if (!selectedProvider) return
    try {
      const response = await managementApi.oauthStart(selectedProvider)
      const data = response.data as any
      setOauthState(data.state)
      setOauthStatus('pending')

      if (data.flow_type === 'device' && data.user_code) {
        setDeviceCode({ userCode: data.user_code, verificationUrl: data.auth_url || data.verification_url })
        window.open(data.auth_url || data.verification_url, '_blank')
      } else if (data.auth_url) {
        setAuthUrl(data.auth_url)
        window.open(data.auth_url, '_blank')
      }
    } catch {
      toast.error('Failed to start OAuth flow')
      setOauthStatus('error')
    }
  }

  const startManualOAuth = async () => {
    if (!selectedProvider) return
    try {
      const response = await managementApi.oauthStart(selectedProvider, true)
      const data = response.data as any
      setManualState(data.state)
      setManualAuthUrl(data.auth_url)
      setUrlCopied(false)
    } catch {
      toast.error('Failed to generate OAuth URL')
    }
  }

  const copyAuthUrl = async () => {
    if (!manualAuthUrl) return
    try {
      await navigator.clipboard.writeText(manualAuthUrl)
      setUrlCopied(true)
      toast.success('URL copied to clipboard')
      setTimeout(() => setUrlCopied(false), 2000)
    } catch {
      toast.error('Failed to copy URL')
    }
  }

  const parseCodeFromInput = (input: string): string => {
    const trimmed = input.trim()
    if (trimmed.includes('?code=') || trimmed.includes('&code=')) {
      try {
        const url = new URL(trimmed)
        return url.searchParams.get('code') || trimmed
      } catch {
        const match = trimmed.match(/[?&]code=([^&#]+)/)
        return match ? match[1] : trimmed
      }
    }
    return trimmed
  }

  const submitManualCode = async () => {
    if (!manualState || !manualCode.trim()) return
    const code = parseCodeFromInput(manualCode)
    if (!code) {
      toast.error('Please enter the callback URL or code')
      return
    }
    setIsSubmittingCode(true)
    try {
      const response = await managementApi.oauthComplete(manualState, code)
      if (response.data.status === 'ok') {
        toast.success('Account connected successfully')
        queryClient.invalidateQueries({ queryKey: ['auth-files'] })
        setOauthDialogOpen(false)
        resetOauthState()
      } else {
        toast.error(response.data.error || 'Failed to complete authentication')
      }
    } catch (err: any) {
      toast.error(err?.response?.data?.error || 'Failed to complete authentication')
    } finally {
      setIsSubmittingCode(false)
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
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => openTestDialog(account)}
                        disabled={isTesting}
                        title="Test account"
                      >
                        <Play className="h-4 w-4 text-green-500" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => viewJson(account.name)}
                        title="View JSON"
                      >
                        <FileJson className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => refreshMutation.mutate(account.name)}
                        disabled={refreshMutation.isPending}
                        title="Refresh token"
                      >
                        <RefreshCw className={`h-4 w-4 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          if (confirm('Delete this account?')) {
                            deleteMutation.mutate(account.name)
                          }
                        }}
                        disabled={deleteMutation.isPending}
                        title="Delete account"
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
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
                    {account.expires_at && (
                      <div>
                        <span className="text-muted-foreground">Token Expires:</span>
                        <p className="font-medium">{new Date(account.expires_at).toLocaleString()}</p>
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
              <div className="space-y-3">
                <div className="flex gap-2 border-b">
                  <button
                    onClick={() => setAddMode('oauth')}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                      addMode === 'oauth' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    OAuth Login
                  </button>
                  <button
                    onClick={() => setAddMode('manual')}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                      addMode === 'manual' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    OAuth Token
                  </button>
                  <button
                    onClick={() => setAddMode('import')}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                      addMode === 'import' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    Import JSON
                  </button>
                </div>

                {addMode === 'oauth' && (
                  <>
                    <p className="text-sm text-muted-foreground">Select a provider to connect your account:</p>
                    <div className="grid gap-2">
                      {OAUTH_PROVIDERS.map((provider) => (
                        <button
                          key={provider.id}
                          onClick={() => setSelectedProvider(provider.id)}
                          className={`flex items-center gap-3 p-3 rounded-lg border text-left transition-colors hover:bg-accent/50 ${
                            selectedProvider === provider.id ? 'border-primary bg-primary/10 ring-1 ring-primary' : 'border-input bg-card'
                          }`}
                        >
                          <span className="text-2xl">{provider.icon}</span>
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-foreground">{provider.name}</div>
                            <div className="text-xs text-muted-foreground truncate">{provider.description}</div>
                          </div>
                          {provider.flowType === 'device' && (
                            <Badge variant="outline" className="text-xs shrink-0">Device</Badge>
                          )}
                        </button>
                      ))}
                    </div>
                    <Button onClick={startOAuth} disabled={!selectedProvider} className="w-full">
                      Connect {selectedProvider ? OAUTH_PROVIDERS.find(p => p.id === selectedProvider)?.name : 'Account'}
                    </Button>
                  </>
                )}

                {addMode === 'manual' && (
                  <div className="space-y-3">
                    <p className="text-sm text-muted-foreground">
                      Manual OAuth flow for remote/headless setups. Generate URL, login in browser, paste callback URL back.
                    </p>
                    
                    {!manualAuthUrl ? (
                      <>
                        <div className="grid gap-2">
                          {MANUAL_OAUTH_PROVIDERS.map((provider) => (
                            <button
                              key={provider.id}
                              onClick={() => setSelectedProvider(provider.id)}
                              className={`flex items-center gap-3 p-3 rounded-lg border text-left transition-colors hover:bg-accent/50 ${
                                selectedProvider === provider.id ? 'border-primary bg-primary/10 ring-1 ring-primary' : 'border-input bg-card'
                              }`}
                            >
                              <span className="text-2xl">{provider.icon}</span>
                              <div className="flex-1 min-w-0">
                                <div className="font-medium text-foreground">{provider.name}</div>
                                <div className="text-xs text-muted-foreground truncate">{provider.description}</div>
                              </div>
                            </button>
                          ))}
                        </div>
                        <Button onClick={startManualOAuth} disabled={!selectedProvider} className="w-full">
                          Generate Auth URL
                        </Button>
                      </>
                    ) : (
                      <div className="space-y-4">
                        <div className="space-y-2">
                          <label className="text-sm font-medium">1. Copy this URL and open in your browser:</label>
                          <div className="flex gap-2">
                            <input
                              type="text"
                              value={manualAuthUrl}
                              readOnly
                              className="flex-1 p-2 rounded-md border bg-muted font-mono text-xs truncate"
                            />
                            <Button variant="outline" size="sm" onClick={copyAuthUrl}>
                              {urlCopied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                            </Button>
                            <Button variant="outline" size="sm" onClick={() => window.open(manualAuthUrl, '_blank')}>
                              <ExternalLink className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                        
                        <div className="space-y-2">
                          <label className="text-sm font-medium">2. After login, paste the code or redirect URL:</label>
                          <p className="text-xs text-muted-foreground">
                            {selectedProvider === 'claude' || selectedProvider === 'codex'
                              ? 'You\'ll see a page with "Authentication Code" - click "Copy Code" and paste it here.'
                              : 'After authorizing, you\'ll be redirected to a page that won\'t load. Copy the full URL from your browser\'s address bar and paste it here.'}
                          </p>
                          <textarea
                            value={manualCode}
                            onChange={(e) => setManualCode(e.target.value)}
                            placeholder={selectedProvider === 'claude' || selectedProvider === 'codex'
                              ? 'Paste the authentication code here...'
                              : 'Paste the redirect URL or code here (e.g. http://localhost:8085/oauth2callback?code=...)'}
                            className="w-full h-20 p-2 rounded-md border bg-muted font-mono text-xs resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                          />
                        </div>
                        
                        <Button onClick={submitManualCode} disabled={isSubmittingCode || !manualCode.trim()} className="w-full">
                          {isSubmittingCode ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <CheckCircle className="h-4 w-4 mr-2" />}
                          Complete Authentication
                        </Button>
                        
                        <Button variant="outline" onClick={() => { setManualAuthUrl(null); setManualState(null); setManualCode(''); }} className="w-full">
                          Start Over
                        </Button>
                      </div>
                    )}
                  </div>
                )}

                {addMode === 'import' && (
                  <div className="space-y-3">
                    <p className="text-sm text-muted-foreground">Paste raw JSON credential (must include "type" field):</p>
                    <textarea
                      value={importJson}
                      onChange={(e) => setImportJson(e.target.value)}
                      placeholder='{"type": "claude", "email": "...", "token": {...}}'
                      className="w-full h-48 p-3 rounded-md border bg-muted font-mono text-sm resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <Button onClick={handleImportJson} disabled={isImporting || !importJson.trim()} className="w-full">
                      {isImporting ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <Upload className="h-4 w-4 mr-2" />}
                      Import Account
                    </Button>
                  </div>
                )}
              </div>
            )}

            {oauthStatus === 'pending' && (
              <div className="space-y-4">
                {deviceCode ? (
                  <div className="bg-muted p-4 rounded-md space-y-3">
                    <p className="text-sm font-medium">Enter this code on the authorization page:</p>
                    <div className="bg-background p-3 rounded border text-center">
                      <code className="text-2xl font-mono font-bold tracking-widest">{deviceCode.userCode}</code>
                    </div>
                    <Button variant="outline" size="sm" className="w-full" onClick={() => window.open(deviceCode.verificationUrl, '_blank')}>
                      <ExternalLink className="h-4 w-4 mr-2" />
                      Open Verification Page
                    </Button>
                  </div>
                ) : (
                  <div className="bg-muted p-4 rounded-md">
                    <p className="text-sm">Please complete the authentication in your browser.</p>
                    {authUrl && (
                      <Button variant="outline" size="sm" className="mt-2" onClick={() => window.open(authUrl, '_blank')}>
                        <ExternalLink className="h-4 w-4 mr-2" />
                        Open Auth Page
                      </Button>
                    )}
                  </div>
                )}
                <div className="flex items-center justify-center py-4">
                  <Loader2 className="h-8 w-8 animate-spin text-primary" />
                  <span className="ml-2 text-sm text-muted-foreground">Waiting for authorization...</span>
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

      <Dialog open={jsonDialogOpen} onOpenChange={setJsonDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh]">
          <DialogHeader>
            <div className="flex items-center justify-between">
              <DialogTitle className="flex items-center gap-2">
                <FileJson className="h-5 w-5" />
                {jsonFileName}
              </DialogTitle>
              <Button variant="outline" size="sm" onClick={copyJson} className="mr-8">
                {copied ? <Check className="h-4 w-4 mr-1" /> : <Copy className="h-4 w-4 mr-1" />}
                {copied ? 'Copied' : 'Copy'}
              </Button>
            </div>
          </DialogHeader>
          <div className="overflow-auto max-h-[60vh] bg-muted rounded-md p-4">
            <pre className="text-sm font-mono whitespace-pre-wrap break-all">{jsonContent}</pre>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={testDialogOpen} onOpenChange={(open) => {
        setTestDialogOpen(open)
        if (!open) {
          setTestingAccount(null)
          setTestResponse('')
        }
      }}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Play className="h-5 w-5 text-green-500" />
              Test Account
              {testingAccount && (
                <Badge variant="secondary" className="ml-2 capitalize">
                  {testingAccount.provider || testingAccount.type}
                </Badge>
              )}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">Message</label>
              <textarea
                value={testMessage}
                onChange={(e) => setTestMessage(e.target.value)}
                placeholder="Enter your message..."
                disabled={isTesting}
                className="w-full h-24 p-3 rounded-lg border bg-background text-sm resize-none focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary disabled:opacity-50 disabled:cursor-not-allowed transition-all"
              />
            </div>

            <Button 
              onClick={runTest} 
              disabled={isTesting || !testMessage.trim()} 
              className="w-full"
            >
              {isTesting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Testing...
                </>
              ) : (
                <>
                  <Zap className="h-4 w-4 mr-2" />
                  Send Test Request
                </>
              )}
            </Button>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium text-foreground">Response</label>
                {!isTesting && testResponse && (
                  <div className="flex items-center gap-1.5 text-green-600">
                    <CheckCircle className="h-3.5 w-3.5" />
                    <span className="text-xs font-medium">Complete</span>
                  </div>
                )}
              </div>
              <div 
                className="relative rounded-lg border bg-zinc-950 dark:bg-zinc-900 min-h-[120px] max-h-[280px] overflow-auto scroll-smooth"
                ref={(el) => {
                  if (el && isTesting) {
                    el.scrollTop = el.scrollHeight
                  }
                }}
              >
                <div className="p-4">
                  {isTesting && !testResponse && (
                    <div className="flex items-center gap-2 text-zinc-400">
                      <div className="flex gap-1">
                        <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                        <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                        <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                      </div>
                      <span className="text-sm">Waiting for response...</span>
                    </div>
                  )}
                  {testResponse && (
                    <div className="text-sm font-mono text-zinc-100 whitespace-pre-wrap leading-relaxed">
                      {testResponse}
                      {isTesting && (
                        <span className="inline-block w-2 h-4 ml-0.5 bg-green-400 animate-pulse" />
                      )}
                    </div>
                  )}
                  {!isTesting && !testResponse && (
                    <span className="text-zinc-500 text-sm italic">Response will appear here...</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
