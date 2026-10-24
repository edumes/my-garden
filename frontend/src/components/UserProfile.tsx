import { Coins, Leaf, LockKeyhole, Sprout, Trophy, User } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';
import { PlantType } from '../types/api';
import { Button } from './ui/button';

export function UserProfile() {
  const { user: profile, refreshUser } = useAuth();
  const [catalog, setCatalog] = useState<PlantType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    let active = true;
    void refreshUser();
    apiService.getPlantTypes()
      .then(response => { if (active) { setCatalog(response.plant_types ?? []); setError(false); } })
      .catch(() => { if (active) setError(true); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [refreshUser, retry]);
  if (!profile) return null;

  const available = catalog.filter(plant => plant.unlock_level <= profile.level);
  const levels = [...new Set(catalog.map(plant => plant.unlock_level))].sort((a, b) => a - b);
  const nextSeedLevel = levels.find(level => level > profile.level);
  const earnedThisLevel = profile.experience - profile.level_experience;
  const levelCost = profile.next_level_experience - profile.level_experience;

  return (
    <main className="garden-page progress-page max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-6 space-y-6">
      <header className="garden-heading">
        <div className="progress-avatar" aria-hidden="true">{profile.username.charAt(0).toUpperCase()}</div>
        <div className="min-w-0 flex-1"><p className="garden-eyebrow"><Trophy size={14} /> Every seed is a step forward</p><h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight">Your growing story</h1><p className="text-muted-foreground mt-2 break-words">@{profile.username}</p></div>
        <Button asChild variant="outline"><Link to="/garden">Back to gardens</Link></Button>
      </header>
      <section className="quest-panel progress-level" aria-label="Level progress">
        <div className="flex justify-between items-center gap-3"><div><p className="garden-eyebrow">Current level</p><h2 className="text-4xl">Level {profile.level}</h2></div><Trophy className="w-12 h-12" aria-hidden="true" /></div>
        <progress aria-label={`Progress to level ${profile.level + 1}`} className="w-full mt-5" value={earnedThisLevel} max={levelCost} />
        <div className="flex flex-wrap justify-between gap-2 text-sm mt-2"><span>{earnedThisLevel} / {levelCost} XP this level</span><strong>{profile.next_level_experience - profile.experience} XP to level {profile.level + 1}</strong></div>
        <p className="text-sm mt-3">{profile.experience} total XP earned</p>
      </section>
      <div className="progress-milestones">
        <section className="game-panel p-5" aria-label="Garden unlocks"><Sprout aria-hidden="true" /><h2>Room to grow</h2><p className="progress-number">{profile.garden_limit ?? '-'}</p><p>gardens you can own</p><p className="text-sm text-muted-foreground mt-3">{profile.next_garden_level ? `Your next garden slot unlocks at level ${profile.next_garden_level}. Shared gardens use no slots.` : 'Loading garden unlocks...'}</p></section>
        <section className="game-panel p-5" aria-label="Seed unlocks"><Leaf aria-hidden="true" /><h2>Your seed collection</h2><p className="progress-number">{loading || error ? '-' : available.length}</p><p>seed types unlocked</p><p className="text-sm text-muted-foreground mt-3">{loading ? 'Loading the seed catalog...' : error ? 'Seed unlock information is unavailable.' : nextSeedLevel ? `More seeds unlock at level ${nextSeedLevel}.` : catalog.length ? 'You have unlocked every seed type in the catalog!' : 'No seed types are catalogued yet.'}</p></section>
        <section className="game-panel p-5" aria-label="Coin balance"><Coins aria-hidden="true" /><h2>Your seed fund</h2><p className="progress-number">{profile.coins}</p><p>coins available</p><Button asChild variant="outline" className="mt-4"><Link to="/store">Explore seeds</Link></Button></section>
      </div>
      <section aria-label="Seed milestones" className="space-y-4">
        <div><h2 className="text-2xl font-extrabold">Something new around the corner</h2><p className="text-sm text-muted-foreground mt-2">Seeds unlock as you level up. Buy them with coins in the store.</p></div>
        {error ? <div role="alert" className="game-panel p-5"><p>Unable to load seed milestones. Your XP and garden unlocks are still shown above.</p><Button variant="outline" className="mt-3" onClick={() => { setLoading(true); setRetry(value => value + 1); }}>Retry milestones</Button></div> : loading ? <p role="status">Loading milestones...</p> : <div className="progress-seed-tiers">{levels.map(level => {
          const unlocked = profile.level >= level;
          return <article className={`game-panel seed-tier ${unlocked ? 'is-unlocked' : ''}`} aria-label={`Level ${level} seed unlocks`} key={level}>
            <div className="flex justify-between gap-3"><h3 className="font-extrabold">Level {level}</h3><span className="text-xs font-bold flex items-center gap-1">{unlocked ? <><Leaf size={14} />Available now</> : <><LockKeyhole size={14} />Reach level {level}</>}</span></div>
            <ul className="mt-4 space-y-2">{catalog.filter(plant => plant.unlock_level === level).map(plant => <li key={plant.id} className="flex items-center gap-2 text-sm"><span aria-hidden="true">{plant.icon}</span>{plant.name}</li>)}</ul>
          </article>;
        })}</div>}
      </section>
      <div className="grid sm:grid-cols-2 gap-5">
        <section className="game-panel p-5 space-y-3"><h2 className="font-extrabold">Keep your story growing</h2><p className="text-sm">Plant seeds to earn XP. Harvesting pays XP based on the plant's base growing time in minutes.</p><p className="text-sm text-muted-foreground">Harvest rewards go to the player who grew the plant, including in shared gardens. Ripe plants wait until you harvest them.</p></section>
        <section className="game-panel p-5 space-y-3" aria-label="Account information"><h2 className="font-extrabold flex items-center gap-2"><User size={18} /> Your account</h2><p className="text-sm break-words">{profile.first_name || profile.last_name ? `${profile.first_name ?? ''} ${profile.last_name ?? ''}`.trim() : profile.username}</p><p className="text-sm text-muted-foreground break-all">{profile.email}</p><p className="text-xs text-muted-foreground">Gardening since {new Date(profile.created_at).toLocaleDateString()}</p><Button asChild variant="outline"><Link to="/settings">Edit profile</Link></Button></section>
      </div>
    </main>
  );
}
