import { CheckCircle, DollarSign, Loader, Plus, Search, ShoppingCart, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { ListingStatus, Plant, PlantListing } from '../../types/api';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { CreatePlantListingModal } from './CreatePlantListingModal';

interface PlantListingsTabProps {
  onRefresh: () => void;
}

export function PlantListingsTab({ onRefresh }: PlantListingsTabProps) {
  const [listings, setListings] = useState<PlantListing[]>([]);
  const [userPlants, setUserPlants] = useState<Plant[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [priceRange, setPriceRange] = useState({ min: '', max: '' });
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [listingsResponse, gardensResponse] = await Promise.all([
        apiService.getPlantListings({ status: filterStatus === 'all' ? undefined : filterStatus }),
        apiService.getGardens()
      ]);
      setListings(listingsResponse.listings);

      // Get user's plants from all gardens
      const allPlants: Plant[] = [];
      gardensResponse.owned_gardens.forEach(garden => {
        if (garden.plants) {
          allPlants.push(...garden.plants);
        }
      });
      setUserPlants(allPlants);
    } catch (error) {
      console.error('Failed to load plant listings:', error);
      toast.error('Failed to load plant listings');
    } finally {
      setLoading(false);
    }
  };

  const handleBuyListing = async (listingId: string) => {
    try {
      // In a real implementation, you would sign the transaction here
      const signature = "mygarden"; // This should be generated using the user's private key
      await apiService.buyPlantListing(listingId, { signature });
      toast.success('Plant purchased successfully');
      loadData();
      onRefresh();
    } catch (error) {
      if (error instanceof Error) {
        console.error('Failed to buy plant listing: ', error.message);
        toast.error('Failed to buy plant listing: ' + error.message);
      }
    }
  };

  const getStatusBadge = (status: ListingStatus) => {
    const statusConfig = {
      active: { variant: 'default' as const, text: 'Active', icon: ShoppingCart },
      sold: { variant: 'secondary' as const, text: 'Sold', icon: CheckCircle },
      expired: { variant: 'destructive' as const, text: 'Expired', icon: XCircle },
      cancelled: { variant: 'destructive' as const, text: 'Cancelled', icon: XCircle },
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

  const filteredListings = listings.filter(listing => {
    const matchesSearch = listing.plant.plant_type.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      listing.seller.username.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesFilter = filterStatus === 'all' || listing.status === filterStatus;
    const matchesPrice = (!priceRange.min || listing.price >= parseInt(priceRange.min)) &&
      (!priceRange.max || listing.price <= parseInt(priceRange.max));
    return matchesSearch && matchesFilter && matchesPrice;
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
          <h2 className="text-2xl font-bold">Plant Listings</h2>
          <p className="text-muted-foreground">Buy and sell plants in the marketplace</p>
        </div>
        <Button onClick={() => setShowCreateModal(true)} className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          List Plant
        </Button>
      </div>

      {/* Filters */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by plant type or seller..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>

        <Select value={filterStatus} onValueChange={setFilterStatus}>
          <SelectTrigger>
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="sold">Sold</SelectItem>
            <SelectItem value="expired">Expired</SelectItem>
            <SelectItem value="cancelled">Cancelled</SelectItem>
          </SelectContent>
        </Select>

        <Input
          type="number"
          placeholder="Min price"
          value={priceRange.min}
          onChange={(e) => setPriceRange(prev => ({ ...prev, min: e.target.value }))}
        />

        <Input
          type="number"
          placeholder="Max price"
          value={priceRange.max}
          onChange={(e) => setPriceRange(prev => ({ ...prev, max: e.target.value }))}
        />
      </div>

      {/* Listings Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredListings.map((listing) => (
          <Card key={listing.id} className="hover:shadow-md transition-shadow">
            <CardHeader>
              <div className="flex justify-between items-start">
                <div>
                  <CardTitle className="text-lg">{listing.plant.plant_type.name}</CardTitle>
                  <CardDescription>Listed by {listing.seller.username}</CardDescription>
                </div>
                {getStatusBadge(listing.status)}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Price:</span>
                <span className="font-semibold text-green-600 flex items-center gap-1">
                  <DollarSign className="h-4 w-4" />
                  {listing.price} coins
                </span>
              </div>

              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Plant Stage:</span>
                <Badge variant="outline" className="capitalize">
                  {listing.plant.stage}
                </Badge>
              </div>

              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Expires:</span>
                <span className="text-sm">{new Date(listing.expires_at).toLocaleDateString()}</span>
              </div>

              <div className="pt-2">
                {listing.status === 'active' && (
                  <Button
                    onClick={() => handleBuyListing(listing.id)}
                    className="w-full"
                    size="sm"
                  >
                    Buy Plant
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredListings.length === 0 && (
        <div className="text-center py-8">
          <ShoppingCart className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">No plant listings found</h3>
          <p className="text-muted-foreground">
            {searchTerm || filterStatus !== 'all' || priceRange.min || priceRange.max
              ? 'Try adjusting your search or filters'
              : 'Be the first to list a plant for sale!'
            }
          </p>
        </div>
      )}

      {/* Create Listing Modal */}
      <CreatePlantListingModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        userPlants={userPlants}
        onSuccess={() => {
          setShowCreateModal(false);
          loadData();
          onRefresh();
        }}
      />
    </div>
  );
} 