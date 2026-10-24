import { GardenPermission, Plant } from '../types/api';
import { PlantCell } from './PlantCell';

interface GardenGridProps {
  plants: Plant[];
  permissions: GardenPermission[];
  busy: boolean;
  onPlantAction: (plantId: string, action: 'harvest' | 'remove') => void;
  onPlantSeed: (position: number) => void;
}

export function GardenGrid({ plants, permissions, busy, onPlantAction, onPlantSeed }: GardenGridProps) {
  return (
    <section className="game-grid" aria-label="Garden plots">
      <div className="mb-4"><h2 className="font-extrabold text-xl">Your little patch</h2><p className="text-sm mt-1">{plants.some(plant => plant.stage === 'harvestable') ? 'Harvest time! Ripe plants will wait for you.' : 'Plants grow automatically. Check back for your next harvest.'}</p></div>
      {plants.length === 0 && <p className="text-center text-sm mb-4">{permissions.includes('plant') ? 'Your garden is empty. Buy seeds in the store, then choose a plot to plant.' : 'This garden is empty. You have viewing access.'}</p>}
      <div className="grid grid-cols-3 gap-2 sm:gap-4 mx-auto">
        {Array.from({ length: 9 }, (_, position) => (
          <PlantCell key={position} position={position} plant={plants.find(p => p.position === position)}
            permissions={permissions} busy={busy} onPlantAction={onPlantAction} onPlantSeed={onPlantSeed} />
        ))}
      </div>
    </section>
  );
}
