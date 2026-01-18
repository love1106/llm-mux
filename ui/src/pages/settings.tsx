import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { managementApi } from '@/lib/api'
import { Save, RefreshCw } from 'lucide-react'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/components/ui/toast'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const { managementKey, setManagementKey } = useAuthStore()
  const [keyInput, setKeyInput] = useState(managementKey || '')
  const [configYaml, setConfigYaml] = useState('')

  const { data: configData, isLoading: configLoading } = useQuery({
    queryKey: ['config-yaml'],
    queryFn: () => managementApi.getConfigYAML(),
    enabled: !!managementKey,
  })

  const { data: debugData } = useQuery({
    queryKey: ['debug'],
    queryFn: () => managementApi.getDebug(),
    enabled: !!managementKey,
  })

  const saveConfigMutation = useMutation({
    mutationFn: (yaml: string) => managementApi.putConfigYAML(yaml),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config-yaml'] })
      toast.success('Configuration saved!')
    },
    onError: (error: Error) => {
      toast.error(`Failed to save: ${error.message}`)
    },
  })

  const toggleDebugMutation = useMutation({
    mutationFn: (value: boolean) => managementApi.setDebug(value),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['debug'] }),
  })

  const handleSaveKey = () => {
    setManagementKey(keyInput || null)
    queryClient.invalidateQueries()
    toast.success('Management key saved')
  }

  const yamlContent = configYaml || configData?.data || ''
  const isDebug = debugData?.data?.data?.debug || false

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">Configure llm-mux server settings</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Management Key</CardTitle>
          <CardDescription>API key for accessing management endpoints</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <input
              type="password"
              value={keyInput}
              onChange={(e) => setKeyInput(e.target.value)}
              placeholder="Enter management key..."
              className="flex-1 px-3 py-2 border rounded-md bg-background text-sm"
            />
            <Button onClick={handleSaveKey}>Save Key</Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Run <code className="bg-muted px-1 rounded">llm-mux init</code> to generate a key
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Debug Mode</CardTitle>
          <CardDescription>Enable verbose logging for troubleshooting</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            variant={isDebug ? 'default' : 'outline'}
            onClick={() => toggleDebugMutation.mutate(!isDebug)}
            disabled={toggleDebugMutation.isPending || !managementKey}
          >
            {isDebug ? 'Enabled' : 'Disabled'}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
          <CardDescription>Edit config.yaml directly</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!managementKey ? (
            <p className="text-muted-foreground">Enter management key to view configuration</p>
          ) : configLoading ? (
            <p className="text-muted-foreground">Loading configuration...</p>
          ) : (
            <>
              <textarea
                value={yamlContent}
                onChange={(e) => setConfigYaml(e.target.value)}
                className="w-full h-64 p-4 font-mono text-sm border rounded-md bg-muted"
                spellCheck={false}
              />
              <div className="flex gap-2">
                <Button
                  onClick={() => saveConfigMutation.mutate(configYaml || yamlContent)}
                  disabled={saveConfigMutation.isPending}
                >
                  <Save className="h-4 w-4 mr-2" />
                  Save Configuration
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    setConfigYaml('')
                    queryClient.invalidateQueries({ queryKey: ['config-yaml'] })
                  }}
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Reset
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
