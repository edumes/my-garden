import { useEffect, useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';

import { SelectTrigger } from '@radix-ui/react-select';
import { Activity, BarChart3, BlocksIcon, BrainIcon, Clock, CpuIcon, DatabaseIcon, GlobeIcon, LayoutIcon, LineChartIcon, NetworkIcon, SearchIcon, ServerIcon, User } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../services/api';
import SearchSelect from './SearchSelect';
import { Badge } from './ui/badge';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select, SelectContent, SelectItem, SelectValue } from './ui/select';

interface AuditLog {
    id: string;
    user_id?: string;
    action: string;
    resource: string;
    resource_id?: string;
    details: any;
    ip_address: string;
    user_agent: string;
    status: string;
    created_at: string;
    user?: {
        username: string;
        email: string;
    };
}

interface AuditStats {
    total_actions: number;
    successful_actions: number;
    failed_actions: number;
    error_actions: number;
    unique_users: number;
    top_actions: Array<{ action: string; count: number }>;
    top_resources: Array<{ resource: string; count: number }>;
}

export function AuditLogs() {
    const [logs, setLogs] = useState<AuditLog[]>([]);
    const [stats, setStats] = useState<AuditStats | null>(null);
    const [loading, setLoading] = useState(false);
    const [filters, setFilters] = useState({
        action: '',
        resource: '',
        status: '',
        start_date: '',
        end_date: '',
        limit: 50,
        offset: 0,
        service_category: '',
    });
    const [total, setTotal] = useState(0);
    const [view, setView] = useState<'logs' | 'stats'>('logs');

    const loadAuditLogs = async () => {
        setLoading(true);
        try {
            const response = await apiService.getAuditLogs(filters) as any;
            setLogs(response.logs);
            setTotal(response.total);
        } catch (error) {
            console.error('Failed to load audit logs:', error);
            toast.error('Failed to load audit logs');
        } finally {
            setLoading(false);
        }
    };

    const loadStats = async () => {
        setLoading(true);
        try {
            const response = await apiService.getAuditStats(30) as any;
            setStats(response.stats);
        } catch (error) {
            console.error('Failed to load audit stats:', error);
            toast.error('Failed to load audit stats');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (view === 'logs') {
            loadAuditLogs();
        } else {
            loadStats();
        }
    }, [view, filters]);

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleString();
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'success':
                return 'bg-green-100 text-green-800';
            case 'failure':
                return 'bg-red-100 text-red-800';
            case 'error':
                return 'bg-orange-100 text-orange-800';
            default:
                return 'bg-gray-100 text-gray-800';
        }
    };

    const getActionIcon = (action: string) => {
        switch (action) {
            case 'user_register':
            case 'user_login':
            case 'user_logout':
                return <User className="w-4 h-4" />;
            case 'garden_create':
            case 'garden_update':
            case 'garden_delete':
            case 'garden_view':
                return <Activity className="w-4 h-4" />;
            case 'plant_seed':
            case 'plant_harvest':
            case 'plant_remove':
                return <Activity className="w-4 h-4" />;
            case 'seed_purchase':
                return <BarChart3 className="w-4 h-4" />;
            default:
                return <Activity className="w-4 h-4" />;
        }
    };

    return (
        <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8">
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <h1 className="text-2xl font-bold">Audit System</h1>
                    <div className="flex space-x-2">
                        <Button
                            variant={view === 'logs' ? 'default' : 'outline'}
                            onClick={() => setView('logs')}
                        >
                            <Activity className="w-4 h-4 mr-2" />
                            Audit Logs
                        </Button>
                        <Button
                            variant={view === 'stats' ? 'default' : 'outline'}
                            onClick={() => setView('stats')}
                        >
                            <BarChart3 className="w-4 h-4 mr-2" />
                            Statistics
                        </Button>
                    </div>
                </div>

                {view === 'logs' && (
                    <>
                        <Card>
                            <CardHeader>
                                <CardTitle>Filters</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                                    <div>
                                        <Label htmlFor="action">Action</Label>
                                        <Select
                                            value={filters.action}
                                            onValueChange={(value) => setFilters({ ...filters, action: value })}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="All actions" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="all">All actions</SelectItem>
                                                <SelectItem value="user_register">User Register</SelectItem>
                                                <SelectItem value="user_login">User Login</SelectItem>
                                                <SelectItem value="garden_create">Garden Create</SelectItem>
                                                <SelectItem value="plant_seed">Plant Seed</SelectItem>
                                                <SelectItem value="plant_harvest">Plant Harvest</SelectItem>
                                                <SelectItem value="seed_purchase">Seed Purchase</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                    <div>
                                        <Label htmlFor="resource">Resource</Label>
                                        <Select
                                            value={filters.resource}
                                            onValueChange={(value) => setFilters({ ...filters, resource: value })}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="All resources" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="all">All resources</SelectItem>
                                                <SelectItem value="user">User</SelectItem>
                                                <SelectItem value="garden">Garden</SelectItem>
                                                <SelectItem value="plant">Plant</SelectItem>
                                                <SelectItem value="store">Store</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                    <div>
                                        <SearchSelect
                                            items={[
                                                { value: 'all', label: 'All', icon: LineChartIcon, number: 2451 },
                                                { value: 'success', label: 'Success', icon: BrainIcon, number: 1832 },
                                                { value: 'failure', label: 'Failure', icon: DatabaseIcon, number: 1654 },
                                                { value: 'error', label: 'Error', icon: CpuIcon, number: 943 },
                                            ]}
                                            value={filters.status}
                                            onChange={(value) => setFilters({ ...filters, status: value })}
                                            label="Status"
                                            placeholder="All statuses"
                                        />
                                    </div>
                                </div>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                                    <div>
                                        <Label htmlFor="start_date">Start Date</Label>
                                        <Input
                                            type="date"
                                            value={filters.start_date}
                                            onChange={(e) => setFilters({ ...filters, start_date: e.target.value })}
                                        />
                                    </div>
                                    <div>
                                        <Label htmlFor="end_date">End Date</Label>
                                        <Input
                                            type="date"
                                            value={filters.end_date}
                                            onChange={(e) => setFilters({ ...filters, end_date: e.target.value })}
                                        />
                                    </div>
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader>
                                <CardTitle>
                                    Audit Logs ({total} total)
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                {loading ? (
                                    <div className="text-center py-8">Loading...</div>
                                ) : (
                                    <div className="space-y-4">
                                        {logs.map((log) => (
                                            <div key={log.id} className="border rounded-lg p-4 space-y-2">
                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center space-x-2">
                                                        {getActionIcon(log.action)}
                                                        <span className="font-medium">{log.action}</span>
                                                        <Badge className={getStatusColor(log.status)}>
                                                            {log.status}
                                                        </Badge>
                                                    </div>
                                                    <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                                                        <Clock className="w-4 h-4" />
                                                        {formatDate(log.created_at)}
                                                    </div>
                                                </div>
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                                                    <div>
                                                        <span className="font-medium">Resource:</span> {log.resource}
                                                        {log.resource_id && (
                                                            <span className="ml-2 text-muted-foreground">({log.resource_id})</span>
                                                        )}
                                                    </div>
                                                    <div>
                                                        <span className="font-medium">User:</span>{' '}
                                                        {log.user ? log.user.username : 'Anonymous'}
                                                    </div>
                                                    <div>
                                                        <span className="font-medium">IP:</span> {log.ip_address}
                                                    </div>
                                                </div>
                                                {log.details && (
                                                    <div className="mt-2">
                                                        <span className="font-medium">Details:</span>
                                                        <pre className="mt-1 text-xs bg-muted p-2 rounded overflow-x-auto">
                                                            {JSON.stringify(log.details, null, 2)}
                                                        </pre>
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </>
                )}

                {view === 'stats' && stats && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Total Actions</CardTitle>
                                <Activity className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{stats.total_actions}</div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Successful</CardTitle>
                                <Activity className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold text-green-600">{stats.successful_actions}</div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Failed</CardTitle>
                                <Activity className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold text-red-600">{stats.failed_actions}</div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Unique Users</CardTitle>
                                <User className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{stats.unique_users}</div>
                            </CardContent>
                        </Card>

                        <Card className="md:col-span-2">
                            <CardHeader>
                                <CardTitle>Top Actions</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    {stats.top_actions.map((action) => (
                                        <div key={action.action} className="flex justify-between items-center">
                                            <span className="text-sm">{action.action}</span>
                                            <Badge>{action.count}</Badge>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="md:col-span-2">
                            <CardHeader>
                                <CardTitle>Top Resources</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    {stats.top_resources.map((resource) => (
                                        <div key={resource.resource} className="flex justify-between items-center">
                                            <span className="text-sm">{resource.resource}</span>
                                            <Badge>{resource.count}</Badge>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                )}
            </div>
        </div>
    );
} 