import { Plant } from '../types/api';
import { PlantCell } from './PlantCell';

interface GardenGridProps {
  plants: Plant[];
  onPlantAction: (plantId: string, action: 'harvest' | 'remove') => void;
  onPlantSeed: (position: number) => void;
}

export function GardenGrid({ plants, onPlantAction, onPlantSeed }: GardenGridProps) {
  const gridPositions = Array.from({ length: 9 }, (_, i) => i);

  return (
    <div className="bg-green-50 rounded-xl p-3 sm:p-6 border-2 border-green-200">
      <div className="grid grid-cols-3 gap-2 sm:gap-4 max-w-xs sm:max-w-md mx-auto">
        {gridPositions.map((position) => {
          const plant = plants.find(p => p.position === position);
          return (
            <PlantCell
              key={position}
              position={position}
              plant={plant}
              onPlantAction={onPlantAction}
              onPlantSeed={onPlantSeed}
            />
          );
        })}
      </div>
    </div>
  );
}