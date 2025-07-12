import { Droplet, Plus, Scissors, Trash2, Zap } from 'lucide-react';
import { useState } from 'react';
import { Plant } from '../types/api';

interface PlantCellProps {
  position: number;
  plant?: Plant;
  onPlantAction: (plantId: string, action: 'water' | 'fertilize' | 'harvest' | 'remove') => void;
  onPlantSeed: (position: number) => void;
}

const stageIcons = {
  seed: '🌱',
  sprout: '🌿',
  growing: '🌱',
  mature: '🌿',
  harvestable: '🌾',
  withered: '🥀',
};

const stageColors = {
  seed: 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-600',
  sprout: 'bg-green-100 dark:bg-green-900/20 border-green-300 dark:border-green-600',
  growing: 'bg-green-200 dark:bg-green-800/20 border-green-400 dark:border-green-500',
  mature: 'bg-green-300 dark:bg-green-700/20 border-green-500 dark:border-green-400',
  harvestable: 'bg-yellow-200 dark:bg-yellow-800/20 border-yellow-400 dark:border-yellow-500',
  withered: 'bg-gray-200 dark:bg-gray-700/20 border-gray-400 dark:border-gray-500',
};

export function PlantCell({ position, plant, onPlantAction, onPlantSeed }: PlantCellProps) {
  const [showActions, setShowActions] = useState(false);

  if (!plant) {
    return (
      <div
        onClick={() => onPlantSeed(position)}
        className="aspect-square bg-brown-100 dark:bg-brown-900/20 border-2 border-brown-300 dark:border-brown-600 rounded-lg flex items-center justify-center cursor-pointer hover:bg-brown-200 dark:hover:bg-brown-800/20 transition-colors group min-h-[60px] sm:min-h-[80px]"
      >
        <Plus className="w-6 h-6 sm:w-8 sm:h-8 text-brown-500 dark:text-brown-400 group-hover:text-brown-600 dark:group-hover:text-brown-300" />
      </div>
    );
  }

  const canWater = plant.water_level < 100;
  const canFertilize = plant.last_fertilized_at === null ||
    (plant.last_fertilized_at && new Date().getTime() - new Date(plant.last_fertilized_at).getTime() > 24 * 60 * 60 * 1000);
  const canHarvest = plant.stage === 'harvestable';

  return (
    <div className="relative">
      <div
        className={`aspect-square ${stageColors[plant.stage]} border-2 rounded-lg flex flex-col items-center justify-center cursor-pointer transition-all hover:shadow-md min-h-[60px] sm:min-h-[80px]`}
        onClick={() => setShowActions(!showActions)}
      >
        <div className="text-lg sm:text-2xl mb-1">{stageIcons[plant.stage]}</div>
        <div className="text-xs text-center px-1">
          <div className="font-medium truncate text-foreground">{plant.plant_type.name}</div>
          <div className="text-muted-foreground">{Math.round(plant.growth_progress)}%</div>
        </div>

        {/* Health and water indicators */}
        <div className="absolute top-1 sm:top-2 right-1 sm:right-2 flex flex-col space-y-1">
          <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-red-500" style={{ opacity: plant.health / 100 }} />
          <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-blue-500" style={{ opacity: plant.water_level / 100 }} />
        </div>
      </div>

      {showActions && (
        <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 bg-card rounded-lg shadow-lg border border-border p-1 sm:p-2 z-10">
          <div className="flex space-x-1 sm:space-x-2">
            {canWater && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onPlantAction(plant.id, 'water');
                  setShowActions(false);
                }}
                className="p-1.5 sm:p-2 bg-blue-100 dark:bg-blue-900/20 hover:bg-blue-200 dark:hover:bg-blue-800/20 rounded-lg transition-colors"
                title="Water plant"
              >
                <Droplet className="w-3 h-3 sm:w-4 sm:h-4 text-blue-600" />
              </button>
            )}
            {canFertilize && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onPlantAction(plant.id, 'fertilize');
                  setShowActions(false);
                }}
                className="p-1.5 sm:p-2 bg-yellow-100 dark:bg-yellow-900/20 hover:bg-yellow-200 dark:hover:bg-yellow-800/20 rounded-lg transition-colors"
                title="Fertilize plant"
              >
                <Zap className="w-3 h-3 sm:w-4 sm:h-4 text-yellow-600" />
              </button>
            )}
            {canHarvest && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onPlantAction(plant.id, 'harvest');
                  setShowActions(false);
                }}
                className="p-1.5 sm:p-2 bg-green-100 dark:bg-green-900/20 hover:bg-green-200 dark:hover:bg-green-800/20 rounded-lg transition-colors"
                title="Harvest plant"
              >
                <Scissors className="w-3 h-3 sm:w-4 sm:h-4 text-green-600" />
              </button>
            )}
            <button
              onClick={(e) => {
                e.stopPropagation();
                onPlantAction(plant.id, 'remove');
                setShowActions(false);
              }}
              className="p-1.5 sm:p-2 bg-red-100 dark:bg-red-900/20 hover:bg-red-200 dark:hover:bg-red-800/20 rounded-lg transition-colors"
              title="Remove plant"
            >
              <Trash2 className="w-3 h-3 sm:w-4 sm:h-4 text-red-600" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}