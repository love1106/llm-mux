# Phase 4: Accounts Page

## Context
Build the Accounts management page for handling OAuth authentication flows and auth file management.

## Overview
Create a comprehensive interface for managing authentication accounts, including OAuth flow initiation, status tracking, and auth file management with provider-specific configurations.

## Requirements
- Display all auth files with provider information
- OAuth flow UI with start/status/cancel operations
- Provider configuration management
- Real-time status updates during OAuth flow
- Batch operations for auth file management

## Implementation Steps

### 1. Accounts Page Component
Create `src/pages/AccountsPage.tsx`:
```typescript
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Plus, RefreshCw, Trash2, Download } from 'lucide-react';
import { api } from '@/lib/api';
import { OAuthFlowDialog } from '@/components/accounts/OAuthFlowDialog';
import { AuthFilesList } from '@/components/accounts/AuthFilesList';

export function AccountsPage() {
  const [oauthDialogOpen, setOAuthDialogOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: authFiles, isLoading } = useQuery({
    queryKey: ['auth-files'],
    queryFn: api.getAuthFiles,
    refetchInterval: 5000, // Poll every 5 seconds
  });

  const { data: providers } = useQuery({
    queryKey: ['providers'],
    queryFn: api.getProviders,
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteAuthFile,
    onSuccess: () => {
      queryClient.invalidateQueries(['auth-files']);
    },
  });

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Accounts</h1>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries(['auth-files'])}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => setOAuthDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Account
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Accounts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{authFiles?.length || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Active Providers</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {[...new Set(authFiles?.map(f => f.provider))].length || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">OAuth Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant="success">Connected</Badge>
          </CardContent>
        </Card>
      </div>

      {/* Auth Files List */}
      <AuthFilesList
        authFiles={authFiles || []}
        isLoading={isLoading}
        onDelete={(filename) => deleteMutation.mutate(filename)}
      />

      {/* OAuth Flow Dialog */}
      <OAuthFlowDialog
        open={oauthDialogOpen}
        onClose={() => setOAuthDialogOpen(false)}
        providers={providers || []}
      />
    </div>
  );
}
```

### 2. Auth Files List Component
Create `src/components/accounts/AuthFilesList.tsx`:
```typescript
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Trash2, Download, CheckCircle, XCircle, Clock } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface AuthFile {
  filename: string;
  provider: string;
  email?: string;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'pending';
}

interface Props {
  authFiles: AuthFile[];
  isLoading: boolean;
  onDelete: (filename: string) => void;
}

export function AuthFilesList({ authFiles, isLoading, onDelete }: Props) {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'expired':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-yellow-500" />;
    }
  };

  const getProviderColor = (provider: string) => {
    const colors: Record<string, string> = {
      openai: 'bg-green-500',
      anthropic: 'bg-purple-500',
      google: 'bg-blue-500',
      azure: 'bg-cyan-500',
    };
    return colors[provider.toLowerCase()] || 'bg-gray-500';
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <Card>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Provider</TableHead>
            <TableHead>Account</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Expires</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {authFiles.map((file) => (
            <TableRow key={file.filename}>
              <TableCell>
                <div className="flex items-center gap-2">
                  <div className={`w-2 h-2 rounded-full ${getProviderColor(file.provider)}`} />
                  <span className="font-medium">{file.provider}</span>
                </div>
              </TableCell>
              <TableCell>{file.email || file.filename}</TableCell>
              <TableCell>
                <div className="flex items-center gap-1">
                  {getStatusIcon(file.status)}
                  <Badge variant={file.status === 'active' ? 'success' : 'secondary'}>
                    {file.status}
                  </Badge>
                </div>
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatDistanceToNow(new Date(file.created_at), { addSuffix: true })}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {file.expires_at
                  ? formatDistanceToNow(new Date(file.expires_at), { addSuffix: true })
                  : 'Never'}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => window.open(`/v1/management/auth-files/download/${file.filename}`)}
                  >
                    <Download className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(file.filename)}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
          {authFiles.length === 0 && (
            <TableRow>
              <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                No auth files found. Click "Add Account" to connect a provider.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Card>
  );
}
```

### 3. OAuth Flow Dialog
Create `src/components/accounts/OAuthFlowDialog.tsx`:
```typescript
import { useState, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, CheckCircle, ExternalLink } from 'lucide-react';
import { api } from '@/lib/api';

interface Props {
  open: boolean;
  onClose: () => void;
  providers: string[];
}

export function OAuthFlowDialog({ open, onClose, providers }: Props) {
  const [selectedProvider, setSelectedProvider] = useState('');
  const [oauthState, setOAuthState] = useState<string | null>(null);
  const [status, setStatus] = useState<'idle' | 'pending' | 'success' | 'error'>('idle');
  const [authUrl, setAuthUrl] = useState<string | null>(null);

  const startOAuthMutation = useMutation({
    mutationFn: api.startOAuth,
    onSuccess: (data) => {
      setOAuthState(data.state);
      setAuthUrl(data.auth_url);
      setStatus('pending');
      window.open(data.auth_url, '_blank');
    },
  });

  useEffect(() => {
    if (!oauthState || status !== 'pending') return;

    const interval = setInterval(async () => {
      try {
        const result = await api.getOAuthStatus(oauthState);
        if (result.status === 'completed') {
          setStatus('success');
          clearInterval(interval);
          setTimeout(onClose, 2000);
        } else if (result.status === 'failed') {
          setStatus('error');
          clearInterval(interval);
        }
      } catch (error) {
        // Still pending
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [oauthState, status]);

  const handleStart = () => {
    if (selectedProvider) {
      startOAuthMutation.mutate(selectedProvider);
    }
  };

  const handleCancel = async () => {
    if (oauthState) {
      await api.cancelOAuth(oauthState);
    }
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Account</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {status === 'idle' && (
            <>
              <div>
                <Label htmlFor="provider">Provider</Label>
                <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {providers.map((provider) => (
                      <SelectItem key={provider} value={provider}>
                        {provider}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <Button onClick={handleStart} disabled={!selectedProvider} className="w-full">
                Start OAuth Flow
              </Button>
            </>
          )}

          {status === 'pending' && (
            <div className="space-y-4">
              <Alert>
                <AlertDescription className="space-y-2">
                  <p>Please complete the authentication in your browser.</p>
                  {authUrl && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => window.open(authUrl, '_blank')}
                    >
                      <ExternalLink className="h-4 w-4 mr-2" />
                      Open Auth Page
                    </Button>
                  )}
                </AlertDescription>
              </Alert>

              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-8 w-8 animate-spin" />
              </div>

              <Button variant="outline" onClick={handleCancel} className="w-full">
                Cancel
              </Button>
            </div>
          )}

          {status === 'success' && (
            <Alert className="border-green-500 bg-green-50">
              <CheckCircle className="h-4 w-4 text-green-500" />
              <AlertDescription>
                Account successfully connected! Closing...
              </AlertDescription>
            </Alert>
          )}

          {status === 'error' && (
            <Alert variant="destructive">
              <AlertDescription>
                Authentication failed. Please try again.
              </AlertDescription>
            </Alert>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

### 4. API Extensions
Update `src/lib/api.ts`:
```typescript
// Add to APIClient class

async getProviders() {
  const { data } = await this.client.get<APIResponse<string[]>>('/providers');
  return data.data;
}

async cancelOAuth(state: string) {
  await this.client.post(`/oauth/cancel/${state}`);
}

async downloadAuthFile(filename: string) {
  const { data } = await this.client.get(`/auth-files/download/${filename}`, {
    responseType: 'blob',
  });

  // Create download link
  const url = window.URL.createObjectURL(new Blob([data]));
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  link.remove();
}
```

## Todo List
- [ ] Create main Accounts page component
- [ ] Build auth files list with table view
- [ ] Implement OAuth flow dialog with provider selection
- [ ] Add OAuth status polling during authentication
- [ ] Create delete confirmation dialog
- [ ] Implement batch operations for auth files
- [ ] Add auth file download functionality
- [ ] Create provider-specific icons and colors
- [ ] Add expiry status tracking

## Success Criteria
- [ ] Auth files display in sortable table
- [ ] OAuth flow completes successfully
- [ ] Real-time status updates during OAuth
- [ ] Auth files can be deleted with confirmation
- [ ] Provider information displays correctly
- [ ] Expiry dates show relative time
- [ ] Download auth files works