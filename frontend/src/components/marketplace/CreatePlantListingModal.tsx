import { useState } from 'react';
import { Button } from '../ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';
import { Label } from '../ui/label';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Badge } from '../ui/badge';
import { Loader } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { Plant } from '../../types/api';

interface CreatePlantListingModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userPlants: Plant[];
  onSuccess: () => void;
}

export function CreatePlantListingModal({
  open,
  onOpenChange,
  userPlants,
  onSuccess
}: CreatePlantListingModalProps) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    plant_id: '',
    price: '',
    expires_at: ''
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.plant_id || !formData.price || !formData.expires_at) {
      toast.error('Please fill in all fields');
      return;
    }

    const price = parseInt(formData.price);
    if (isNaN(price) || price <= 0) {
      toast.error('Please enter a valid price');
      return;
    }

    const expiresAt = new Date(formData.expires_at);
    if (expiresAt <= new Date()) {
      toast.error('Expiration date must be in the future');
      return;
    }

    try {
      setLoading(true);
      await apiService.createPlantListing({
        plant_id: formData.plant_id,
        price,
        expires_at: expiresAt.toISOString()
      });
      
      toast.success('Plant listing created successfully');
      setFormData({ plant_id: '', price: '', expires_at: '' });
      onSuccess();
    } catch (error) {
      console.error('Failed to create plant listing:', error);
      toast.error('Failed to create plant listing');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setFormData({ plant_id: '', price: '', expires_at: '' });
      onOpenChange(false);
    }
  };

  const availablePlants = userPlants.filter(plant => 
    plant.stage === 'harvestable' || plant.stage === 'mature'
  );

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>List Plant for Sale</DialogTitle>
          <DialogDescription>
            Choose a plant from your garden to list for sale in the marketplace.
          </DialogDescription>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="plant">Select Plant</Label>
            <Select
              value={formData.plant_id}
              onValueChange={(value) => setFormData(prev => ({ ...prev, plant_id: value }))}
            >
              <SelectTrigger>
                <SelectValue placeholder="Choose a plant to list" />
              </SelectTrigger>
              <SelectContent>
                {availablePlants.length === 0 ? (
                  <SelectItem value="" disabled>
                    No harvestable plants available
                  </SelectItem>
                ) : (
                  availablePlants.map((plant) => (
                    <SelectItem key={plant.id} value={plant.id}>
                      <div className="flex items-center justify-between w-full">
                        <span>{plant.plant_type.name}</span>
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="capitalize text-xs">
                            {plant.stage}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            Garden: {plant.garden.name}
                          </span>
                        </div>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            {availablePlants.length === 0 && (
              <p className="text-sm text-muted-foreground">
                You need harvestable or mature plants to create listings.
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="price">Price (coins)</Label>
            <Input
              id="price"
              type="number"
              placeholder="Enter price in coins"
              value={formData.price}
              onChange={(e) => setFormData(prev => ({ ...prev, price: e.target.value }))}
              min="1"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="expires_at">Expiration Date</Label>
            <Input
              id="expires_at"
              type="datetime-local"
              value={formData.expires_at}
              onChange={(e) => setFormData(prev => ({ ...prev, expires_at: e.target.value }))}
              min={new Date().toISOString().slice(0, 16)}
            />
            <p className="text-xs text-muted-foreground">
              Listings expire after 7 days by default
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={loading || availablePlants.length === 0}
            >
              {loading && <Loader className="w-4 h-4 mr-2 animate-spin" />}
              Create Listing
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
} 