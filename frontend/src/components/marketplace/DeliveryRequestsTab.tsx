import { useCallback, useEffect, useState } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Loader, Package, Clock, CheckCircle, Plus, Search } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { DeliveryRequest, PlantType, RequestStatus } from '../../types/api';
import { CreateDeliveryRequestModal } from './CreateDeliveryRequestModal';

interface DeliveryRequestsTabProps {
  onRefresh: () => void;
}

export function DeliveryRequestsTab({ onRefresh }: DeliveryRequestsTabProps) {
  const [requests, setRequests] = useState<DeliveryRequest[]>([]);
  const [plantTypes, setPlantTypes] = useState<PlantType[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  // Status filtering happens on the client; the API returns open requests only.
  const loadData = useCallback(() =>
    Promise.all([
      apiService.getDeliveryRequests(),
      apiService.getPlantTypes()
    ])
      .then(([requestsResponse, plantTypesResponse]) => {
        setRequests(requestsResponse.delivery_requests);
        setPlantTypes(plantTypesResponse.plant_types);
      })
      .catch((error) => {
        console.error('Failed to load delivery requests:', error);
        toast.error('Failed to load delivery requests');
      })
      .finally(() => setLoading(false)),
  []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const reload = () => {
    setLoading(true);
    return loadData();
  };

  const handleAcceptRequest = async (requestId: string) => {
    try {
      await apiService.acceptDeliveryRequest(requestId);
      toast.success('Delivery request accepted successfully');
      reload();
      onRefresh();
    } catch (error) {
      console.error('Failed to accept delivery request:', error);
      toast.error(`Failed to accept delivery request: ${(error as Error).message}`);
    }
  };

  const handleCompleteRequest = async (requestId: string) => {
    try {
      const { reward_paid } = await apiService.completeDeliveryRequest(requestId);
      toast.success(`Delivery completed: you earned ${reward_paid} coins`);
      reload();
      onRefresh();
    } catch (error) {
      console.error('Failed to complete delivery request:', error);
      toast.error(`Failed to complete delivery request: ${(error as Error).message}`);
    }
  };

  const getStatusBadge = (status: RequestStatus) => {
    const statusConfig = {
      open: { variant: 'default' as const, text: 'Open', icon: Package },
      accepted: { variant: 'secondary' as const, text: 'Accepted', icon: Clock },
      completed: { variant: 'default' as const, text: 'Completed', icon: CheckCircle },
    };
    const config = statusConfig[status];
    const Icon = config.icon;
    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.text}
      </Badge>
    );
  };

  const filteredRequests = requests.filter(request => {
    const matchesSearch = request.plant_type.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         request.requester.username.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesFilter = filterStatus === 'all' || request.status === filterStatus;
    return matchesSearch && matchesFilter;
  });

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader className="w-6 h-6 text-green-600 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold">Delivery Requests</h2>
          <p className="text-muted-foreground">Fulfill plant delivery requests and earn rewards</p>
        </div>
        <Button onClick={() => setShowCreateModal(true)} className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          Create Request
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by plant type or requester..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={filterStatus} onValueChange={setFilterStatus}>
          <SelectTrigger className="w-full sm:w-48">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="open">Open</SelectItem>
            <SelectItem value="accepted">Accepted</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Requests Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredRequests.map((request) => (
          <Card key={request.id} className="hover:shadow-md transition-shadow">
            <CardHeader>
              <div className="flex justify-between items-start">
                <div>
                  <CardTitle className="text-lg">{request.plant_type.name}</CardTitle>
                  <CardDescription>Requested by {request.requester.username}</CardDescription>
                </div>
                {getStatusBadge(request.status)}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Reward:</span>
                <span className="font-semibold text-green-600">{request.reward} coins</span>
              </div>
              
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Deadline:</span>
                <span className="text-sm">{new Date(request.deadline).toLocaleDateString()}</span>
              </div>

              {request.acceptor && (
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Accepted by:</span>
                  <span className="text-sm">{request.acceptor.username}</span>
                </div>
              )}

              <div className="pt-2">
                {request.status === 'open' && (
                  <Button 
                    onClick={() => handleAcceptRequest(request.id)}
                    className="w-full"
                    size="sm"
                  >
                    Accept Request
                  </Button>
                )}
                
                {request.status === 'accepted' && request.acceptor && (
                  <Button 
                    onClick={() => handleCompleteRequest(request.id)}
                    className="w-full"
                    size="sm"
                    variant="secondary"
                  >
                    Complete Delivery
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredRequests.length === 0 && (
        <div className="text-center py-8">
          <Package className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">No delivery requests found</h3>
          <p className="text-muted-foreground">
            {searchTerm || filterStatus !== 'all' 
              ? 'Try adjusting your search or filters'
              : 'Be the first to create a delivery request!'
            }
          </p>
        </div>
      )}

      {/* Create Request Modal */}
      <CreateDeliveryRequestModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        plantTypes={plantTypes}
        onSuccess={() => {
          setShowCreateModal(false);
          reload();
          onRefresh();
        }}
      />
    </div>
  );
} 