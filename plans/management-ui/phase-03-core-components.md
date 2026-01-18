# Phase 3: Core Components

## Context
Build the foundational components for the Management UI including layout, navigation, API client, and shared utilities.

## Overview
Create the core infrastructure components that all pages will use, including the main layout with navigation, API client with authentication, and common UI patterns.

## Requirements
- Responsive layout with sidebar navigation
- API client with X-Management-Key authentication
- Global error handling and toast notifications
- Loading states and suspense boundaries
- Dark mode support

## Implementation Steps

### 1. API Client Setup
Create `src/lib/api.ts`:
```typescript
import axios, { AxiosInstance } from 'axios';
import { toast } from '@/components/ui/use-toast';

interface APIResponse<T> {
  data: T;
  meta: {
    timestamp: string;
    version: string;
  };
}

class APIClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: '/v1/management',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor for auth
    this.client.interceptors.request.use((config) => {
      const apiKey = localStorage.getItem('management-key');
      if (apiKey) {
        config.headers['X-Management-Key'] = apiKey;
      }
      return config;
    });

    // Response interceptor for errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          toast({
            title: 'Authentication Required',
            description: 'Please provide a valid management key',
            variant: 'destructive',
          });
        }
        return Promise.reject(error);
      }
    );
  }

  // Configuration endpoints
  async getConfig() {
    const { data } = await this.client.get<APIResponse<any>>('/config');
    return data.data;
  }

  async getConfigYAML() {
    const { data } = await this.client.get('/config.yaml', {
      headers: { Accept: 'application/yaml' },
    });
    return data;
  }

  async updateConfigYAML(yaml: string) {
    await this.client.put('/config.yaml', yaml, {
      headers: { 'Content-Type': 'application/yaml' },
    });
  }

  // Auth files endpoints
  async getAuthFiles() {
    const { data } = await this.client.get<APIResponse<any>>('/auth-files');
    return data.data;
  }

  async deleteAuthFile(filename: string) {
    await this.client.delete(`/auth-files/${filename}`);
  }

  // OAuth endpoints
  async startOAuth(provider: string) {
    const { data } = await this.client.post<APIResponse<any>>('/oauth/start', {
      provider,
    });
    return data.data;
  }

  async getOAuthStatus(state: string) {
    const { data } = await this.client.get<APIResponse<any>>(`/oauth/status/${state}`);
    return data.data;
  }

  // Usage endpoints
  async getUsage(params?: { from?: string; to?: string; groupBy?: string }) {
    const { data } = await this.client.get<APIResponse<any>>('/usage', { params });
    return data.data;
  }

  // Logs endpoints
  async getLogs(params?: { limit?: number; level?: string }) {
    const { data } = await this.client.get<APIResponse<any>>('/logs', { params });
    return data.data;
  }
}

export const api = new APIClient();
```

### 2. Main Layout Component
Create `src/components/layout/MainLayout.tsx`:
```typescript
import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  BarChart3,
  FileText,
  Settings,
  Key,
} from 'lucide-react';
import { cn } from '@/lib/utils';

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Accounts', href: '/accounts', icon: Users },
  { name: 'Usage', href: '/usage', icon: BarChart3 },
  { name: 'Logs', href: '/logs', icon: FileText },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function MainLayout() {
  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <div className="w-64 border-r bg-card">
        <div className="flex h-16 items-center px-6 border-b">
          <h1 className="text-xl font-bold">llm-mux</h1>
        </div>
        <nav className="p-4 space-y-1">
          {navigation.map((item) => (
            <NavLink
              key={item.name}
              to={item.href}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'hover:bg-muted'
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.name}
            </NavLink>
          ))}
        </nav>

        {/* API Key Status */}
        <div className="absolute bottom-4 left-4 right-4">
          <ApiKeyStatus />
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-auto">
        <Outlet />
      </div>
    </div>
  );
}
```

### 3. API Key Management Component
Create `src/components/ApiKeyStatus.tsx`:
```typescript
import { useState, useEffect } from 'react';
import { Key, Check, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { api } from '@/lib/api';

export function ApiKeyStatus() {
  const [isOpen, setIsOpen] = useState(false);
  const [apiKey, setApiKey] = useState('');
  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    const key = localStorage.getItem('management-key');
    if (key) {
      setApiKey(key);
      checkKeyValidity(key);
    }
  }, []);

  const checkKeyValidity = async (key: string) => {
    try {
      localStorage.setItem('management-key', key);
      await api.getConfig();
      setIsValid(true);
    } catch {
      setIsValid(false);
    }
  };

  const handleSave = () => {
    localStorage.setItem('management-key', apiKey);
    checkKeyValidity(apiKey);
    setIsOpen(false);
  };

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        className="w-full justify-start"
        onClick={() => setIsOpen(true)}
      >
        <Key className="h-4 w-4 mr-2" />
        API Key
        {isValid ? (
          <Check className="h-4 w-4 ml-auto text-green-500" />
        ) : (
          <X className="h-4 w-4 ml-auto text-red-500" />
        )}
      </Button>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Management API Key</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="api-key">API Key</Label>
              <Input
                id="api-key"
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter your management API key"
              />
            </div>
            <Button onClick={handleSave}>Save</Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
```

### 4. React Query Setup
Create `src/lib/query-client.ts`:
```typescript
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000, // 30 seconds
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});
```

### 5. Global Store Setup
Create `src/stores/global.ts`:
```typescript
import { create } from 'zustand';

interface GlobalState {
  theme: 'light' | 'dark' | 'system';
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}

export const useGlobalStore = create<GlobalState>((set) => ({
  theme: 'system',
  setTheme: (theme) => set({ theme }),
  sidebarCollapsed: false,
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
}));
```

### 6. Error Boundary Component
Create `src/components/ErrorBoundary.tsx`:
```typescript
import { Component, ReactNode } from 'react';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center min-h-screen">
          <Card className="p-6 max-w-md">
            <div className="flex items-center gap-3 mb-4">
              <AlertCircle className="h-6 w-6 text-destructive" />
              <h2 className="text-lg font-semibold">Something went wrong</h2>
            </div>
            <p className="text-sm text-muted-foreground mb-4">
              {this.state.error?.message || 'An unexpected error occurred'}
            </p>
            <Button onClick={() => window.location.reload()}>
              Reload Page
            </Button>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### 7. Loading Component
Create `src/components/Loading.tsx`:
```typescript
import { Loader2 } from 'lucide-react';

export function Loading() {
  return (
    <div className="flex items-center justify-center p-8">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );
}

export function LoadingPage() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <Loader2 className="h-12 w-12 animate-spin text-muted-foreground" />
    </div>
  );
}
```

## Todo List
- [ ] Create API client with authentication interceptors
- [ ] Build main layout with responsive sidebar
- [ ] Implement API key management component
- [ ] Setup React Query for data fetching
- [ ] Create global Zustand store
- [ ] Build error boundary component
- [ ] Create loading components
- [ ] Add toast notification system
- [ ] Implement dark mode toggle

## Success Criteria
- [ ] Layout renders with navigation sidebar
- [ ] API client authenticates with X-Management-Key
- [ ] Error handling shows user-friendly messages
- [ ] Loading states display during data fetching
- [ ] API key can be set and validated
- [ ] Navigation highlights active route
- [ ] Responsive design works on mobile