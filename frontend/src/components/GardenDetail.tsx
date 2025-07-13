import { ArrowLeft, Calendar, Droplet, TreePine, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Garden } from '../types/api';
import { GardenGrid } from './GardenGrid';
import { PlantSeedModal } from './PlantSeedModal';
import { Button } from './ui/button';

interface GardenDetailProps {
  garden: Garden;
  onBack: () => void;
}

export function GardenDetail({ garden: initialGarden, onBack }: GardenDetailProps) {
  const [garden, setGarden] = useState<Garden>(initialGarden);
  const [showPlantModal, setShowPlantModal] = useState(false);
  const [selectedPosition, setSelectedPosition] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [plantTypes, setPlantTypes] = useState<import('../types/api').PlantType[]>([]);
  const [plantTypesLoading, setPlantTypesLoading] = useState(true);

  useEffect(() => {
    loadGardenDetails();
    loadPlantTypes();
  }, []);

  const loadGardenDetails = async () => {
    try {
      const response = await apiService.getGarden(garden.id);
      setGarden(response.garden);
    } catch (error) {
      console.error('Failed to load garden details:', error);
    }
  };

  const loadPlantTypes = async () => {
    setPlantTypesLoading(true);
    try {
      const response = await apiService.getPlantTypes();
      setPlantTypes(response.plant_types);
    } catch (error) {
      console.error('Failed to load plant types:', error);
    } finally {
      setPlantTypesLoading(false);
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
    <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8 space-y-4 sm:space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            onClick={onBack}
            variant="secondary"
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
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
          plantTypes={plantTypes}
          position={selectedPosition!}
        />
      )}
      {plantTypesLoading && (
        <div className="text-center text-muted-foreground">Loading plant types...</div>
      )}
    </div>
  );
}