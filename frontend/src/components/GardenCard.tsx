import { Calendar, Droplet, Settings, TreePine, Zap } from 'lucide-react';
import { Garden } from '../types/api';
import { Card } from './ui/card';

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
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEdit();
            }}
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <Settings className="w-4 h-4" />
          </button>
        </div>

        {garden.description && (
          <p className="text-muted-foreground text-sm mb-4">{garden.description}</p>
        )}

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div className="flex items-center space-x-2">
            <TreePine className="w-4 h-4 text-green-500" />
            <span className="text-sm text-muted-foreground">{plantCount}/{maxPlants} plants</span>
          </div>
          <div className="flex items-center space-x-2">
            <Droplet className="w-4 h-4 text-blue-500" />
            <span className="text-sm text-muted-foreground">{garden.water_level}% water</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-center space-x-2">
            <Zap className="w-4 h-4 text-yellow-500" />
            <span className="text-sm text-muted-foreground">{garden.fertilizer_level}% fertilizer</span>
          </div>
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {new Date(garden.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>

        {/* Upgrades indicators */}
        {(garden.has_sprinkler || garden.has_greenhouse || garden.has_composter) && (
          <div className="mt-4 pt-4 border-t border-border">
            <div className="flex space-x-2">
              {garden.has_sprinkler && (
                <span className="text-xs bg-blue-100 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200 px-2 py-1 rounded-full">
                  Sprinkler
                </span>
              )}
              {garden.has_greenhouse && (
                <span className="text-xs bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200 px-2 py-1 rounded-full">
                  Greenhouse
                </span>
              )}
              {garden.has_composter && (
                <span className="text-xs bg-brown-100 dark:bg-brown-900/20 text-brown-800 dark:text-brown-200 px-2 py-1 rounded-full">
                  Composter
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </Card>
  );
}