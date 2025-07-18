import { Calendar, Droplet, Settings, TreePine, Zap } from 'lucide-react';
import { Garden } from '../types/api';
import { Card } from './ui/card';
import { Button } from './ui/button';

interface GardenCardProps {
  garden: Garden;
  onClick: () => void;
  onEdit: () => void;
}

export function GardenCard({ garden, onClick, onEdit }: GardenCardProps) {
  const plantCount = garden.plants?.length || 0;
  const maxPlants = garden.size;

  return (
    <Card className="rounded-xl shadow-sm border border-border hover:shadow-md transition-shadow cursor-pointer">
      <div onClick={onClick} className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-foreground">{garden.name}</h3>
          <Button
            onClick={(e) => {
              e.stopPropagation();
              onEdit();
            }}
            variant="ghost"
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <Settings className="w-4 h-4" />
          </Button>
        </div>

        {garden.description && (
          <p className="text-muted-foreground text-sm mb-4">{garden.description}</p>
        )}

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div className="flex items-center space-x-2">
            <TreePine className="w-4 h-4 text-green-500" />
            <span className="text-sm text-muted-foreground">{plantCount}/{maxPlants} plants</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {new Date(garden.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>
    </Card>
  );
}