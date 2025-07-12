import { Loader, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Garden } from '../types/api';
import { CreateGardenModal } from './CreateGardenModal';
import { GardenCard } from './GardenCard';

interface GardenListProps {
  onSelectGarden: (garden: Garden) => void;
}

export function GardenList({ onSelectGarden }: GardenListProps) {
  const [gardens, setGardens] = useState<Garden[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadGardens();
  }, []);

  const loadGardens = async () => {
    try {
      const response = await apiService.getGardens();
      setGardens(response.gardens || []);
    } catch (error) {
      console.error('Failed to load gardens:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateGarden = async (data: { name: string; description?: string }) => {
    try {
      const response = await apiService.createGarden(data);
      setGardens([...gardens, response.garden]);
    } catch (error) {
      console.error('Failed to create garden:', error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader className="w-8 h-8 text-green-600 animate-spin" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 sm:mb-6 space-y-3 sm:space-y-0">
        <h2 className="text-xl sm:text-2xl font-bold text-foreground">My Gardens</h2>
        <CreateGardenModal
          onClose={() => {}}
          onSubmit={handleCreateGarden}
        />
      </div>

      {gardens.length === 0 ? (
        <div className="text-center py-8 sm:py-12">
          <div className="w-24 h-24 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
            <Plus className="w-12 h-12 text-green-600" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-foreground mb-2">No gardens yet</h3>
          <p className="text-sm sm:text-base text-muted-foreground mb-4 px-4">Create your first garden to start growing plants!</p>
          <CreateGardenModal
            onClose={() => {}}
            onSubmit={handleCreateGarden}
          />
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
          {gardens.map((garden) => (
            <GardenCard
              key={garden.id}
              garden={garden}
              onClick={() => onSelectGarden(garden)}
              onEdit={() => {/* TODO: Implement edit functionality */ }}
            />
          ))}
        </div>
      )}
    </div>
  );
}