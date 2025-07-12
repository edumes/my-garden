import { ArrowLeft, Calendar, Droplet, TreePine, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Garden } from '../types/api';
import { GardenGrid } from './GardenGrid';
import { PlantSeedModal } from './PlantSeedModal';

interface GardenDetailProps {
  garden: Garden;
  onBack: () => void;
}

// Mock plant types for demo
const MOCK_PLANT_TYPES = [
  {
    id: '57b89191-3429-453d-9718-14f3c67d83f8',
    name: 'Tomato',
    description: 'Juicy red tomatoes',
    icon: '🍅',
    rarity: 'common',
    season: 'summer',
    weather: 'sunny',
    growth_time: 7,
    water_needs: 80,
    fertilizer_needs: 60,
    min_level: 1,
    yield: 3,
    harvest_value: 10,
    experience_value: 15,
  },
  {
    id: 'd904633e-9317-43a3-af0a-abfcb11cd0ea',
    name: 'Carrot',
    description: 'Crunchy orange carrots',
    icon: '🥕',
    rarity: 'common',
    season: 'spring',
    weather: 'cloudy',
    growth_time: 5,
    water_needs: 60,
    fertilizer_needs: 40,
    min_level: 1,
    yield: 2,
    harvest_value: 8,
    experience_value: 12,
  },
  {
    id: 'a81ab4f6-4a88-43e1-ac64-f14208cd220e',
    name: 'Sunflower',
    description: 'Bright yellow sunflowers',
    icon: '🌻',
    rarity: 'rare',
    season: 'summer',
    weather: 'sunny',
    growth_time: 10,
    water_needs: 70,
    fertilizer_needs: 50,
    min_level: 3,
    yield: 1,
    harvest_value: 25,
    experience_value: 30,
  },
];

export function GardenDetail({ garden: initialGarden, onBack }: GardenDetailProps) {
  const [garden, setGarden] = useState<Garden>(initialGarden);
  const [showPlantModal, setShowPlantModal] = useState(false);
  const [selectedPosition, setSelectedPosition] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadGardenDetails();
  }, []);

  const loadGardenDetails = async () => {
    try {
      const response = await apiService.getGarden(garden.id);
      setGarden(response.garden);
    } catch (error) {
      console.error('Failed to load garden details:', error);
    }
  };

  const handlePlantAction = async (plantId: string, action: 'water' | 'fertilize' | 'harvest' | 'remove') => {
    setLoading(true);
    try {
      const plant = garden.plants?.find(p => p.id === plantId);
      if (!plant) return;

      switch (action) {
        case 'water':
          await apiService.waterPlant(garden.id, plantId, 30);
          break;
        case 'fertilize':
          await apiService.fertilizePlant(garden.id, plantId, 20);
          break;
        case 'harvest':
          await apiService.harvestPlant(garden.id, plantId);
          break;
        case 'remove':
          await apiService.removePlant(garden.id, plantId);
          break;
      }

      await loadGardenDetails();
    } catch (error) {
      console.error(`Failed to ${action} plant:`, error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlantSeed = (position: number) => {
    setSelectedPosition(position);
    setShowPlantModal(true);
  };

  const handlePlantSubmit = async (plantTypeId: string) => {
    if (selectedPosition === null) return;

    setLoading(true);
    try {
      await apiService.plantSeed(garden.id, {
        plant_type_id: plantTypeId,
        position: selectedPosition,
      });

      setShowPlantModal(false);
      setSelectedPosition(null);
      await loadGardenDetails();
    } catch (error) {
      console.error('Failed to plant seed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4 sm:space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={onBack}
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold text-foreground">{garden.name}</h1>
            {garden.description && (
              <p className="text-sm sm:text-base text-muted-foreground">{garden.description}</p>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6">
        <div className="lg:col-span-3 order-2 lg:order-1">
          <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
            <h2 className="text-base sm:text-lg font-semibold text-foreground mb-4">Garden Grid</h2>
            <GardenGrid
              plants={garden.plants || []}
              onPlantAction={handlePlantAction}
              onPlantSeed={handlePlantSeed}
            />
          </div>
        </div>

        <div className="space-y-4 sm:space-y-6 order-1 lg:order-2">
          <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
            <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Garden Stats</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <TreePine className="w-4 h-4 text-green-500" />
                  <span className="text-sm text-muted-foreground">Plants</span>
                </div>
                <span className="text-sm sm:text-base font-medium">{garden.plants?.length || 0}/{garden.size}</span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Droplet className="w-4 h-4 text-blue-500" />
                  <span className="text-sm text-muted-foreground">Water Level</span>
                </div>
                <span className="text-sm sm:text-base font-medium">{garden.water_level}%</span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Zap className="w-4 h-4 text-yellow-500" />
                  <span className="text-sm text-muted-foreground">Fertilizer</span>
                </div>
                <span className="text-sm sm:text-base font-medium">{garden.fertilizer_level}%</span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Calendar className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Soil Quality</span>
                </div>
                <span className="text-sm sm:text-base font-medium">{garden.soil_quality}%</span>
              </div>
            </div>
          </div>

          {(garden.has_sprinkler || garden.has_greenhouse || garden.has_composter) && (
            <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
              <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Upgrades</h3>
              <div className="space-y-2">
                {garden.has_sprinkler && (
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                    <span className="text-sm text-muted-foreground">Auto Sprinkler System</span>
                  </div>
                )}
                {garden.has_greenhouse && (
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                    <span className="text-sm text-muted-foreground">Greenhouse Protection</span>
                  </div>
                )}
                {garden.has_composter && (
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-brown-500 rounded-full"></div>
                    <span className="text-sm text-muted-foreground">Compost System</span>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {showPlantModal && (
        <PlantSeedModal
          open={showPlantModal}
          onClose={() => {
            setShowPlantModal(false);
            setSelectedPosition(null);
          }}
          onSubmit={handlePlantSubmit}
          plantTypes={MOCK_PLANT_TYPES}
          position={selectedPosition!}
        />
      )}
    </div>
  );
}