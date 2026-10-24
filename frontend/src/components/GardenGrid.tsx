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
    <div className="bg-green-50 dark:bg-green-950 rounded-xl p-3 sm:p-6 border-2 border-green-200">
      {plants.length === 0 && <p className="text-center text-sm mb-4">Your garden is empty. Buy seeds in the store, then choose a plot to plant.</p>}
      <div className="grid grid-cols-3 gap-2 sm:gap-4 mx-auto">
        {Array.from({ length: 9 }, (_, position) => (
          <PlantCell key={position} position={position} plant={plants.find(p => p.position === position)}
            permissions={permissions} busy={busy} onPlantAction={onPlantAction} onPlantSeed={onPlantSeed} />
        ))}
      </div>
    </div>
  );
}
