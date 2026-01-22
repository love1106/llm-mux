import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { managementApi } from '@/lib/api'
import { Save, RefreshCw, Download, Upload, Key, Check, X, Eye, EyeOff, Copy } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/components/ui/toast'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const { managementKey, setManagementKey, isAuthenticated } = useAuthStore()
  const [keyInput, setKeyInput] = useState(managementKey || '')
  const [showKey, setShowKey] = useState(false)
  const [configYaml, setConfigYaml] = useState('')
  const [hasChanges, setHasChanges] = useState(false)

  const { data: configData, isLoading: configLoading } = useQuery({
    queryKey: ['config-yaml'],
    queryFn: () => managementApi.getConfigYAML(),
    enabled: isAuthenticated,
  })

  const { data: debugData } = useQuery({
    queryKey: ['debug'],
    queryFn: () => managementApi.getDebug(),
    enabled: isAuthenticated,
  })

  useEffect(() => {
    if (configData?.data && !configYaml) {
      setConfigYaml(configData.data)
    }
  }, [configData, configYaml])

  const saveConfigMutation = useMutation({
    mutationFn: (yaml: string) => managementApi.putConfigYAML(yaml),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config-yaml'] })
      setHasChanges(false)
      toast.success('Configuration saved!')
    },
    onError: (error: Error) => {
      toast.error(`Failed to save: ${error.message}`)
    },
  })

  const toggleDebugMutation = useMutation({
    mutationFn: (value: boolean) => managementApi.setDebug(value),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['debug'] })
      toast.success('Debug mode updated')
    },
    onError: () => toast.error('Failed to update debug mode'),
  })

  const handleSaveKey = () => {
    const trimmedKey = keyInput.trim() || null
    setManagementKey(trimmedKey)
    setKeyInput(trimmedKey || '')
    queryClient.invalidateQueries()
    toast.success('Management key saved')
  }

  const handleValidateKey = async () => {
    const trimmedKey = keyInput.trim()
    if (!trimmedKey) {
      toast.error('Please enter a key')
      return
    }
    setKeyInput(trimmedKey)
    setManagementKey(trimmedKey)
    try {
      await managementApi.getDebug()
      toast.success('Key is valid!')
      queryClient.invalidateQueries()
    } catch {
      toast.error('Invalid key')
      setManagementKey(null)
    }
  }

  const handleExportConfig = () => {
    const blob = new Blob([configYaml], { type: 'application/yaml' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'config.yaml'
    a.click()
    window.URL.revokeObjectURL(url)
    toast.success('Configuration exported')
  }

  const handleImportConfig = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        const content = e.target?.result as string
        setConfigYaml(content)
        setHasChanges(true)
        toast.success('Configuration imported - remember to save!')
      }
      reader.readAsText(file)
    }
    event.target.value = ''
  }

  const handleResetConfig = () => {
    if (configData?.data) {
      setConfigYaml(configData.data)
      setHasChanges(false)
      toast.success('Configuration reset')
    }
  }

  const isDebug = debugData?.data?.data?.debug || false

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
          <p className="text-muted-foreground">Configure llm-mux server settings</p>
        </div>
        {isAuthenticated && (
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleExportConfig}>
              <Download className="h-4 w-4 mr-2" />
              Export
            </Button>
            <label>
              <Button variant="outline" asChild>
                <span>
                  <Upload className="h-4 w-4 mr-2" />
                  Import
                </span>
              </Button>
              <input type="file" accept=".yaml,.yml" onChange={handleImportConfig} className="hidden" />
            </label>
          </div>
        )}
      </div>

      <Tabs defaultValue="auth" className="space-y-4">
        <TabsList>
          <TabsTrigger value="auth">Authentication</TabsTrigger>
          <TabsTrigger value="runtime" disabled={!isAuthenticated}>
            Runtime
          </TabsTrigger>
          <TabsTrigger value="config" disabled={!isAuthenticated}>
            Configuration
          </TabsTrigger>
        </TabsList>

        <TabsContent value="auth">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                Management Key
              </CardTitle>
              <CardDescription>API key for accessing management endpoints</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Input
                    type={showKey ? 'text' : 'password'}
                    value={keyInput}
                    onChange={(e) => setKeyInput(e.target.value)}
                    placeholder="Enter management key..."
                    className="pr-10"
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute right-1 top-1/2 -translate-y-1/2 h-7 w-7 p-0"
                    onClick={() => setShowKey(!showKey)}
                  >
                    {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                </div>
                <Button variant="outline" onClick={() => navigator.clipboard.writeText(keyInput)}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>

              <div className="flex gap-2">
                <Button onClick={handleSaveKey}>
                  <Save className="h-4 w-4 mr-2" />
                  Save Key
                </Button>
                <Button variant="outline" onClick={handleValidateKey}>
                  <Check className="h-4 w-4 mr-2" />
                  Validate
                </Button>
              </div>

              <div className="flex items-center gap-2 p-3 rounded-md bg-muted">
                {isAuthenticated ? (
                  <>
                    <Check className="h-4 w-4 text-green-500" />
                    <span className="text-sm text-green-600 dark:text-green-400">Key is configured and valid</span>
                  </>
                ) : (
                  <>
                    <X className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">No valid key configured</span>
                  </>
                )}
              </div>

              <p className="text-xs text-muted-foreground">
                Run <code className="bg-muted px-1 py-0.5 rounded">llm-mux init</code> to generate a management key
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="runtime">
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-base">Debug Mode</CardTitle>
                    <CardDescription>Enable verbose logging for troubleshooting</CardDescription>
                  </div>
                  <Switch
                    checked={isDebug}
                    onCheckedChange={(checked) => toggleDebugMutation.mutate(checked)}
                    disabled={toggleDebugMutation.isPending}
                  />
                </div>
              </CardHeader>
            </Card>

            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-base">Server Status</CardTitle>
                    <CardDescription>Current server information</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">API Endpoint:</span>
                    <p className="font-mono">http://localhost:8317</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">UI Endpoint:</span>
                    <p className="font-mono">http://localhost:8318</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Debug Mode:</span>
                    <p className={isDebug ? 'text-green-500' : 'text-muted-foreground'}>{isDebug ? 'Enabled' : 'Disabled'}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Configuration Editor</CardTitle>
              <CardDescription>Edit config.yaml directly. Changes require save to apply.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {configLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading configuration...</div>
              ) : (
                <>
                  <div className="relative">
                    <textarea
                      value={configYaml}
                      onChange={(e) => {
                        setConfigYaml(e.target.value)
                        setHasChanges(true)
                      }}
                      className="w-full h-96 p-4 font-mono text-sm border rounded-md bg-muted resize-none focus:outline-none focus:ring-2 focus:ring-ring"
                      spellCheck={false}
                      placeholder="# YAML configuration will appear here..."
                    />
                    {hasChanges && (
                      <div className="absolute top-2 right-2 px-2 py-1 text-xs bg-yellow-500 text-white rounded">
                        Unsaved changes
                      </div>
                    )}
                  </div>

                  <div className="flex justify-between">
                    <Button variant="outline" onClick={handleResetConfig} disabled={!hasChanges}>
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Reset
                    </Button>
                    <Button
                      onClick={() => saveConfigMutation.mutate(configYaml)}
                      disabled={!hasChanges || saveConfigMutation.isPending}
                    >
                      <Save className="h-4 w-4 mr-2" />
                      {saveConfigMutation.isPending ? 'Saving...' : 'Save Configuration'}
                    </Button>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
