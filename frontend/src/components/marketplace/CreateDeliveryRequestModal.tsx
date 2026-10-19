import { useState } from 'react';
import { Button } from '../ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';
import { Label } from '../ui/label';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Loader } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { PlantType } from '../../types/api';

interface CreateDeliveryRequestModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  plantTypes: PlantType[];
  onSuccess: () => void;
}

export function CreateDeliveryRequestModal({
  open,
  onOpenChange,
  plantTypes,
  onSuccess
}: CreateDeliveryRequestModalProps) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    plant_type_id: '',
    reward: '',
    deadline: ''
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.plant_type_id || !formData.reward || !formData.deadline) {
      toast.error('Please fill in all fields');
      return;
    }

    const reward = parseInt(formData.reward);
    if (isNaN(reward) || reward <= 0) {
      toast.error('Please enter a valid reward amount');
      return;
    }

    const deadline = new Date(formData.deadline);
    if (deadline <= new Date()) {
      toast.error('Deadline must be in the future');
      return;
    }

    try {
      setLoading(true);
      await apiService.createDeliveryRequest({
        plant_type_id: formData.plant_type_id,
        reward,
        deadline: deadline.toISOString()
      });
      
      toast.success('Delivery request created successfully');
      setFormData({ plant_type_id: '', reward: '', deadline: '' });
      onSuccess();
    } catch (error) {
      console.error('Failed to create delivery request:', error);
      toast.error('Failed to create delivery request');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setFormData({ plant_type_id: '', reward: '', deadline: '' });
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Delivery Request</DialogTitle>
          <DialogDescription>
            Request a specific plant to be delivered to your garden. Set a reward and deadline.
          </DialogDescription>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="plant_type">Plant Type</Label>
            <Select
              value={formData.plant_type_id}
              onValueChange={(value) => setFormData(prev => ({ ...prev, plant_type_id: value }))}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a plant type" />
              </SelectTrigger>
              <SelectContent>
                {plantTypes.map((plantType) => (
                  <SelectItem key={plantType.id} value={plantType.id}>
                    <div className="flex items-center gap-2">
                      <span>{plantType.name}</span>
                      <span className="text-muted-foreground text-xs">
                        ({plantType.seed_price} coins)
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reward">Reward (coins)</Label>
            <Input
              id="reward"
              type="number"
              placeholder="Enter reward amount"
              value={formData.reward}
              onChange={(e) => setFormData(prev => ({ ...prev, reward: e.target.value }))}
              min="1"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="deadline">Deadline</Label>
            <Input
              id="deadline"
              type="datetime-local"
              value={formData.deadline}
              onChange={(e) => setFormData(prev => ({ ...prev, deadline: e.target.value }))}
              min={new Date().toISOString().slice(0, 16)}
            />
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
            <Button type="submit" disabled={loading}>
              {loading && <Loader className="w-4 h-4 mr-2 animate-spin" />}
              Create Request
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
} 