import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { useAuth } from '@/contexts/AuthContext';
import { formatGrowthTime } from '@/lib/utils';
import { apiService } from '@/services/api';
import { BuySeedRequest, PlantType } from '@/types/api';
import { Coins, Minus, Package, Plus, ShoppingCart } from 'lucide-react';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

interface StoreProps {
  onClose: () => void;
}

export function Store({ onClose }: StoreProps) {
  const [inventory, setInventory] = useState<PlantType[]>([]);
  const [loading, setLoading] = useState(true);
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [purchasing, setPurchasing] = useState<string | null>(null);
  const { user, refreshUser } = useAuth();

  useEffect(() => {
    loadInventory();
  }, []);

  const loadInventory = async () => {
    try {
      setLoading(true);
      const response = await apiService.getStoreInventory();
      setInventory(response.inventory);

      const initialQuantities: Record<string, number> = {};
      response.inventory.forEach((item) => {
        initialQuantities[item.id] = 1;
      });
      setQuantities(initialQuantities);
    } catch (err) {
      toast.error('Failed to load store inventory');
      console.error('Error loading inventory:', err);
    } finally {
      setLoading(false);
    }
  };

  const updateQuantity = (plantTypeId: string, change: number) => {
    const currentQuantity = quantities[plantTypeId] || 1;
    const newQuantity = Math.max(1, Math.min(10, currentQuantity + change));
    setQuantities(prev => ({
      ...prev,
      [plantTypeId]: newQuantity
    }));
  };

  const getTotalCost = (plantType: PlantType) => {
    const quantity = quantities[plantType.id] || 1;
    return plantType.seed_price * quantity;
  };

  const canAfford = (plantType: PlantType) => {
    if (!user) return false;
    return user.coins >= getTotalCost(plantType);
  };

  const handlePurchase = async (plantType: PlantType) => {
    if (!user) return;

    const quantity = quantities[plantType.id] || 1;
    const totalCost = getTotalCost(plantType);

    if (user.coins < totalCost) {
      toast.error('Insufficient coins');
      return;
    }

    try {
      setPurchasing(plantType.id);

      const request: BuySeedRequest = {
        plant_type_id: plantType.id,
        quantity: quantity
      };

      const response = await apiService.buySeed(request);

      await refreshUser();

      setQuantities(prev => ({
        ...prev,
        [plantType.id]: 1
      }));

      const successMsg = `Successfully purchased ${quantity} ${plantType.name} seed${quantity > 1 ? 's' : ''} for ${totalCost} coins!`;
      toast.success(successMsg);

    } catch (err: any) {
      toast.error(err.message || 'Failed to purchase seeds');
      console.error('Error purchasing seeds:', err);
    } finally {
      setPurchasing(null);
    }
  };

  if (loading) {
    return (
      <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
        <div className="text-center">
          <Package className="w-8 h-8 mx-auto mb-4 animate-pulse" />
          <p className="text-muted-foreground">Loading store...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-background border rounded-lg shadow-lg max-w-4xl w-full max-h-[90vh] overflow-hidden">
        <div className="flex items-center justify-between p-6 border-b">
          <div className="flex items-center space-x-2">
            <ShoppingCart className="w-6 h-6 text-primary" />
            <h2 className="text-2xl font-bold">Garden Store</h2>
          </div>
          <div className="flex items-center space-x-2">
            <Coins className="w-5 h-5 text-yellow-500" />
            <span className="font-semibold">{user?.coins || 0} coins</span>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            ✕
          </Button>
        </div>

        <div className="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {inventory.map((plantType) => (
              <Card key={plantType.id} className="relative">
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center space-x-2">
                      <span className="text-2xl">{plantType.icon}</span>
                      <div>
                        <CardTitle className="text-lg">{plantType.name}</CardTitle>
                        <CardDescription className="text-sm">
                          {plantType.description}
                        </CardDescription>
                      </div>
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                    <div>Growth: {formatGrowthTime(plantType.growth_time)}</div>
                    <div>Yield: {plantType.yield}</div>
                    <div>Value: {plantType.harvest_value} coins</div>
                    <div className="font-semibold text-primary">
                      Price: {plantType.seed_price} coins
                    </div>
                  </div>

                  <Separator />

                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Quantity:</span>
                      <div className="flex items-center space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => updateQuantity(plantType.id, -1)}
                          disabled={quantities[plantType.id] <= 1}
                        >
                          <Minus className="w-3 h-3" />
                        </Button>
                        <span className="w-8 text-center font-medium">
                          {quantities[plantType.id] || 1}
                        </span>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => updateQuantity(plantType.id, 1)}
                          disabled={quantities[plantType.id] >= 10}
                        >
                          <Plus className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>

                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Total:</span>
                      <span className="font-semibold text-primary">
                        {getTotalCost(plantType)} coins
                      </span>
                    </div>

                    <Button
                      className="w-full"
                      onClick={() => handlePurchase(plantType)}
                      disabled={!canAfford(plantType) || purchasing === plantType.id}
                    >
                      {purchasing === plantType.id ? (
                        'Purchasing...'
                      ) : (
                        <>
                          <ShoppingCart className="w-4 h-4 mr-2" />
                          Buy Seeds
                        </>
                      )}
                    </Button>

                    {!canAfford(plantType) && (
                      <p className="text-xs text-destructive text-center">
                        Not enough coins
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
} 