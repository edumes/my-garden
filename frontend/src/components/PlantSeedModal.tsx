import { Clock, Droplet, Star, Zap } from 'lucide-react';
import { useState } from 'react';
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

interface PlantType {
  id: string;
  name: string;
  description: string;
  icon: string;
  rarity: string;
  season: string;
  weather: string;
  growth_time: number;
  water_needs: number;
  fertilizer_needs: number;
  min_level: number;
  yield: number;
  harvest_value: number;
  experience_value: number;
}

interface PlantSeedModalProps {
  onClose: () => void;
  onSubmit: (plantTypeId: string) => void;
  plantTypes: PlantType[];
  position: number;
  open: boolean;
}

const rarityColors = {
  common: 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 border-gray-300 dark:border-gray-600',
  uncommon: 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200 border-green-300 dark:border-green-600',
  rare: 'bg-blue-100 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200 border-blue-300 dark:border-blue-600',
  epic: 'bg-purple-100 dark:bg-purple-900/20 text-purple-800 dark:text-purple-200 border-purple-300 dark:border-purple-600',
  legendary: 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 border-yellow-300 dark:border-yellow-600',
};

export function PlantSeedModal({ onClose, onSubmit, plantTypes, position, open }: PlantSeedModalProps) {
  const [selectedPlantType, setSelectedPlantType] = useState<string | null>(null);

  const handleSubmit = () => {
    if (selectedPlantType) {
      onSubmit(selectedPlantType);
      setSelectedPlantType(null);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      onClose();
      setSelectedPlantType(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] sm:max-h-128 flex flex-col">
        <DialogHeader>
          <DialogTitle>Plant Seed - Position {position + 1}</DialogTitle>
          <DialogDescription>
            Choose a plant type to plant in this position. Each plant has different growth requirements and rewards.
          </DialogDescription>
        </DialogHeader>
        
        <div className="overflow-y-auto max-h-[60vh] sm:max-h-80 flex-1 py-4">
          <div className="grid grid-cols-1 gap-3 sm:gap-4">
            {plantTypes.map((plantType) => (
              <div
                key={plantType.id}
                onClick={() => setSelectedPlantType(plantType.id)}
                className={`p-3 sm:p-4 border-2 rounded-lg cursor-pointer transition-all hover:shadow-md ${
                  selectedPlantType === plantType.id
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
                  <span
                    className={`text-xs px-2 py-1 rounded-full border ${
                      rarityColors[plantType.rarity as keyof typeof rarityColors]
                    }`}
                  >
                    {plantType.rarity}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                  <div className="flex items-center space-x-1">
                    <Clock className="w-3 h-3" />
                    <span>{plantType.growth_time} days</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Star className="w-3 h-3" />
                    <span>{plantType.experience_value} XP</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Droplet className="w-3 h-3" />
                    <span>{plantType.water_needs}% water</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Zap className="w-3 h-3" />
                    <span>{plantType.fertilizer_needs}% fertilizer</span>
                  </div>
                </div>

                <div className="mt-2 pt-2 border-t border-border">
                  <div className="flex justify-between text-xs">
                    <span className="text-muted-foreground">Yield: {plantType.yield}</span>
                    <span className="text-green-600 dark:text-green-400 font-medium">
                      ${plantType.harvest_value} each
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleSubmit}
            disabled={!selectedPlantType}
            className="bg-green-600 hover:bg-green-700"
          >
            Plant Seed
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}