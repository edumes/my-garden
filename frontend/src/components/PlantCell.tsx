import { Clock, Leaf, Plus, Scissors, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { formatGrowthTime } from '../lib/utils';
import { Plant } from '../types/api';
import { Tooltip, TooltipContent, TooltipTrigger } from './animate-ui/components/tooltip';

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

const stageNames = {
  seed: 'Seed',
  sprout: 'Sprout',
  growing: 'Growing',
  mature: 'Mature',
  harvestable: 'Ready to harvest',
  withered: 'Withered',
};

const stageColors = {
  seed: 'bg-accent border-accent-foreground/20',
  sprout: 'bg-green-100 dark:bg-green-900/20 border-green-300 dark:border-green-600',
  growing: 'bg-green-200 dark:bg-green-800/20 border-green-400 dark:border-green-500',
  mature: 'bg-green-300 dark:bg-green-700/20 border-green-500 dark:border-green-400',
  harvestable: 'dark:bg-yellow-900 border-yellow-400 dark:border-yellow-500',
  withered: 'bg-muted border-muted-foreground/30',
};

export function PlantCell({ position, plant, onPlantAction, onPlantSeed }: PlantCellProps) {
  const [showActions, setShowActions] = useState(false);
  const [timeoutId, setTimeoutId] = useState<NodeJS.Timeout | null>(null);

  if (!plant) {
    return (
      <div
        onClick={() => onPlantSeed(position)}
        className="aspect-square bg-secondary border-2 border-border rounded-lg flex items-center justify-center cursor-pointer hover:bg-accent transition-colors group min-h-[60px] sm:min-h-[80px]"
      >
        <Plus className="w-6 h-6 sm:w-8 sm:h-8 text-muted-foreground group-hover:text-foreground" />
      </div>
    );
  }

  const canHarvest = plant.stage === 'harvestable';

  const handleMouseEnter = () => {
    if (timeoutId) {
      clearTimeout(timeoutId);
      setTimeoutId(null);
    }
    setShowActions(true);
  };

  const handleMouseLeave = () => {
    const id = setTimeout(() => {
      setShowActions(false);
    }, 200);
    // setTimeoutId(id);
  };

  const getTimeSincePlanted = () => {
    const plantedAt = new Date(plant.planted_at);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - plantedAt.getTime()) / (1000 * 60));

    if (diffInMinutes < 60) {
      return `${diffInMinutes} min`;
    } else if (diffInMinutes < 1440) {
      const hours = Math.floor(diffInMinutes / 60);
      return `${hours}h`;
    } else {
      const days = Math.floor(diffInMinutes / 1440);
      return `${days}d`;
    }
  };

    return (
    <div className="relative">
      <Tooltip side="right" sideOffset={8}>
        <TooltipTrigger>
          <div
            className={`aspect-square ${stageColors[plant.stage]} border-2 rounded-lg flex flex-col items-center justify-center cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-105 min-h-[60px] sm:min-h-[80px]`}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
          >
            <div className="text-lg sm:text-2xl mb-1">{plant.plant_type.icon}</div>
            <div className="text-xs text-center px-1">
              <div className="font-medium truncate text-foreground">{plant.plant_type.name}</div>
              <div className="text-muted-foreground">{Math.round(plant.growth_progress)}%</div>
            </div>

            {/* Progress bar */}
            <div className="w-full max-w-[80%] mt-1">
              <div className="w-full bg-muted/30 dark:bg-muted/20 rounded-full h-1.5 overflow-hidden">
                <div
                  className="h-1.5 rounded-full transition-all duration-500 ease-out relative"
                  style={{
                    width: `${plant.growth_progress}%`,
                    background: plant.growth_progress >= 100
                      ? 'linear-gradient(90deg, #22c55e, #16a34a)'
                      : plant.growth_progress >= 75
                        ? 'linear-gradient(90deg, #eab308, #ca8a04)'
                        : plant.growth_progress >= 50
                          ? 'linear-gradient(90deg, #f59e42, #eab308)'
                          : 'linear-gradient(90deg, #fbbf24, #fde68a)'
                  }}
                >
                  {/* Shimmer effect for growing plants */}
                  {plant.growth_progress < 100 && (
                    <div className="absolute inset-0 bg-white/20 animate-pulse" />
                  )}
                </div>
              </div>
            </div>
          </div>
        </TooltipTrigger>
          <TooltipContent>
            <div className="space-y-2">
              {/* Plant Name and Stage */}
              <div className="text-center">
                <div className="text-sm font-semibold text-foreground">{plant.plant_type.name}</div>
                <div className="text-xs text-muted-foreground flex items-center justify-center space-x-1">
                  <span>{stageIcons[plant.stage]}</span>
                  <span>{stageNames[plant.stage]}</span>
                </div>
              </div>

              {/* Growth Progress */}
              <div className="space-y-1">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-muted-foreground">Growth</span>
                  <span className="font-medium">{Math.round(plant.growth_progress)}%</span>
                </div>
                <div className="w-full bg-muted/30 rounded-full h-1.5 overflow-hidden">
                  <div
                    className="h-1.5 rounded-full transition-all duration-300"
                    style={{
                      width: `${plant.growth_progress}%`,
                      background: plant.growth_progress >= 100
                        ? 'linear-gradient(90deg, #22c55e, #16a34a)'
                        : plant.growth_progress >= 75
                          ? 'linear-gradient(90deg, #eab308, #ca8a04)'
                          : plant.growth_progress >= 50
                            ? 'linear-gradient(90deg, #f59e42, #eab308)'
                            : 'linear-gradient(90deg, #fbbf24, #fde68a)'
                    }}
                  />
                </div>
              </div>

              {/* Plant Info */}
              <div className="space-y-1 text-xs">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1 text-muted-foreground">
                    <Clock className="w-3 h-3" />
                    <span>Planted</span>
                  </div>
                  <span className="font-medium">{getTimeSincePlanted()}</span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1 text-muted-foreground">
                    <Leaf className="w-3 h-3" />
                    <span>Yield</span>
                  </div>
                  <span className="font-medium">{plant.plant_type.yield}x</span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1 text-muted-foreground">
                    <Clock className="w-3 h-3" />
                    <span>Growth time</span>
                  </div>
                  <span className="font-medium">{formatGrowthTime(plant.plant_type.growth_time)}</span>
                </div>
              </div>

              {/* Harvest Value */}
              {canHarvest && (
                <div className="pt-1 border-t border-border">
                  <div className="flex items-center justify-center space-x-1 text-xs">
                    <span className="text-muted-foreground">Harvest value:</span>
                    <span className="font-semibold text-yellow-600">{plant.plant_type.harvest_value * plant.plant_type.yield} coins</span>
                  </div>
                </div>
              )}
            </div>
          </TooltipContent>
        </Tooltip>

      {/* Action Buttons */}
      <div
        className={`absolute top-full left-1/2 transform -translate-x-1/2 mt-2 bg-card rounded-lg shadow-lg border border-border p-1 sm:p-2 z-10 transition-all duration-200 ease-in-out ${showActions
          ? 'opacity-100 scale-100 translate-y-0'
          : 'opacity-0 scale-95 translate-y-1 pointer-events-none'
          }`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        <div className="flex space-x-1 sm:space-x-2">
          {canHarvest && (
            <Tooltip side="top" sideOffset={4}>
              <TooltipTrigger>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onPlantAction(plant.id, 'harvest');
                  }}
                  className="p-1.5 sm:p-2 bg-green-100 dark:bg-green-900/20 hover:bg-green-200 dark:hover:bg-green-800/20 rounded-lg transition-all duration-150 hover:scale-105"
                >
                  <Scissors className="w-3 h-3 sm:w-4 sm:h-4 text-green-600" />
                </button>
              </TooltipTrigger>
              <TooltipContent>
                Harvest plant
              </TooltipContent>
            </Tooltip>
          )}
          <Tooltip side="top" sideOffset={4}>
            <TooltipTrigger>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onPlantAction(plant.id, 'remove');
                }}
                className="p-1.5 sm:p-2 bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-all duration-150 hover:scale-105"
              >
                <Trash2 className="w-3 h-3 sm:w-4 sm:h-4 text-destructive" />
              </button>
            </TooltipTrigger>
            <TooltipContent>
              Remove plant
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </div>
  );
}