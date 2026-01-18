# Phase 5: Usage Statistics

## Context
Build the Usage Statistics page with interactive charts and data visualization for monitoring LLM usage patterns.

## Overview
Create comprehensive usage analytics with time-series charts, account-level breakdowns, and export capabilities using Recharts for visualization.

## Requirements
- Time-series charts for requests and tokens
- Account-level usage breakdown
- Time range filtering (hour, day, week, month)
- Data export to CSV/JSON
- Real-time updates with polling

## Implementation Steps

### 1. Usage Page Component
Create `src/pages/UsagePage.tsx`:
```typescript
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Download, RefreshCw } from 'lucide-react';
import { api } from '@/lib/api';
import { UsageChart } from '@/components/usage/UsageChart';
import { AccountUsageTable } from '@/components/usage/AccountUsageTable';
import { UsageStats } from '@/components/usage/UsageStats';
import { format, subHours, subDays, subWeeks, subMonths } from 'date-fns';

type TimeRange = '1h' | '24h' | '7d' | '30d';
type GroupBy = 'minute' | 'hour' | 'day';

export function UsagePage() {
  const [timeRange, setTimeRange] = useState<TimeRange>('24h');
  const [groupBy, setGroupBy] = useState<GroupBy>('hour');

  const getTimeParams = () => {
    const now = new Date();
    let from: Date;
    let grouping: GroupBy;

    switch (timeRange) {
      case '1h':
        from = subHours(now, 1);
        grouping = 'minute';
        break;
      case '24h':
        from = subDays(now, 1);
        grouping = 'hour';
        break;
      case '7d':
        from = subWeeks(now, 1);
        grouping = 'day';
        break;
      case '30d':
        from = subMonths(now, 1);
        grouping = 'day';
        break;
    }

    return {
      from: from.toISOString(),
      to: now.toISOString(),
      groupBy: grouping,
    };
  };

  const { data: usage, isLoading, refetch } = useQuery({
    queryKey: ['usage', timeRange],
    queryFn: () => api.getUsage(getTimeParams()),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  const handleExport = (format: 'csv' | 'json') => {
    if (!usage) return;

    const filename = `usage-${format(new Date(), 'yyyy-MM-dd-HHmm')}.${format}`;

    if (format === 'json') {
      const blob = new Blob([JSON.stringify(usage, null, 2)], { type: 'application/json' });
      downloadBlob(blob, filename);
    } else {
      const csv = convertToCSV(usage);
      const blob = new Blob([csv], { type: 'text/csv' });
      downloadBlob(blob, filename);
    }
  };

  const downloadBlob = (blob: Blob, filename: string) => {
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Usage Statistics</h1>
        <div className="flex items-center gap-2">
          <Select value={timeRange} onValueChange={(v: TimeRange) => setTimeRange(v)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1h">Last Hour</SelectItem>
              <SelectItem value="24h">Last 24h</SelectItem>
              <SelectItem value="7d">Last 7 Days</SelectItem>
              <SelectItem value="30d">Last 30 Days</SelectItem>
            </SelectContent>
          </Select>

          <Button variant="outline" size="icon" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4" />
          </Button>

          <Button variant="outline" onClick={() => handleExport('csv')}>
            <Download className="h-4 w-4 mr-2" />
            CSV
          </Button>

          <Button variant="outline" onClick={() => handleExport('json')}>
            <Download className="h-4 w-4 mr-2" />
            JSON
          </Button>
        </div>
      </div>

      {/* Summary Stats */}
      <UsageStats usage={usage} isLoading={isLoading} />

      {/* Charts */}
      <Tabs defaultValue="requests" className="space-y-4">
        <TabsList>
          <TabsTrigger value="requests">Requests</TabsTrigger>
          <TabsTrigger value="tokens">Tokens</TabsTrigger>
          <TabsTrigger value="accounts">By Account</TabsTrigger>
          <TabsTrigger value="models">By Model</TabsTrigger>
        </TabsList>

        <TabsContent value="requests">
          <Card>
            <CardHeader>
              <CardTitle>Request Volume</CardTitle>
            </CardHeader>
            <CardContent>
              <UsageChart
                data={usage?.timeSeries || []}
                dataKey="requests"
                color="#8884d8"
                isLoading={isLoading}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="tokens">
          <Card>
            <CardHeader>
              <CardTitle>Token Usage</CardTitle>
            </CardHeader>
            <CardContent>
              <UsageChart
                data={usage?.timeSeries || []}
                dataKey="tokens"
                color="#82ca9d"
                isLoading={isLoading}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="accounts">
          <Card>
            <CardHeader>
              <CardTitle>Usage by Account</CardTitle>
            </CardHeader>
            <CardContent>
              <AccountUsageTable data={usage?.byAccount || []} isLoading={isLoading} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="models">
          <Card>
            <CardHeader>
              <CardTitle>Usage by Model</CardTitle>
            </CardHeader>
            <CardContent>
              <ModelUsageChart data={usage?.byModel || []} isLoading={isLoading} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

### 2. Usage Chart Component
Create `src/components/usage/UsageChart.tsx`:
```typescript
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { format } from 'date-fns';
import { Loading } from '@/components/Loading';

interface Props {
  data: Array<{
    timestamp: string;
    [key: string]: any;
  }>;
  dataKey: string;
  color: string;
  isLoading: boolean;
}

export function UsageChart({ data, dataKey, color, isLoading }: Props) {
  if (isLoading) {
    return <Loading />;
  }

  const formattedData = data.map(item => ({
    ...item,
    time: format(new Date(item.timestamp), 'HH:mm'),
  }));

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={formattedData}>
        <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
        <XAxis
          dataKey="time"
          className="text-xs"
        />
        <YAxis className="text-xs" />
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '6px',
          }}
        />
        <Line
          type="monotone"
          dataKey={dataKey}
          stroke={color}
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 4 }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

### 3. Usage Stats Component
Create `src/components/usage/UsageStats.tsx`:
```typescript
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface Props {
  usage: any;
  isLoading: boolean;
}

export function UsageStats({ usage, isLoading }: Props) {
  if (isLoading || !usage) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Loading...</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-8 bg-muted rounded animate-pulse" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  const stats = [
    {
      title: 'Total Requests',
      value: usage.totalRequests?.toLocaleString() || '0',
      change: usage.requestsChange || 0,
    },
    {
      title: 'Total Tokens',
      value: usage.totalTokens?.toLocaleString() || '0',
      change: usage.tokensChange || 0,
    },
    {
      title: 'Avg Tokens/Request',
      value: Math.round(usage.totalTokens / usage.totalRequests || 0).toLocaleString(),
      change: 0,
    },
    {
      title: 'Active Accounts',
      value: usage.activeAccounts || '0',
      change: usage.accountsChange || 0,
    },
  ];

  const getTrendIcon = (change: number) => {
    if (change > 0) return <TrendingUp className="h-4 w-4 text-green-500" />;
    if (change < 0) return <TrendingDown className="h-4 w-4 text-red-500" />;
    return <Minus className="h-4 w-4 text-muted-foreground" />;
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      {stats.map((stat) => (
        <Card key={stat.title}>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {stat.title}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="text-2xl font-bold">{stat.value}</div>
              <div className="flex items-center gap-1">
                {getTrendIcon(stat.change)}
                <span className="text-xs text-muted-foreground">
                  {Math.abs(stat.change)}%
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
```

### 4. Account Usage Table
Create `src/components/usage/AccountUsageTable.tsx`:
```typescript
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Progress } from '@/components/ui/progress';

interface AccountUsage {
  account: string;
  provider: string;
  requests: number;
  tokens: number;
  percentage: number;
}

interface Props {
  data: AccountUsage[];
  isLoading: boolean;
}

export function AccountUsageTable({ data, isLoading }: Props) {
  if (isLoading) {
    return <div>Loading...</div>;
  }

  const maxTokens = Math.max(...data.map(d => d.tokens), 1);

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Account</TableHead>
          <TableHead>Provider</TableHead>
          <TableHead className="text-right">Requests</TableHead>
          <TableHead className="text-right">Tokens</TableHead>
          <TableHead>Usage</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map((item) => (
          <TableRow key={item.account}>
            <TableCell className="font-medium">{item.account}</TableCell>
            <TableCell>{item.provider}</TableCell>
            <TableCell className="text-right">{item.requests.toLocaleString()}</TableCell>
            <TableCell className="text-right">{item.tokens.toLocaleString()}</TableCell>
            <TableCell className="w-40">
              <div className="flex items-center gap-2">
                <Progress value={(item.tokens / maxTokens) * 100} className="flex-1" />
                <span className="text-xs text-muted-foreground">
                  {item.percentage.toFixed(1)}%
                </span>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

### 5. Model Usage Pie Chart
Create `src/components/usage/ModelUsageChart.tsx`:
```typescript
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';

interface Props {
  data: Array<{
    model: string;
    requests: number;
    tokens: number;
  }>;
  isLoading: boolean;
}

const COLORS = ['#8884d8', '#82ca9d', '#ffc658', '#ff8042', '#8dd1e1'];

export function ModelUsageChart({ data, isLoading }: Props) {
  if (isLoading) {
    return <div>Loading...</div>;
  }

  const chartData = data.map((item, index) => ({
    name: item.model,
    value: item.tokens,
    fill: COLORS[index % COLORS.length],
  }));

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={chartData}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          outerRadius={80}
          label
        >
          {chartData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
}
```

### 6. CSV Export Utility
Create `src/lib/csv-export.ts`:
```typescript
export function convertToCSV(data: any): string {
  if (!data || !data.timeSeries || data.timeSeries.length === 0) {
    return '';
  }

  const headers = Object.keys(data.timeSeries[0]);
  const csv = [
    headers.join(','),
    ...data.timeSeries.map((row: any) =>
      headers.map(header => {
        const value = row[header];
        return typeof value === 'string' && value.includes(',')
          ? `"${value}"`
          : value;
      }).join(',')
    )
  ].join('\n');

  return csv;
}
```

## Todo List
- [ ] Create main Usage page with time range filters
- [ ] Build time-series line charts with Recharts
- [ ] Implement usage statistics cards
- [ ] Create account usage breakdown table
- [ ] Add model usage pie chart
- [ ] Implement CSV and JSON export
- [ ] Add real-time polling for updates
- [ ] Create responsive chart layouts
- [ ] Add trend indicators to stats

## Success Criteria
- [ ] Charts display time-series data correctly
- [ ] Time range filtering updates charts
- [ ] Export to CSV/JSON works
- [ ] Account breakdown shows percentages
- [ ] Model usage pie chart renders
- [ ] Real-time updates every 30 seconds
- [ ] Responsive design on mobile devices