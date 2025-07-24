import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Clock, Coins, Package, ShoppingCart, Star } from 'lucide-react';
import { useEffect, useState } from 'react';
import { formatGrowthTime } from '../lib/utils';
import { apiService } from '../services/api';
import { PlantType, SeedInventory } from '../types/api';

interface PlantSeedModalProps {
  onClose: () => void;
  onPlant: (plantType: PlantType) => void;
  open: boolean;
  onOpenStore?: () => void;
}

export function PlantSeedModal({ onClose, onPlant, open, onOpenStore }: PlantSeedModalProps) {
  const [selectedPlantType, setSelectedPlantType] = useState<PlantType | null>(null);
  const [plantTypes, setPlantTypes] = useState<PlantType[]>([]);
  const [userSeedInventory, setUserSeedInventory] = useState<SeedInventory[]>([]);
  const [loadingInventory, setLoadingInventory] = useState(false);

  useEffect(() => {
    if (open) {
      loadUserSeedInventory();
      loadPlantTypes();
    }
  }, [open]);

  const loadUserSeedInventory = async () => {
    setLoadingInventory(true);
    try {
      const response = await apiService.getUserSeedInventory();
      setUserSeedInventory(response.seed_inventory);
    } catch (error) {
      console.error('Failed to load user seed inventory:', error);
    } finally {
      setLoadingInventory(false);
    }
  };

  const loadPlantTypes = async () => {
    try {
      const response = await apiService.getPlantTypes();
      setPlantTypes(response.plant_types);
    } catch (error) {
      console.error('Failed to load plant types:', error);
    }
  };

  const getAvailableSeeds = () => {
    const availablePlantTypes: (PlantType & { availableQuantity: number })[] = [];
    
    plantTypes.forEach(plantType => {
      const seedItem = userSeedInventory.find(item => item.plant_type_id === plantType.id);
      if (seedItem && seedItem.quantity > 0) {
        availablePlantTypes.push({
          ...plantType,
          availableQuantity: seedItem.quantity
        });
      }
    });
    
    return availablePlantTypes;
  };

  const handleSubmit = () => {
    if (selectedPlantType) {
      onPlant(selectedPlantType);
      setSelectedPlantType(null);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      onClose();
      setSelectedPlantType(null);
    }
  };

  const availableSeeds = getAvailableSeeds();

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] sm:max-h-128 flex flex-col">
        <DialogHeader>
          <DialogTitle>Plant Seed</DialogTitle>
          <DialogDescription>
            Choose a plant type to plant in this position. You can only plant seeds you have purchased.
            {onOpenStore && (
              <div className="mt-2 p-2 bg-blue-900/20 rounded-lg">
                <div className="flex items-center space-x-2 text-sm text-blue-700 dark:text-blue-300">
                  <ShoppingCart className="w-4 h-4" />
                  <span>Need more seeds? Visit the store to buy them!</span>
                </div>
              </div>
            )}
          </DialogDescription>
        </DialogHeader>
        
        <div className="flex-1 overflow-y-auto">
          {loadingInventory ? (
            <div className="text-center py-8">
              <Package className="w-8 h-8 mx-auto mb-4 animate-pulse" />
              <p className="text-muted-foreground">Loading your seeds...</p>
            </div>
          ) : availableSeeds.length === 0 ? (
            <div className="text-center py-8">
              <Package className="w-8 h-8 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground mb-4">You don't have any seeds!</p>
              {onOpenStore && (
                <Button onClick={() => {
                  onOpenStore();
                  onClose();
                }} className="flex items-center space-x-2 mx-auto">
                  <ShoppingCart className="w-4 h-4" />
                  <span>Buy Seeds</span>
                </Button>
              )}
            </div>
          ) :
            <div className="grid grid-cols-1 gap-3 sm:gap-4">
              {availableSeeds.map((plantType) => (
                <div
                  key={plantType.id}
                  onClick={() => setSelectedPlantType(plantType)}
                  className={`p-3 sm:p-4 border-2 rounded-lg cursor-pointer transition-all hover:shadow-md ${
                    selectedPlantType?.id === plantType.id
                      ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
                      : 'border-border hover:border-muted-foreground'
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <span className="text-xl sm:text-2xl">{plantType.icon}</span>
                      <div>
                        <h4 className="text-sm sm:text-base font-medium text-foreground">{plantType.name}</h4>
                        <p className="text-sm text-muted-foreground">{plantType.description}</p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-1 bg-green-100 dark:bg-green-900/20 px-2 py-1 rounded-full">
                      <Package className="w-3 h-3 text-green-600" />
                      <span className="text-xs font-medium text-green-700 dark:text-green-300">
                        {plantType.availableQuantity}
                      </span>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                    <div className="flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>{formatGrowthTime(plantType.growth_time)}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <Star className="w-3 h-3" />
                      <span>Yield: {plantType.yield}</span>
                    </div>
                  </div>

                  <div className="mt-2 pt-2 border-t border-border">
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-muted-foreground">Harvest Value</span>
                      <span className="text-green-600 dark:text-green-400 font-medium">
                        {plantType.harvest_value} coins
                      </span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-muted-foreground">Seed Price</span>
                      <span className="text-yellow-600 dark:text-yellow-400 font-medium flex items-center space-x-1">
                        <Coins className="w-3 h-3" />
                        <span>{plantType.seed_price} coins</span>
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          }
        </div>

        <DialogFooter>
          {onOpenStore && (
            <Button
              onClick={() => {
                onOpenStore();
                onClose();
              }}
              variant="outline"
              className="flex items-center space-x-2"
            >
              <ShoppingCart className="w-4 h-4" />
              <span>Store</span>
            </Button>
          )}
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleSubmit}
            disabled={!selectedPlantType || availableSeeds.length === 0}
            className="bg-green-600 hover:bg-green-700"
          >
            Plant Seed
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}