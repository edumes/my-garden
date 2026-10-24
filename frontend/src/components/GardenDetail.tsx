import { ArrowLeft, Coins, Leaf, Shield, User, Users } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';
import { Garden, GardenPermission, PlantType } from '../types/api';
import { GardenGrid } from './GardenGrid';
import { PlantSeedModal } from './PlantSeedModal';
import { Button } from './ui/button';
import { WeatherWidget } from './WeatherWidget';

interface GardenDetailProps {
  garden: Garden;
  isOwner: boolean;
  permissions: GardenPermission[];
  onBack: () => void;
  onOpenStore?: () => void;
}

export function GardenDetail({ garden: initialGarden, isOwner, permissions, onBack, onOpenStore }: GardenDetailProps) {
  const [capabilities, setCapabilities] = useState(permissions);
  const can = (permission: GardenPermission) => capabilities.includes(permission);
  const [syncError, setSyncError] = useState(false);
  const [busy, setBusy] = useState(false);
  const [garden, setGarden] = useState<Garden>(initialGarden);
  const [showPlantModal, setShowPlantModal] = useState(false);
  const [selectedPosition, setSelectedPosition] = useState<number | null>(null);
  const [plantTypes, setPlantTypes] = useState<PlantType[]>([]);
  const [plantTypesLoading, setPlantTypesLoading] = useState(true);
  const { user, refreshUser } = useAuth();
  const [hasSeeds, setHasSeeds] = useState<boolean | null>(null);

  const gardenId = initialGarden.id;

  useEffect(() => {
    if (user?.experience !== 0) return;
    let active = true;
    apiService.getUserSeedInventory()
      .then(response => { if (active) setHasSeeds(response.seed_inventory.some(seed => seed.quantity > 0)); })
      .catch(() => { if (active) setHasSeeds(null); });
    return () => { active = false; };
  }, [user?.experience, user?.coins]);

  const loadGardenDetails = useCallback(() =>
    apiService.getGarden(gardenId)
      .then((response) => {
        setGarden(response.garden);
        setCapabilities(response.permissions);
        setSyncError(false);
      })
      .catch(() => setSyncError(true)),
  [gardenId]);

  useEffect(() => {
    loadGardenDetails();
    const interval = setInterval(loadGardenDetails, 2000);
    return () => clearInterval(interval);
  }, [loadGardenDetails]);

  useEffect(() => {
    apiService.getPlantTypes()
      .then((response) => setPlantTypes(response.plant_types))
      .catch(() => toast.error('Unable to load plant types. Reload the page to retry.'))
      .finally(() => setPlantTypesLoading(false));
  }, []);

  const handlePlantAction = async (plantId: string, action: 'harvest' | 'remove') => {
    if (action === 'harvest' && !can('harvest')) {
      toast.error('You don\'t have permission to harvest in this garden');
      return;
    }
    if (action === 'remove' && !can('manage')) {
      toast.error('You don\'t have permission to remove plants from this garden');
      return;
    }

    if (busy) return;
    setBusy(true);
    try {
      const plant = garden.plants?.find(p => p.id === plantId);
      if (!plant) return;

      switch (action) {
        case 'harvest': {
          const { reward } = await apiService.harvestPlant(garden.id, plantId);
          await refreshUser();

          if (reward.user_id !== user?.id) {
            toast.success(`Harvested! ${reward.coins} coins and ${reward.experience} XP went to the player who grew it`);
          } else if (reward.level_up) {
            toast.success(`Harvested! +${reward.coins} coins, +${reward.experience} XP | Level up! You're now level ${reward.level}`);
          } else {
            toast.success(`Harvested! +${reward.coins} coins, +${reward.experience} XP`);
          }
          break;
        }
        case 'remove':
          await apiService.removePlant(garden.id, plantId);
          toast.success('Plant removed successfully');
          break;
      }

      await loadGardenDetails();
    } catch (error) {
      console.error(`Failed to ${action} plant:`, error);
      toast.error(error instanceof Error ? error.message : `Failed to ${action} plant`);
    } finally {
      setBusy(false);
    }
  };

  const handlePlantSeed = (position: number) => {
    if (!can('plant')) {
      toast.error('You don\'t have permission to plant in this garden');
      return;
    }
    setSelectedPosition(position);
    setShowPlantModal(true);
  };

  const handlePlantSubmit = async (plantTypeId: string) => {
    if (selectedPosition === null || busy) return;
    setBusy(true);

    try {
      const { reward } = await apiService.plantSeed(garden.id, {
        plant_type_id: plantTypeId,
        position: selectedPosition,
      });

      setShowPlantModal(false);
      setSelectedPosition(null);
      await Promise.all([loadGardenDetails(), refreshUser()]);
      toast.success(reward.level_up
        ? `Seed planted successfully! Level up! You're now level ${reward.level}`
        : 'Seed planted successfully!');
    } catch (error) {
      console.error('Failed to plant seed:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to plant seed');
    } finally {
      setBusy(false);
    }
  };

  const plants = garden.plants ?? [];
  const readyCount = plants.filter(plant => plant.stage === 'harvestable').length;
  const plantTypeCounts = plants.reduce((counts, plant) => {
    counts[plant.plant_type.name] = (counts[plant.plant_type.name] ?? 0) + 1;
    return counts;
  }, {} as Record<string, number>);
  let dateFormat: Intl.DateTimeFormat;
  try {
    dateFormat = new Intl.DateTimeFormat(user?.language || undefined, { timeZone: user?.timezone || undefined });
  } catch {
    // Older saved preferences may contain unsupported locale or timezone values.
    dateFormat = new Intl.DateTimeFormat();
  }

  return (
    <div className="garden-page max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8 space-y-5 sm:space-y-6">
      <header className="garden-heading">
        <Button aria-label="Back to gardens" onClick={onBack} variant="outline" size="icon"><ArrowLeft /></Button>
        <div className="min-w-0 flex-1">
          <p className="garden-eyebrow"><Leaf size={14} /> A little care, a little growth</p>
          <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight break-words">{garden.name}</h1>
          {garden.description && <p className="text-sm text-muted-foreground mt-2 break-words">{garden.description}</p>}
        </div>
        <span className="garden-access"><Shield size={14} />{isOwner ? 'Owner' : <><Users size={14} />Shared</>}</span>
      </header>
      <section className="garden-summary" aria-label="Garden summary">
        {[{ count: plants.length, label: 'Plants growing here' }, { count: readyCount, label: 'Ready to harvest' }, { count: 9 - plants.length, label: 'Plots available' }].map(item => (
          <div key={item.label}><strong>{item.count}</strong><span>{item.label}</span></div>
        ))}
      </section>
      {user?.experience === 0 && can('plant') && <section className="quest-panel" aria-label="First planting quest">
        <h2>First quest: plant a seed</h2>
        <ol><li>1. Garden created!</li><li>{hasSeeds ? '2. Seeds ready!' : '2. Buy your seeds'}</li><li>3. Choose an empty plot</li></ol>
        <p>{hasSeeds ? 'Your seeds are ready. Choose any empty plot below, select a seed and plant it to earn your first 2 XP.' : 'Open the store and buy a level 1 seed. Then choose an empty plot below to plant it.'}</p>
        {hasSeeds === false && onOpenStore && <Button className="mt-3" onClick={onOpenStore}>Buy first seeds</Button>}
      </section>}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6">
        <div className="lg:col-span-3 order-1 min-w-0">
          {syncError && <div role="alert" className="mb-4 p-3 border rounded-lg">
            Garden updates unavailable. Check your connection. Showing the last loaded state.
            <Button variant="outline" onClick={loadGardenDetails}>Retry</Button>
          </div>}
          <GardenGrid
            permissions={syncError ? [] : capabilities}
            busy={busy}
            plants={garden.plants || []}
            onPlantAction={(plantId, action) => handlePlantAction(plantId, action)}
            onPlantSeed={handlePlantSeed}
          />
        </div>

        <div className="space-y-4 sm:space-y-6 order-2 min-w-0">
          <WeatherWidget />
          <section className="game-panel p-4 sm:p-5" aria-label="Gardener stats">
            <h2 className="font-bold flex items-center gap-2"><User size={18} /> Gardener Stats</h2>
            <div className="mt-4 space-y-3 text-sm">
              <div className="flex justify-between"><span>Level</span><strong>{user?.level}</strong></div>
              <div className="flex justify-between"><span>Experience</span><strong>{user?.experience} XP</strong></div>
              <div className="flex justify-between"><span className="flex items-center gap-1"><Coins size={16} />Coins</span><strong>{user?.coins}</strong></div>
            </div>
            {user && <><progress aria-label="Level progress" className="w-full mt-4" value={user.experience - user.level_experience} max={user.next_level_experience - user.level_experience} /><p className="text-xs text-muted-foreground mt-1">{user.next_level_experience - user.experience} XP to level {user.level + 1}</p></>}
          </section>
          <section className="game-panel p-4 sm:p-5 text-sm space-y-3" aria-label="About this garden">
            <h2 className="font-bold">About this garden</h2>
            <p><strong>Access:</strong> {capabilities.join(', ')}</p>
            <p className="text-muted-foreground">Created {dateFormat.format(new Date(garden.created_at))}</p>
            <p className="text-muted-foreground">Updated {dateFormat.format(new Date(garden.updated_at))}</p>
            {plants.length > 0 && <div className="border-t pt-3 space-y-2"><h3 className="font-semibold">Plant Types</h3>{Object.entries(plantTypeCounts).map(([name, count]) => <p key={name} className="flex justify-between"><span>{name}</span><span>{count}</span></p>)}</div>}
          </section>
        </div>
      </div>

      {showPlantModal && (
        <PlantSeedModal
          busy={busy}
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
