import { toast } from 'sonner';
import { Loader, Plus, Users } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Button } from './ui/button';
import { apiService } from '../services/api';
import { Garden, SharedGardenResponse } from '../types/api';
import { CreateGardenModal } from './CreateGardenModal';
import { GardenCard } from './GardenCard';
import JoinGardenModal from './JoinGardenModal';
import ShareGardenModal from './ShareGardenModal';

interface GardenListProps {
  onSelectGarden: (garden: Garden) => void;
}

export function GardenList({ onSelectGarden }: GardenListProps) {
  const { user } = useAuth();
  const [loadError, setLoadError] = useState(false);
  const [gardenData, setGardenData] = useState<SharedGardenResponse>({ owned_gardens: [], shared_gardens: [] });
  const [loading, setLoading] = useState(true);
  const [shareGarden, setShareGarden] = useState<Garden | null>(null);

  const loadGardens = useCallback(() =>
    apiService.getGardens()
      .then(data => { setGardenData(data); setLoadError(false); })
      .catch(() => setLoadError(true))
      .finally(() => setLoading(false)),
  []);

  useEffect(() => {
    loadGardens();
  }, [loadGardens]);

  const handleCreateGarden = async (data: { name: string; description?: string }) => {
    try {
      const response = await apiService.createGarden(data);
      setGardenData(prev => ({
        ...prev,
        owned_gardens: [...(prev.owned_gardens ?? []), response.garden]
      }));
      if (user?.experience === 0 && !gardenData.owned_gardens?.length) onSelectGarden(response.garden);
    } catch (error) {
      console.error('Failed to create garden:', error);
      toast.error(`Failed to create garden: ${(error as Error).message}`);
    }
  };

  const handleGardenJoined = () => {
    loadGardens();
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Loader className="w-8 h-8 text-green-600 animate-spin" />
      </div>
    );
  }

  if (loadError) return <div role="alert"><p>Unable to load your gardens. Check your connection.</p><Button onClick={loadGardens}>Retry</Button></div>;

  return (
    <div className="garden-overview">
      {user && <section className="quest-panel" aria-label="Your adventure">
        <div className="flex flex-wrap justify-between gap-2"><h2>Your growing adventure</h2><span className="font-bold">Level {user.level}</span></div>
        <p>{user.next_level_experience - user.experience} XP to your next level</p>
        <progress aria-label="Level progress" className="w-full mt-3" value={user.experience - user.level_experience} max={user.next_level_experience - user.level_experience} />
        {user.experience === 0 && <>
          <h3 className="mt-4">First quest: plant a seed</h3>
          <ol><li>1. Create a garden</li><li>2. Buy seeds</li><li>3. Plant your first seed</li></ol>
          <p>Start with New Garden. Your first seed begins earning experience!</p>
        </>}
      </section>}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 sm:mb-6 space-y-3 sm:space-y-0">
        <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground">Your gardens</h2>
        <div className="flex items-center space-x-2">
          <JoinGardenModal onGardenJoined={handleGardenJoined} />
          <CreateGardenModal
            onClose={() => { }}
            onSubmit={handleCreateGarden}
          />
        </div>
      </div>

      {(gardenData.owned_gardens?.length || 0) === 0 && (gardenData.shared_gardens?.length || 0) === 0 ? (
        <div className="game-panel text-center py-8 sm:py-12">
          <div className="w-24 h-24 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
            <Plus className="w-12 h-12 text-green-600" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-foreground mb-2">No gardens yet</h3>
          <p className="text-sm sm:text-base text-muted-foreground mb-4 px-4">Create your first garden to start growing plants!</p>
          <CreateGardenModal
            onClose={() => { }}
            onSubmit={handleCreateGarden}
          />
        </div>
      ) : (
        <div className="space-y-8">
          {(gardenData.owned_gardens?.length || 0) > 0 && (
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-4">My Gardens</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-6">
                {(gardenData.owned_gardens || []).map((garden: Garden) => (
                  <GardenCard
                    key={garden.id}
                    garden={garden}
                    onClick={() => onSelectGarden(garden)}
                    onShare={() => setShareGarden(garden)}
                  />
                ))}
              </div>
            </div>
          )}

          {(gardenData.shared_gardens?.length || 0) > 0 && (
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-4 flex items-center">
                <Users className="w-5 h-5 mr-2" />
                Shared Gardens
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-6">
                {(gardenData.shared_gardens || []).map((sharedGarden) => (
                  <GardenCard
                    key={sharedGarden.garden.id}
                    garden={sharedGarden.garden}
                    permissions={sharedGarden.permission}
                    isShared={true}
                    sharedBy={sharedGarden.shared_by.username}
                    onClick={() => onSelectGarden(sharedGarden.garden)}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {shareGarden && (
        <ShareGardenModal
          garden={shareGarden}
          open={true}
          onOpenChange={(open) => {
            if (!open) {
              setShareGarden(null);
            }
          }}
        />
      )}
    </div>
  );
}