import { ArrowLeft, Calendar, Clock, Coins, Leaf, Shield, TreePine, TrendingUp, User, Users } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';
import { Garden, PlantType } from '../types/api';
import { GardenGrid } from './GardenGrid';
import { PlantSeedModal } from './PlantSeedModal';
import { Button } from './ui/button';

interface GardenDetailProps {
  garden: Garden;
  permission?: string;
  isShared?: boolean;
  onBack: () => void;
  onOpenStore?: () => void;
}

export function GardenDetail({ garden: initialGarden, permission, isShared, onBack, onOpenStore }: GardenDetailProps) {
  const [garden, setGarden] = useState<Garden>(initialGarden);
  const [showPlantModal, setShowPlantModal] = useState(false);
  const [selectedPosition, setSelectedPosition] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [plantTypes, setPlantTypes] = useState<PlantType[]>([]);
  const [plantTypesLoading, setPlantTypesLoading] = useState(true);
  const { refreshUser } = useAuth();

  useEffect(() => {
    loadGardenDetails();
    loadPlantTypes();
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      loadGardenDetails();
    }, 2000);

    return () => clearInterval(interval);
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
    if (isShared) {
      if (action === 'harvest' && permission !== 'manage' && permission !== 'harvest') {
        toast.error('You don\'t have permission to harvest in this garden');
        return;
      }
      if (action === 'remove' && permission !== 'manage') {
        toast.error('You don\'t have permission to remove plants from this garden');
        return;
      }
    }

    setLoading(true);
    try {
      const plant = garden.plants?.find(p => p.id === plantId);
      if (!plant) return;

      switch (action) {
        case 'harvest':
          const harvestResponse = await apiService.harvestPlant(garden.id, plantId);
          await refreshUser();

          const message = `Harvested! +${harvestResponse.harvest.coins_earned} coins`;
          if (harvestResponse.harvest.level_up) {
            toast.success(`${message} | Level up! You're now level ${harvestResponse.harvest.new_level}`);
          } else {
            toast.success(message);
          }
          break;
        case 'remove':
          await apiService.removePlant(garden.id, plantId);
          toast.success('Plant removed successfully');
          break;
      }

      await loadGardenDetails();
    } catch (error) {
      console.error(`Failed to ${action} plant:`, error);
      toast.error(`Failed to ${action} plant`);
    } finally {
      setLoading(false);
    }
  };

  const handlePlantSeed = (position: number) => {
    if (isShared && permission !== 'manage' && permission !== 'plant') {
      toast.error('You don\'t have permission to plant in this garden');
      return;
    }
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
      toast.success('Seed planted successfully!');
    } catch (error) {
      console.error('Failed to plant seed:', error);
      toast.error('Failed to plant seed');
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
            <div className="flex items-center space-x-2">
              <h1 className="text-xl sm:text-2xl font-bold text-foreground">{garden.name}</h1>
              {isShared && (
                <div className="flex items-center space-x-1">
                  <Users className="w-4 h-4 text-blue-500" />
                  <span className="text-sm text-blue-600 dark:text-blue-400">Shared</span>
                </div>
              )}
              {permission && (
                <div className="flex items-center space-x-1">
                  <Shield className="w-4 h-4 text-green-500" />
                  <span className="text-sm text-green-600 dark:text-green-400 capitalize">{permission}</span>
                </div>
              )}
            </div>
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
              onPlantAction={(plantId, action) => handlePlantAction(plantId, action)}
              onPlantSeed={handlePlantSeed}
            />
          </div>
        </div>

        <div className="space-y-4 sm:space-y-6 order-1 lg:order-2">
          <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
            <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Garden Stats</h3>
            <div className="space-y-4">
              {/* Plant Statistics */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <TreePine className="w-4 h-4 text-green-500" />
                    <span className="text-sm text-muted-foreground">Total Plants</span>
                  </div>
                  <span className="text-sm sm:text-base font-medium">{garden.plants?.length || 0}</span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <Leaf className="w-4 h-4 text-yellow-500" />
                    <span className="text-sm text-muted-foreground">Ready to Harvest</span>
                  </div>
                  <span className="text-sm sm:text-base font-medium">
                    {garden.plants?.filter(p => p.stage === 'harvestable').length || 0}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <TrendingUp className="w-4 h-4 text-blue-500" />
                    <span className="text-sm text-muted-foreground">Growing</span>
                  </div>
                  <span className="text-sm sm:text-base font-medium">
                    {garden.plants?.filter(p => p.stage !== 'harvestable' && p.stage !== 'seed').length || 0}
                  </span>
                </div>
              </div>

              <div className="border-t border-border pt-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <Calendar className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Created</span>
                  </div>
                  <span className="text-sm sm:text-base font-medium">
                    {new Date(garden.created_at).toLocaleDateString()}
                  </span>
                </div>

                <div className="flex items-center justify-between mt-2">
                  <div className="flex items-center space-x-2">
                    <Clock className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Last Updated</span>
                  </div>
                  <span className="text-sm sm:text-base font-medium">
                    {new Date(garden.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </div>

              {/* User Stats */}
              <div className="border-t border-border pt-3">
                <div className="flex items-center space-x-2 mb-3">
                  <User className="w-4 h-4 text-purple-500" />
                  <span className="text-sm font-medium text-foreground">Gardener Stats</span>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">Level</span>
                    <span className="text-xs font-medium">{garden.user.level}</span>
                  </div>

                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">Experience</span>
                    <span className="text-xs font-medium">{garden.user.experience} XP</span>
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-1">
                      <Coins className="w-3 h-3 text-yellow-500" />
                      <span className="text-xs text-muted-foreground">Coins</span>
                    </div>
                    <span className="text-xs font-medium">{garden.user.coins}</span>
                  </div>
                </div>
              </div>

              {/* Plant Type Distribution */}
              {garden.plants && garden.plants.length > 0 && (
                <div className="border-t border-border pt-3">
                  <div className="flex items-center space-x-2 mb-3">
                    <Leaf className="w-4 h-4 text-green-500" />
                    <span className="text-sm font-medium text-foreground">Plant Types</span>
                  </div>

                  <div className="space-y-2">
                    {(() => {
                      const plantTypeCounts = garden.plants!.reduce((acc, plant) => {
                        const typeName = plant.plant_type.name;
                        acc[typeName] = (acc[typeName] || 0) + 1;
                        return acc;
                      }, {} as Record<string, number>);

                      return Object.entries(plantTypeCounts).map(([typeName, count]) => (
                        <div key={typeName} className="flex items-center justify-between">
                          <span className="text-xs text-muted-foreground">{typeName}</span>
                          <span className="text-xs font-medium">{count}</span>
                        </div>
                      ));
                    })()}
                  </div>
                </div>
              )}
            </div>
          </div>
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
          onOpenStore={onOpenStore}
        />
      )}
      {plantTypesLoading && (
        <div className="text-center text-muted-foreground">Loading plant types...</div>
      )}
    </div>
  );
}