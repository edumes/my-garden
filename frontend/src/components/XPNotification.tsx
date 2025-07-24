import React from 'react';
import { toast } from 'sonner';
import { Trophy } from 'lucide-react';

interface XPNotificationProps {
  xpGained: number;
  newLevel?: number;
  oldLevel?: number;
  coinsEarned?: number;
}

export const showXPNotification = ({
  xpGained,
  newLevel,
  oldLevel,
  coinsEarned,
}: XPNotificationProps) => {
  const leveledUp = newLevel && oldLevel && newLevel > oldLevel;

  toast(
    <div className="flex items-center space-x-3">
      <div className="flex-shrink-0">
        <Trophy className="w-6 h-6 text-yellow-500" />
      </div>
      <div>
        <div className="font-medium">
          {leveledUp ? '🎉 Level Up!' : '+XP Gained!'}
        </div>
        <div className="text-sm text-gray-500 space-y-1">
          <div>+{xpGained} XP</div>
          {coinsEarned && <div>+{coinsEarned} Coins</div>}
          {leveledUp && (
            <div className="text-green-500">
              Level {oldLevel} → {newLevel}
            </div>
          )}
        </div>
      </div>
    </div>,
    {
      duration: 3000,
      className: 'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700',
    }
  );
}; 