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
        className="garden-plot-empty"
      >
        <Plus aria-hidden="true" className="w-6 h-6" />
        <span className="text-xs">{can('plant') ? 'Plant seed' : 'Empty plot'}</span>
      </button>
    );
  }

  const ripe = plant.stage === 'harvestable';
  return (
    <div aria-label={`Position ${position + 1}: ${plant.plant_type.name}`} className={`garden-plant ${ripe ? 'is-ripe' : ''}`}>
      <div aria-hidden="true" className="garden-plant-icon">{plant.plant_type.icon}</div>
      <div className="text-sm font-medium">{plant.plant_type.name}</div>
      <p className="text-xs">{stageNames[plant.stage]}</p>
      <progress aria-label={`${plant.plant_type.name} growth`} max={100} value={plant.growth_progress} className="w-full" />
      <p className="text-xs font-semibold">{Math.round(plant.growth_progress)}%</p>
      <details className="garden-plant-info"><summary>Plant info</summary>
      <p className="text-xs text-muted-foreground">Harvest: {plant.plant_type.harvest_value * plant.plant_type.yield} coins</p>
      <p className="text-xs text-muted-foreground">Base growth: {formatGrowthTime(plant.plant_type.growth_time)}</p>
      </details>
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
