import React, { useState } from 'react';
import { Garden } from '../types/api';
import { GardenGrid } from './GardenGrid';
import { WeatherWidget } from './WeatherWidget';
import { ChatBox } from './ui/chat-box';
import { toast } from 'sonner';
import { apiService } from '../services/api';

interface GardenDetailProps {
  garden: Garden;
}

export const GardenDetail: React.FC<GardenDetailProps> = ({ garden }) => {
  const [positionLocks, setPositionLocks] = useState<Record<number, { user_id: string; username: string; action: 'planting' | 'harvesting' }>>({});

  const handlePlantAction = async (plantId: string, action: 'water' | 'fertilize' | 'harvest' | 'remove') => {
    try {
      if (action === 'harvest') {
        const response = await apiService.harvestPlant(garden.id, plantId);
        toast.success(`Successfully harvested plant! Earned ${response.harvest.coins_earned} coins and ${response.harvest.xp_earned} XP`);
      } else if (action === 'remove') {
        await apiService.removePlant(garden.id, plantId);
        toast.success('Plant removed successfully');
      }
    } catch (error) {
      toast.error(`Failed to ${action} plant`);
    }
  };

  const handlePlantSeed = async (position: number) => {
    // This will be handled by the PlantSeedModal in the actual implementation
    toast.info('Opening seed selection...');
  };

  return (
    <div className="container mx-auto p-4 space-y-6">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold">{garden.name}</h1>
          {garden.description && (
            <p className="text-muted-foreground mt-2">{garden.description}</p>
          )}
        </div>
        <WeatherWidget />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <GardenGrid
            plants={garden.plants || []}
            onPlantAction={handlePlantAction}
            onPlantSeed={handlePlantSeed}
            positionLocks={positionLocks}
          />
        </div>
        <div className="lg:col-span-1">
          <ChatBox gardenId={garden.id} />
        </div>
      </div>
    </div>
  );
};