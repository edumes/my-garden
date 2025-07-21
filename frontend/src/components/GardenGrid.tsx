import { Plant } from '../types/api';
// import { MotionEffect } from './animate-ui/effects/motion-effect';
import { PlantCell } from './PlantCell';

interface GardenGridProps {
  plants: Plant[];
  onPlantAction: (plantId: string, action: 'water' | 'fertilize' | 'harvest' | 'remove') => void;
  onPlantSeed: (position: number) => void;
}

export function GardenGrid({ plants, onPlantAction, onPlantSeed }: GardenGridProps) {
  const gridPositions = Array.from({ length: 9 }, (_, i) => i);

  return (
    <div className="bg-green-50 rounded-xl p-3 sm:p-6 border-2 border-green-200">
      <div className="grid grid-cols-3 gap-2 sm:gap-4 max-w-xs sm:max-w-md mx-auto">
        {gridPositions.map((position, index) => {
          const plant = plants.find(p => p.position === position);
          return (
            // <MotionEffect
            //   key={index}
            //   slide={{
            //     direction: 'up',
            //   }}
            //   fade
            //   blur
            //   zoom
            //   inView
            //   delay={0.2 + index * 0.1}
            // >
              <PlantCell
                key={position}
                position={position}
                plant={plant}
                onPlantAction={onPlantAction}
                onPlantSeed={onPlantSeed}
              />
            // </MotionEffect>
          );
        })}
      </div>
    </div>
  );
}