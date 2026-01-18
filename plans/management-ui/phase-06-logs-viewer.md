# Phase 6: Logs Viewer

## Context
Build a real-time log viewer with filtering, search, and level-based highlighting for monitoring llm-mux operations.

## Overview
Create an interactive log viewer that streams logs in real-time, provides advanced filtering capabilities, and displays logs with appropriate formatting and highlighting based on severity levels.

## Requirements
- Real-time log streaming with polling
- Log level filtering (debug, info, warn, error)
- Text search across log messages
- Timestamp-based filtering
- Auto-scroll with pause capability
- Log export functionality

## Implementation Steps

### 1. Logs Page Component
Create `src/pages/LogsPage.tsx`:
```typescript
import { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Search, Download, RefreshCw, Trash2, Pause, Play } from 'lucide-react';
import { api } from '@/lib/api';
import { LogEntry } from '@/components/logs/LogEntry';
import { LogFilters } from '@/components/logs/LogFilters';
import { VirtualizedLogList } from '@/components/logs/VirtualizedLogList';

type LogLevel = 'all' | 'debug' | 'info' | 'warn' | 'error';

export function LogsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [logLevel, setLogLevel] = useState<LogLevel>('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const [isPaused, setIsPaused] = useState(false);
  const [limit, setLimit] = useState(1000);
  const containerRef = useRef<HTMLDivElement>(null);

  const { data: logs, isLoading, refetch } = useQuery({
    queryKey: ['logs', logLevel, limit],
    queryFn: () => api.getLogs({
      level: logLevel === 'all' ? undefined : logLevel,
      limit,
    }),
    refetchInterval: isPaused ? false : 2000, // Poll every 2 seconds when not paused
  });

  const { data: errorLogs } = useQuery({
    queryKey: ['error-logs'],
    queryFn: api.getRequestErrorLogs,
    refetchInterval: isPaused ? false : 5000,
  });

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const filteredLogs = logs?.filter((log: any) => {
    if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase())) {
      return false;
    }
    return true;
  }) || [];

  const handleExport = () => {
    const content = filteredLogs
      .map((log: any) => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join('\n');

    const blob = new Blob([content], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs-${new Date().toISOString()}.txt`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const handleClearLogs = async () => {
    if (confirm('Are you sure you want to clear all logs?')) {
      await api.deleteLogs();
      refetch();
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">Logs</h1>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setIsPaused(!isPaused)}
          >
            {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
          </Button>

          <Button
            variant="outline"
            size="icon"
            onClick={() => refetch()}
          >
            <RefreshCw className="h-4 w-4" />
          </Button>

          <Button
            variant="outline"
            onClick={handleExport}
          >
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>

          <Button
            variant="destructive"
            onClick={handleClearLogs}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Clear
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="mb-4">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search logs..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-9"
              />
            </div>

            <Select value={logLevel} onValueChange={(v: LogLevel) => setLogLevel(v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Levels</SelectItem>
                <SelectItem value="debug">Debug</SelectItem>
                <SelectItem value="info">Info</SelectItem>
                <SelectItem value="warn">Warning</SelectItem>
                <SelectItem value="error">Error</SelectItem>
              </SelectContent>
            </Select>

            <Select value={limit.toString()} onValueChange={(v) => setLimit(Number(v))}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="100">Last 100</SelectItem>
                <SelectItem value="500">Last 500</SelectItem>
                <SelectItem value="1000">Last 1000</SelectItem>
                <SelectItem value="5000">Last 5000</SelectItem>
              </SelectContent>
            </Select>

            <div className="flex items-center gap-2">
              <Switch
                id="auto-scroll"
                checked={autoScroll}
                onCheckedChange={setAutoScroll}
              />
              <Label htmlFor="auto-scroll">Auto-scroll</Label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Error Summary */}
      {errorLogs && errorLogs.length > 0 && (
        <Card className="mb-4 border-destructive">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm text-destructive">
              Recent Errors ({errorLogs.length})
            </CardTitle>
          </CardHeader>
          <CardContent className="max-h-32 overflow-auto">
            {errorLogs.slice(0, 3).map((error: any, i: number) => (
              <div key={i} className="text-sm text-muted-foreground mb-1">
                {error.message}
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Log List */}
      <Card className="flex-1 overflow-hidden">
        <CardContent className="p-0 h-full">
          <VirtualizedLogList
            logs={filteredLogs}
            isLoading={isLoading}
            containerRef={containerRef}
          />
        </CardContent>
      </Card>
    </div>
  );
}
```

### 2. Log Entry Component
Create `src/components/logs/LogEntry.tsx`:
```typescript
import { memo } from 'react';
import { cn } from '@/lib/utils';
import { format } from 'date-fns';

interface LogEntryData {
  id: string;
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  metadata?: Record<string, any>;
}

interface Props {
  log: LogEntryData;
  searchTerm?: string;
}

export const LogEntry = memo(function LogEntry({ log, searchTerm }: Props) {
  const getLevelColor = () => {
    switch (log.level) {
      case 'debug':
        return 'text-gray-500 bg-gray-50 dark:bg-gray-900';
      case 'info':
        return 'text-blue-600 bg-blue-50 dark:bg-blue-900/20';
      case 'warn':
        return 'text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20';
      case 'error':
        return 'text-red-600 bg-red-50 dark:bg-red-900/20';
    }
  };

  const getLevelBadgeColor = () => {
    switch (log.level) {
      case 'debug':
        return 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-300';
      case 'info':
        return 'bg-blue-200 text-blue-700 dark:bg-blue-700 dark:text-blue-200';
      case 'warn':
        return 'bg-yellow-200 text-yellow-700 dark:bg-yellow-700 dark:text-yellow-200';
      case 'error':
        return 'bg-red-200 text-red-700 dark:bg-red-700 dark:text-red-200';
    }
  };

  const highlightSearch = (text: string) => {
    if (!searchTerm) return text;

    const parts = text.split(new RegExp(`(${searchTerm})`, 'gi'));
    return parts.map((part, i) =>
      part.toLowerCase() === searchTerm.toLowerCase() ? (
        <mark key={i} className="bg-yellow-300 dark:bg-yellow-600">
          {part}
        </mark>
      ) : (
        part
      )
    );
  };

  return (
    <div className={cn('px-4 py-2 border-b font-mono text-sm', getLevelColor())}>
      <div className="flex items-start gap-3">
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {format(new Date(log.timestamp), 'HH:mm:ss.SSS')}
        </span>

        <span className={cn('px-2 py-0.5 rounded text-xs uppercase font-semibold', getLevelBadgeColor())}>
          {log.level}
        </span>

        <span className="flex-1 break-all">
          {highlightSearch(log.message)}
        </span>
      </div>

      {log.metadata && Object.keys(log.metadata).length > 0 && (
        <div className="mt-2 pl-28">
          <details className="cursor-pointer">
            <summary className="text-xs text-muted-foreground">
              Metadata ({Object.keys(log.metadata).length} fields)
            </summary>
            <pre className="mt-2 text-xs overflow-auto bg-muted p-2 rounded">
              {JSON.stringify(log.metadata, null, 2)}
            </pre>
          </details>
        </div>
      )}
    </div>
  );
});
```

### 3. Virtualized Log List
Create `src/components/logs/VirtualizedLogList.tsx`:
```typescript
import { useEffect, useRef } from 'react';
import { VariableSizeList as List } from 'react-window';
import { LogEntry } from './LogEntry';
import { Loading } from '@/components/Loading';

interface Props {
  logs: any[];
  isLoading: boolean;
  containerRef: React.RefObject<HTMLDivElement>;
  searchTerm?: string;
}

export function VirtualizedLogList({ logs, isLoading, containerRef, searchTerm }: Props) {
  const listRef = useRef<List>(null);
  const rowHeights = useRef<{ [key: number]: number }>({});

  const getItemSize = (index: number) => {
    // Base height for a single-line log entry
    const baseHeight = 40;
    const log = logs[index];

    if (!log) return baseHeight;

    // Add extra height for multi-line messages
    const lineCount = Math.ceil(log.message.length / 100);
    const metadataHeight = log.metadata ? 30 : 0;

    return baseHeight + (lineCount - 1) * 20 + metadataHeight;
  };

  useEffect(() => {
    // Reset row height cache when logs change
    rowHeights.current = {};
    listRef.current?.resetAfterIndex(0);
  }, [logs]);

  if (isLoading && logs.length === 0) {
    return <Loading />;
  }

  if (logs.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        No logs to display
      </div>
    );
  }

  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>
      <LogEntry log={logs[index]} searchTerm={searchTerm} />
    </div>
  );

  return (
    <div ref={containerRef} className="h-full">
      <List
        ref={listRef}
        height={600}
        itemCount={logs.length}
        itemSize={getItemSize}
        width="100%"
      >
        {Row}
      </List>
    </div>
  );
}
```

### 4. Log Filters Component
Create `src/components/logs/LogFilters.tsx`:
```typescript
import { DatePickerWithRange } from '@/components/ui/date-picker';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';

interface Filter {
  type: 'level' | 'date' | 'text';
  value: any;
}

interface Props {
  filters: Filter[];
  onFilterChange: (filters: Filter[]) => void;
}

export function LogFilters({ filters, onFilterChange }: Props) {
  const removeFilter = (index: number) => {
    onFilterChange(filters.filter((_, i) => i !== index));
  };

  const addFilter = (filter: Filter) => {
    onFilterChange([...filters, filter]);
  };

  return (
    <div className="flex flex-wrap gap-2">
      {filters.map((filter, i) => (
        <Badge key={i} variant="secondary" className="pl-3 pr-1">
          {filter.type}: {filter.value}
          <Button
            variant="ghost"
            size="sm"
            className="h-4 w-4 p-0 ml-2"
            onClick={() => removeFilter(i)}
          >
            <X className="h-3 w-3" />
          </Button>
        </Badge>
      ))}

      {filters.length > 0 && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onFilterChange([])}
        >
          Clear all
        </Button>
      )}
    </div>
  );
}
```

### 5. Log Statistics Component
Create `src/components/logs/LogStats.tsx`:
```typescript
import { Card, CardContent } from '@/components/ui/card';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend } from 'recharts';

interface Props {
  logs: any[];
}

const COLORS = {
  debug: '#6B7280',
  info: '#3B82F6',
  warn: '#F59E0B',
  error: '#EF4444',
};

export function LogStats({ logs }: Props) {
  const levelCounts = logs.reduce((acc, log) => {
    acc[log.level] = (acc[log.level] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const data = Object.entries(levelCounts).map(([level, count]) => ({
    name: level.toUpperCase(),
    value: count,
    fill: COLORS[level as keyof typeof COLORS],
  }));

  return (
    <Card>
      <CardContent className="pt-6">
        <ResponsiveContainer width="100%" height={200}>
          <PieChart>
            <Pie
              data={data}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={80}
              label
            >
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.fill} />
              ))}
            </Pie>
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
```

### 6. API Extensions
Update `src/lib/api.ts`:
```typescript
// Add to APIClient class

async getRequestErrorLogs() {
  const { data } = await this.client.get<APIResponse<any>>('/request-error-logs');
  return data.data;
}

async deleteLogs() {
  await this.client.delete('/logs');
}
```

## Todo List
- [ ] Create main Logs page with filtering
- [ ] Build log entry component with highlighting
- [ ] Implement virtualized list for performance
- [ ] Add real-time polling with pause/play
- [ ] Create search functionality with highlighting
- [ ] Build log level filtering
- [ ] Add log export to text file
- [ ] Implement log statistics visualization
- [ ] Add metadata expansion for detailed logs

## Success Criteria
- [ ] Logs stream in real-time
- [ ] Search filters logs instantly
- [ ] Log levels have distinct colors
- [ ] Auto-scroll works with toggle
- [ ] Export creates downloadable file
- [ ] Virtualized list handles 5000+ logs
- [ ] Pause stops polling updates