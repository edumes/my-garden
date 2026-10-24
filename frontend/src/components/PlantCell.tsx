import { Plus } from 'lucide-react';
import { formatGrowthTime } from '../lib/utils';
import { GardenPermission, Plant } from '../types/api';
import { Button } from './ui/button';

interface PlantCellProps {
  position: number;
  plant?: Plant;
  permissions: GardenPermission[];
  busy: boolean;
  onPlantAction: (plantId: string, action: 'harvest' | 'remove') => void;
  onPlantSeed: (position: number) => void;
}

const stageNames = {
  seed: 'Seed', sprout: 'Sprout', growing: 'Growing',
  mature: 'Mature', harvestable: 'Ready to harvest',
};

export function PlantCell({ position, plant, permissions, busy, onPlantAction, onPlantSeed }: PlantCellProps) {
  const can = (permission: GardenPermission) => permissions.includes(permission);
  if (!plant) {
    return (
      <button
        aria-label={`Plant seed in position ${position + 1}`}
        disabled={busy || !can('plant')}
        onClick={() => onPlantSeed(position)}
        className="aspect-square bg-secondary border-2 border-border rounded-lg flex flex-col items-center justify-center gap-2 hover:bg-accent disabled:opacity-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
      >
        <Plus aria-hidden="true" className="w-6 h-6" />
        <span className="text-xs">{can('plant') ? 'Plant seed' : 'Empty plot'}</span>
      </button>
    );
  }

  const ripe = plant.stage === 'harvestable';
  return (
    <div aria-label={`Position ${position + 1}: ${plant.plant_type.name}`} className={`border-2 rounded-lg p-2 space-y-2 text-center ${ripe ? 'border-yellow-500 bg-yellow-50 dark:bg-yellow-950' : 'border-green-300 bg-card'}`}>
      <div aria-hidden="true" className="text-2xl">{plant.plant_type.icon}</div>
      <div className="text-sm font-medium">{plant.plant_type.name}</div>
      <p className="text-xs">{stageNames[plant.stage]}</p>
      <progress aria-label={`${plant.plant_type.name} growth`} max={100} value={plant.growth_progress} className="w-full" />
      <p className="text-xs">{Math.round(plant.growth_progress)}% - {ripe ? 'Harvest when ready; ripe plants never wither.' : 'Growing automatically; wait for harvest.'}</p>
      <p className="text-xs text-muted-foreground">Harvest: {plant.plant_type.harvest_value * plant.plant_type.yield} coins</p>
      <p className="text-xs text-muted-foreground">Base growth: {formatGrowthTime(plant.plant_type.growth_time)}</p>
      {ripe && can('harvest') && (
        <Button size="sm" className="w-full whitespace-normal h-auto min-h-9 px-1 text-xs" disabled={busy} onClick={() => onPlantAction(plant.id, 'harvest')}>
          Harvest plant
        </Button>
      )}
      {can('manage') && (
        <Button size="sm" variant="outline" className="w-full whitespace-normal h-auto min-h-9 px-1 text-xs" disabled={busy} onClick={() => onPlantAction(plant.id, 'remove')}>
          Remove plant
        </Button>
      )}
    </div>
  );
}
