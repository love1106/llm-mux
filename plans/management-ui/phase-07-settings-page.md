# Phase 7: Settings Page

## Context
Build the Settings page with YAML config editor, boolean toggles, and API key management for comprehensive llm-mux configuration.

## Overview
Create a settings interface featuring a Monaco-based YAML editor with validation, toggle switches for runtime settings, and API key management with CRUD operations.

## Requirements
- YAML config editor with syntax highlighting
- Real-time YAML validation
- Boolean settings toggles
- API key generation and management
- Config backup and restore
- Settings change history

## Implementation Steps

### 1. Settings Page Component
Create `src/pages/SettingsPage.tsx`:
```typescript
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/ui/use-toast';
import { Save, RotateCcw, Download, Upload } from 'lucide-react';
import { api } from '@/lib/api';
import { ConfigEditor } from '@/components/settings/ConfigEditor';
import { BooleanSettings } from '@/components/settings/BooleanSettings';
import { ApiKeysManager } from '@/components/settings/ApiKeysManager';
import { ProxySettings } from '@/components/settings/ProxySettings';

export function SettingsPage() {
  const [configYaml, setConfigYaml] = useState('');
  const [hasChanges, setHasChanges] = useState(false);
  const queryClient = useQueryClient();

  const { data: originalConfig } = useQuery({
    queryKey: ['config-yaml'],
    queryFn: api.getConfigYAML,
    onSuccess: (data) => {
      setConfigYaml(data);
    },
  });

  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const [debug, loggingToFile, usageStats] = await Promise.all([
        api.getDebug(),
        api.getLoggingToFile(),
        api.getUsageStatisticsEnabled(),
      ]);
      return { debug, loggingToFile, usageStats };
    },
  });

  const saveConfigMutation = useMutation({
    mutationFn: api.updateConfigYAML,
    onSuccess: () => {
      toast({
        title: 'Configuration saved',
        description: 'Settings have been updated successfully',
      });
      setHasChanges(false);
      queryClient.invalidateQueries(['config']);
    },
    onError: (error: any) => {
      toast({
        title: 'Save failed',
        description: error.response?.data?.error || 'Failed to save configuration',
        variant: 'destructive',
      });
    },
  });

  const handleSaveConfig = () => {
    saveConfigMutation.mutate(configYaml);
  };

  const handleResetConfig = () => {
    if (originalConfig) {
      setConfigYaml(originalConfig);
      setHasChanges(false);
    }
  };

  const handleExportConfig = () => {
    const blob = new Blob([configYaml], { type: 'application/yaml' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'config.yaml';
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const handleImportConfig = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setConfigYaml(content);
        setHasChanges(true);
      };
      reader.readAsText(file);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Settings</h1>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleExportConfig}
          >
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
          <label>
            <Button variant="outline" as="div">
              <Upload className="h-4 w-4 mr-2" />
              Import
            </Button>
            <input
              type="file"
              accept=".yaml,.yml"
              onChange={handleImportConfig}
              className="hidden"
            />
          </label>
        </div>
      </div>

      <Tabs defaultValue="config" className="space-y-4">
        <TabsList>
          <TabsTrigger value="config">Configuration</TabsTrigger>
          <TabsTrigger value="runtime">Runtime Settings</TabsTrigger>
          <TabsTrigger value="api-keys">API Keys</TabsTrigger>
          <TabsTrigger value="proxy">Proxy</TabsTrigger>
        </TabsList>

        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Configuration Editor</CardTitle>
              <CardDescription>
                Edit the YAML configuration file. Changes require save to apply.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <ConfigEditor
                value={configYaml}
                onChange={(value) => {
                  setConfigYaml(value);
                  setHasChanges(true);
                }}
              />

              <div className="flex justify-between">
                <Button
                  variant="outline"
                  onClick={handleResetConfig}
                  disabled={!hasChanges}
                >
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Reset
                </Button>

                <Button
                  onClick={handleSaveConfig}
                  disabled={!hasChanges || saveConfigMutation.isLoading}
                >
                  <Save className="h-4 w-4 mr-2" />
                  Save Changes
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="runtime">
          <BooleanSettings settings={settings} />
        </TabsContent>

        <TabsContent value="api-keys">
          <ApiKeysManager />
        </TabsContent>

        <TabsContent value="proxy">
          <ProxySettings />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

### 2. Config Editor Component
Create `src/components/settings/ConfigEditor.tsx`:
```typescript
import { useEffect, useRef } from 'react';
import * as monaco from 'monaco-editor';
import { useTheme } from '@/hooks/use-theme';

interface Props {
  value: string;
  onChange: (value: string) => void;
}

export function ConfigEditor({ value, onChange }: Props) {
  const editorRef = useRef<HTMLDivElement>(null);
  const monacoRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const { theme } = useTheme();

  useEffect(() => {
    if (editorRef.current && !monacoRef.current) {
      monacoRef.current = monaco.editor.create(editorRef.current, {
        value,
        language: 'yaml',
        theme: theme === 'dark' ? 'vs-dark' : 'vs',
        minimap: { enabled: false },
        fontSize: 14,
        lineNumbers: 'on',
        scrollBeyondLastLine: false,
        wordWrap: 'on',
        automaticLayout: true,
      });

      monacoRef.current.onDidChangeModelContent(() => {
        onChange(monacoRef.current?.getValue() || '');
      });
    }

    return () => {
      monacoRef.current?.dispose();
    };
  }, []);

  useEffect(() => {
    if (monacoRef.current && value !== monacoRef.current.getValue()) {
      monacoRef.current.setValue(value);
    }
  }, [value]);

  useEffect(() => {
    monaco.editor.setTheme(theme === 'dark' ? 'vs-dark' : 'vs');
  }, [theme]);

  return (
    <div className="border rounded-md overflow-hidden">
      <div ref={editorRef} className="h-[500px]" />
    </div>
  );
}
```

### 3. Boolean Settings Component
Create `src/components/settings/BooleanSettings.tsx`:
```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { toast } from '@/components/ui/use-toast';
import { api } from '@/lib/api';

interface Settings {
  debug: boolean;
  loggingToFile: boolean;
  usageStats: boolean;
}

interface Props {
  settings?: Settings;
}

export function BooleanSettings({ settings }: Props) {
  const queryClient = useQueryClient();

  const createToggleMutation = (
    key: 'debug' | 'loggingToFile' | 'usageStats',
    apiMethod: (value: boolean) => Promise<void>
  ) =>
    useMutation({
      mutationFn: apiMethod,
      onSuccess: () => {
        queryClient.invalidateQueries(['settings']);
        toast({
          title: 'Setting updated',
          description: `${key} has been ${settings?.[key] ? 'disabled' : 'enabled'}`,
        });
      },
    });

  const debugMutation = createToggleMutation('debug', api.setDebug);
  const loggingMutation = createToggleMutation('loggingToFile', api.setLoggingToFile);
  const usageStatsMutation = createToggleMutation('usageStats', api.setUsageStatisticsEnabled);

  const settingsConfig = [
    {
      id: 'debug',
      title: 'Debug Mode',
      description: 'Enable detailed debug logging for troubleshooting',
      value: settings?.debug || false,
      onToggle: (value: boolean) => debugMutation.mutate(value),
    },
    {
      id: 'logging-to-file',
      title: 'Log to File',
      description: 'Save logs to file for persistent storage',
      value: settings?.loggingToFile || false,
      onToggle: (value: boolean) => loggingMutation.mutate(value),
    },
    {
      id: 'usage-stats',
      title: 'Usage Statistics',
      description: 'Collect anonymous usage statistics',
      value: settings?.usageStats || false,
      onToggle: (value: boolean) => usageStatsMutation.mutate(value),
    },
  ];

  return (
    <div className="space-y-4">
      {settingsConfig.map((setting) => (
        <Card key={setting.id}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-base">{setting.title}</CardTitle>
                <CardDescription className="mt-1">
                  {setting.description}
                </CardDescription>
              </div>
              <Switch
                id={setting.id}
                checked={setting.value}
                onCheckedChange={setting.onToggle}
              />
            </div>
          </CardHeader>
        </Card>
      ))}
    </div>
  );
}
```

### 4. API Keys Manager Component
Create `src/components/settings/ApiKeysManager.tsx`:
```typescript
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Plus, Copy, Trash2, Eye, EyeOff } from 'lucide-react';
import { toast } from '@/components/ui/use-toast';
import { api } from '@/lib/api';

export function ApiKeysManager() {
  const [showDialog, setShowDialog] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set());
  const queryClient = useQueryClient();

  const { data: apiKeys, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: api.getApiKeys,
  });

  const createKeyMutation = useMutation({
    mutationFn: api.createApiKey,
    onSuccess: (data) => {
      queryClient.invalidateQueries(['api-keys']);
      toast({
        title: 'API key created',
        description: 'Copy the key now as it won\'t be shown again',
      });
      // Auto-copy to clipboard
      navigator.clipboard.writeText(data.key);
      setShowDialog(false);
      setNewKeyName('');
    },
  });

  const deleteKeyMutation = useMutation({
    mutationFn: api.deleteApiKey,
    onSuccess: () => {
      queryClient.invalidateQueries(['api-keys']);
      toast({
        title: 'API key deleted',
        description: 'The API key has been revoked',
      });
    },
  });

  const toggleKeyVisibility = (id: string) => {
    const newVisible = new Set(visibleKeys);
    if (newVisible.has(id)) {
      newVisible.delete(id);
    } else {
      newVisible.add(id);
    }
    setVisibleKeys(newVisible);
  };

  const maskKey = (key: string) => {
    return key.substring(0, 8) + '...' + key.substring(key.length - 8);
  };

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>API Keys</CardTitle>
            <Button onClick={() => setShowDialog(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Generate Key
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Last Used</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {apiKeys?.map((key: any) => (
                <TableRow key={key.id}>
                  <TableCell className="font-medium">{key.name}</TableCell>
                  <TableCell className="font-mono text-sm">
                    {visibleKeys.has(key.id) ? key.key : maskKey(key.key)}
                  </TableCell>
                  <TableCell>{new Date(key.created).toLocaleDateString()}</TableCell>
                  <TableCell>
                    {key.lastUsed ? new Date(key.lastUsed).toLocaleDateString() : 'Never'}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => toggleKeyVisibility(key.id)}
                      >
                        {visibleKeys.has(key.id) ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigator.clipboard.writeText(key.key)}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteKeyMutation.mutate(key.id)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={showDialog} onOpenChange={setShowDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Generate API Key</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="key-name">Key Name</Label>
              <Input
                id="key-name"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="e.g., Production Server"
              />
            </div>
            <Button
              onClick={() => createKeyMutation.mutate(newKeyName)}
              disabled={!newKeyName || createKeyMutation.isLoading}
            >
              Generate
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
```

### 5. Proxy Settings Component
Create `src/components/settings/ProxySettings.tsx`:
```typescript
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/ui/use-toast';
import { api } from '@/lib/api';

export function ProxySettings() {
  const [proxyUrl, setProxyUrl] = useState('');
  const queryClient = useQueryClient();

  const { data: currentProxy } = useQuery({
    queryKey: ['proxy-url'],
    queryFn: api.getProxyUrl,
    onSuccess: (data) => {
      setProxyUrl(data || '');
    },
  });

  const updateProxyMutation = useMutation({
    mutationFn: api.setProxyUrl,
    onSuccess: () => {
      queryClient.invalidateQueries(['proxy-url']);
      toast({
        title: 'Proxy updated',
        description: proxyUrl ? 'Proxy URL has been set' : 'Proxy has been disabled',
      });
    },
  });

  const handleSave = () => {
    updateProxyMutation.mutate(proxyUrl);
  };

  const handleRemove = () => {
    setProxyUrl('');
    updateProxyMutation.mutate('');
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Proxy Configuration</CardTitle>
        <CardDescription>
          Configure an HTTP/HTTPS proxy for outbound requests
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="proxy-url">Proxy URL</Label>
          <Input
            id="proxy-url"
            type="url"
            value={proxyUrl}
            onChange={(e) => setProxyUrl(e.target.value)}
            placeholder="http://proxy.example.com:8080"
          />
        </div>

        <div className="flex gap-2">
          <Button onClick={handleSave}>Save</Button>
          <Button variant="outline" onClick={handleRemove}>
            Remove Proxy
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
```

### 6. API Extensions
Update `src/lib/api.ts`:
```typescript
// Add to APIClient class

// Boolean settings
async getDebug() {
  const { data } = await this.client.get<APIResponse<boolean>>('/debug');
  return data.data;
}

async setDebug(value: boolean) {
  await this.client.put('/debug', { value });
}

async getLoggingToFile() {
  const { data } = await this.client.get<APIResponse<boolean>>('/logging-to-file');
  return data.data;
}

async setLoggingToFile(value: boolean) {
  await this.client.put('/logging-to-file', { value });
}

async getUsageStatisticsEnabled() {
  const { data } = await this.client.get<APIResponse<boolean>>('/usage-statistics-enabled');
  return data.data;
}

async setUsageStatisticsEnabled(value: boolean) {
  await this.client.put('/usage-statistics-enabled', { value });
}

// API Keys
async getApiKeys() {
  const { data } = await this.client.get<APIResponse<any>>('/api-keys');
  return data.data;
}

async createApiKey(name: string) {
  const { data } = await this.client.put<APIResponse<any>>('/api-keys', { name });
  return data.data;
}

async deleteApiKey(id: string) {
  await this.client.delete(`/api-keys/${id}`);
}

// Proxy
async getProxyUrl() {
  const { data } = await this.client.get<APIResponse<string>>('/proxy-url');
  return data.data;
}

async setProxyUrl(url: string) {
  await this.client.put('/proxy-url', { url });
}
```

## Todo List
- [ ] Create main Settings page with tabs
- [ ] Build Monaco-based YAML editor
- [ ] Implement boolean settings toggles
- [ ] Create API key management interface
- [ ] Add proxy configuration UI
- [ ] Implement config export/import
- [ ] Add YAML validation feedback
- [ ] Create save/reset functionality
- [ ] Add change detection for unsaved changes

## Success Criteria
- [ ] YAML editor has syntax highlighting
- [ ] Config changes validate before save
- [ ] Boolean toggles update immediately
- [ ] API keys can be created and deleted
- [ ] Config export downloads YAML file
- [ ] Import loads new configuration
- [ ] Unsaved changes show warning
- [ ] Proxy settings persist correctly