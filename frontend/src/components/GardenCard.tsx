import { Leaf, Share2, Users } from 'lucide-react';
import { Garden, GardenPermission } from '../types/api';
import { Button } from './ui/button';

interface GardenCardProps {
  garden: Garden;
  permissions?: GardenPermission[];
  isShared?: boolean;
  sharedBy?: string;
  onClick: () => void;
  onShare?: () => void;
}

export function GardenCard({ garden, permissions, isShared, sharedBy, onClick, onShare }: GardenCardProps) {
  const plants = garden.plants ?? [];
  const ready = plants.filter(plant => plant.stage === 'harvestable').length;
  return (
    <article className="garden-card">
      <div className="garden-card-top">
        <span className="garden-eyebrow">{isShared ? <><Users size={14} /> Shared garden</> : <><Leaf size={14} /> Your garden</>}</span>
        {onShare && <Button size="icon" variant="ghost" aria-label={`Share garden ${garden.name}`} onClick={onShare}><Share2 /></Button>}
      </div>
      <button className="garden-card-open" aria-label={`Open garden ${garden.name}`} onClick={onClick}>
        <div className="garden-preview" aria-hidden="true">
          {Array.from({ length: 9 }, (_, position) => {
            const plant = plants.find(p => p.position === position);
            return <span key={position} className={plant?.stage === 'harvestable' ? 'is-ripe' : ''}>{plant ? plant.plant_type.icon : '+'}</span>;
          })}
        </div>
        <h3>{garden.name}</h3>
        {garden.description && <p className="text-sm text-muted-foreground line-clamp-2">{garden.description}</p>}
        <div className="garden-card-counts"><span>{plants.length}/9 plots planted</span><span className={ready ? 'garden-ready' : ''}>{ready} ready to harvest</span></div>
        <span className="garden-open-label">{ready ? 'Time to harvest' : 'Visit garden'} <span aria-hidden="true">&rarr;</span></span>
      </button>
      {isShared && <p className="garden-card-footer">{sharedBy && `Shared by ${sharedBy}. `}{permissions?.join(', ')}</p>}
    </article>
  );
}
