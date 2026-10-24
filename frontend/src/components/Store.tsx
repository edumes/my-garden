import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/contexts/AuthContext';
import { formatGrowthTime } from '@/lib/utils';
import { apiService } from '@/services/api';
import { PlantType, SeedInventory } from '@/types/api';
import { Coins, LockKeyhole, Minus, Package, Plus, ShoppingCart } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { toast } from 'sonner';

interface StoreProps {
  onClose: () => void;
  open: boolean;
}

export function Store({ onClose, open }: StoreProps) {
  const { user, refreshUser } = useAuth();
  const navigate = useNavigate();
  const [catalog, setCatalog] = useState<PlantType[]>([]);
  const [seeds, setSeeds] = useState<SeedInventory[]>([]);
  const [view, setView] = useState<'shop' | 'inventory'>('shop');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [errors, setErrors] = useState({ shop: false, inventory: false });
  const [retry, setRetry] = useState(0);
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [purchasing, setPurchasing] = useState<string | null>(null);
  const [receipt, setReceipt] = useState('');
  const [wasOpen, setWasOpen] = useState(open);
  if (wasOpen !== open) {
    setWasOpen(open);
    if (open) { setLoading(true); setSearch(''); setReceipt(''); }
  }

  useEffect(() => {
    if (!open) return;
    let active = true;
    Promise.allSettled([apiService.getStoreInventory(), apiService.getUserSeedInventory()])
      .then(([shop, inventory]) => {
        if (!active) return;
        if (shop.status === 'fulfilled') setCatalog(shop.value.inventory ?? []);
        if (inventory.status === 'fulfilled') setSeeds(inventory.value.seed_inventory ?? []);
        setErrors({ shop: shop.status === 'rejected', inventory: inventory.status === 'rejected' });
        setLoading(false);
      });
    return () => { active = false; };
  }, [open, retry]);

  const owned = (id: string) => seeds.find(seed => seed.plant_type_id === id)?.quantity ?? 0;
  const seedCount = seeds.reduce((total, seed) => total + seed.quantity, 0);
  const available = view === 'shop' ? catalog : seeds.filter(seed => seed.quantity > 0).map(seed => seed.plant_type);
  const displayed = available.filter(plant => plant.name.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => a.unlock_level - b.unlock_level || a.seed_price - b.seed_price || a.name.localeCompare(b.name));

  const handlePurchase = async (plant: PlantType) => {
    if (!user || purchasing || user.level < plant.unlock_level) return;
    const quantity = quantities[plant.id] ?? 1;
    if (user.coins < plant.seed_price * quantity) return;
    setPurchasing(plant.id);
    try {
      const purchase = await apiService.buySeed({ plant_type_id: plant.id, quantity });
      const message = `Successfully purchased ${purchase.quantity} ${plant.name} seed${purchase.quantity > 1 ? 's' : ''} for ${purchase.total_cost} coins!`;
      setReceipt(message);
      setQuantities(previous => ({ ...previous, [plant.id]: 1 }));
      await Promise.all([
        refreshUser(),
        apiService.getUserSeedInventory()
          .then(response => { setSeeds(response.seed_inventory ?? []); setErrors(previous => ({ ...previous, inventory: false })); })
          .catch(() => setErrors(previous => ({ ...previous, inventory: true }))),
      ]);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Unable to buy seeds. Please try again.');
    } finally {
      setPurchasing(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={isOpen => { if (!isOpen) onClose(); }}>
      <DialogContent className="garden-page seed-shop p-0 gap-0">
        <DialogHeader className="seed-shop-heading">
          <div className="flex items-center gap-2"><ShoppingCart aria-hidden="true" /><DialogTitle className="text-2xl font-extrabold">Garden Store</DialogTitle></div>
          <DialogDescription>A pocketful of seeds. A garden full of possibilities.</DialogDescription>
          <div className="seed-shop-balance"><Coins size={18} aria-hidden="true" /><strong>{user?.coins ?? 0} coins</strong><span>Level {user?.level ?? 1}</span></div>
        </DialogHeader>
        <div className="seed-shop-tools">
          <div className="flex gap-2" role="group" aria-label="Store views">
            <Button variant={view === 'shop' ? 'default' : 'outline'} aria-pressed={view === 'shop'} onClick={() => { setView('shop'); setSearch(''); }}>Shop seeds</Button>
            <Button variant={view === 'inventory' ? 'default' : 'outline'} aria-pressed={view === 'inventory'} onClick={() => { setView('inventory'); setSearch(''); }}><Package size={16} />My seeds{errors.inventory ? '' : ` (${seedCount})`}</Button>
          </div>
          <Input aria-label="Search seeds" placeholder="Find your next plant..." value={search} onChange={event => setSearch(event.target.value)} />
        </div>
        <div className="seed-shop-body" aria-busy={loading}>
          {receipt && <p role="status" className="seed-receipt">{receipt}</p>}
          {loading ? <p role="status" className="text-center p-8">Loading your seed shop...</p> : <>
            {(errors[view] || (view === 'shop' && errors.inventory)) && <div role="alert" className="seed-shop-error"><p>{errors[view] ? 'Unable to load seeds. Check your connection.' : 'Seed counts are unavailable. You can still browse or buy seeds.'}</p><Button variant="outline" onClick={() => { setLoading(true); setRetry(value => value + 1); }}>Retry loading seeds</Button></div>}
            {!errors[view] && <>
              {view === 'inventory' && <div className="seed-inventory-note"><p>Choose a garden and an empty plot to plant your seeds.</p><Button variant="outline" onClick={() => { onClose(); navigate('/garden'); }}>Go to gardens</Button></div>}
              {displayed.length === 0 && <div className="text-center p-8"><Package className="mx-auto mb-3" /><p>{search ? 'No seeds match your search.' : view === 'inventory' ? "You don't have any seeds yet." : 'No seeds are available right now.'}</p>{view === 'inventory' && !search && <Button className="mt-4" onClick={() => setView('shop')}>Browse seeds</Button>}</div>}
              <div className="seed-catalog">
                {displayed.map(plant => {
                  const quantity = quantities[plant.id] ?? 1;
                  const total = plant.seed_price * quantity;
                  const locked = (user?.level ?? 1) < plant.unlock_level;
                  const insufficient = (user?.coins ?? 0) < total;
                  return <article className="seed-card" key={plant.id} aria-label={`${plant.name} seeds`}>
                    <div className="seed-card-top"><span className="seed-icon" aria-hidden="true">{plant.icon}</span><span className="seed-owned">{errors.inventory ? 'Count unavailable' : `${owned(plant.id)} in bag`}</span></div>
                    <h3>{plant.name}</h3><p className="text-sm text-muted-foreground">{plant.description}</p>
                    <dl className="seed-facts"><div><dt>Base growth</dt><dd>{formatGrowthTime(plant.growth_time)}</dd></div><div><dt>Harvest payout</dt><dd>{plant.harvest_value * plant.yield} coins</dd></div></dl>
                    {view === 'shop' ? <div className="seed-purchase">
                      <p className="text-xs text-muted-foreground">{plant.seed_price} coins per seed</p>
                      <div className="flex items-center justify-between gap-2"><span className="text-sm">Quantity</span><div className="flex items-center gap-2">
                        <Button variant="outline" size="icon" aria-label={`Fewer ${plant.name} seeds`} disabled={quantity <= 1 || !!purchasing || locked} onClick={() => setQuantities(previous => ({ ...previous, [plant.id]: quantity - 1 }))}><Minus /></Button>
                        <span className="font-bold w-5 text-center">{quantity}</span>
                        <Button variant="outline" size="icon" aria-label={`More ${plant.name} seeds`} disabled={quantity >= 10 || !!purchasing || locked} onClick={() => setQuantities(previous => ({ ...previous, [plant.id]: quantity + 1 }))}><Plus /></Button>
                      </div></div>
                      <div className="flex justify-between text-sm"><span>Total</span><strong>{total} coins</strong></div>
                      <Button className="w-full" disabled={locked || insufficient || !!purchasing} onClick={() => handlePurchase(plant)}>{purchasing === plant.id ? 'Purchasing...' : locked ? <><LockKeyhole size={16} />Level {plant.unlock_level}</> : 'Buy Seeds'}</Button>
                      <p className="text-xs text-muted-foreground">{locked ? `Unlocks at level ${plant.unlock_level}` : insufficient ? 'Not enough coins for this quantity.' : `After purchase: ${(user?.coins ?? 0) - total} coins`}</p>
                    </div> : <div className="seed-inventory-count"><strong>{owned(plant.id)}</strong><span>ready to plant</span></div>}
                  </article>;
                })}
              </div>
            </>}
          </>}
        </div>
      </DialogContent>
    </Dialog>
  );
}
