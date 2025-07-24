import React from 'react';
import { Card } from './ui/card';
import { Badge } from './ui/badge';
import { Progress } from './ui/progress';
import { useAuth } from '../contexts/AuthContext';

export const UserProfile: React.FC = () => {
  const { user } = useAuth();

  if (!user) {
    return null;
  }

  return (
    <Card className="p-6 space-y-4">
      <div className="flex items-center space-x-4">
        {user.avatar && (
          <img
            src={user.avatar}
            alt={user.username}
            className="w-16 h-16 rounded-full"
          />
        )}
        <div>
          <h2 className="text-2xl font-bold">{user.username}</h2>
          <p className="text-gray-500">
            {user.first_name} {user.last_name}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 mt-4">
        <div className="text-center">
          <div className="text-2xl font-bold">{user.level}</div>
          <div className="text-sm text-gray-500">Level</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold">{user.experience}</div>
          <div className="text-sm text-gray-500">XP</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold">{user.coins}</div>
          <div className="text-sm text-gray-500">Coins</div>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>Level Progress</span>
          <span>{user.level_progress}%</span>
        </div>
        <Progress value={user.level_progress} className="h-2" />
      </div>

      <div className="flex flex-wrap gap-2 mt-4">
        {user.achievements?.map((achievement) => (
          <Badge
            key={achievement.id}
            variant="secondary"
            className="px-3 py-1"
          >
            {achievement.name}
          </Badge>
        ))}
      </div>
    </Card>
  );
};