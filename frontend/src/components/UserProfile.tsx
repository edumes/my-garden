import { Calendar, Coins, Medal, Star, Trophy } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';

export function UserProfile() {
  const { user } = useAuth();
  const [profile, setProfile] = useState(user);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    setLoading(true);
    try {
      const response = await apiService.getUserProfile();
      setProfile(response.user);
    } catch (error) {
      console.error('Failed to load profile:', error);
    } finally {
      setLoading(false);
    }
  };

  if (!profile) return null;

  const experienceToNextLevel = (profile.level * 100) - profile.experience;
  const progressPercentage = (profile.experience % 100);

  return (
    <div className="space-y-4 sm:space-y-6">
      <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
        <div className="flex items-center space-x-4 mb-6">
          <div className="w-12 h-12 sm:w-16 sm:h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center">
            <span className="text-lg sm:text-2xl font-bold text-green-600 dark:text-green-400">
              {profile.username?.charAt(0).toUpperCase()}
            </span>
          </div>
          <div>
            <h2 className="text-lg sm:text-2xl font-bold text-foreground">
              {profile.first_name && profile.last_name
                ? `${profile.first_name} ${profile.last_name}`
                : profile.username}
            </h2>
            <p className="text-sm sm:text-base text-muted-foreground">@{profile.username}</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-6">
          <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Trophy className="w-5 h-5 text-blue-600" />
              <span className="font-medium text-blue-900 dark:text-blue-100">Level</span>
            </div>
            <div className="text-xl sm:text-2xl font-bold text-blue-900 dark:text-blue-100">{profile.level}</div>
            <div className="w-full bg-blue-200 dark:bg-blue-700 rounded-full h-2 mt-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${progressPercentage}%` }}
              ></div>
            </div>
            <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
              {experienceToNextLevel} XP to next level
            </p>
          </div>

          <div className="bg-yellow-50 dark:bg-yellow-900/20 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Coins className="w-5 h-5 text-yellow-600" />
              <span className="font-medium text-yellow-900 dark:text-yellow-100">Coins</span>
            </div>
            <div className="text-xl sm:text-2xl font-bold text-yellow-900 dark:text-yellow-100">{profile.coins}</div>
          </div>

          <div className="bg-green-50 dark:bg-green-900/20 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Star className="w-5 h-5 text-green-600" />
              <span className="font-medium text-green-900 dark:text-green-100">Experience</span>
            </div>
            <div className="text-xl sm:text-2xl font-bold text-green-900 dark:text-green-100">{profile.experience}</div>
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
        <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Account Information</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Email</label>
            <p className="text-sm sm:text-base text-foreground">{profile.email}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Member Since</label>
            <div className="flex items-center space-x-2">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <p className="text-sm sm:text-base text-foreground">
                {new Date(profile.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
          {profile.last_login_at && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Last Login</label>
              <p className="text-sm sm:text-base text-foreground">
                {new Date(profile.last_login_at).toLocaleDateString()}
              </p>
            </div>
          )}
        </div>
      </div>

      {profile.achievements && profile.achievements.length > 0 && (
        <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
          <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Achievements</h3>
          <div className="grid grid-cols-1 gap-3 sm:gap-4">
            {profile.achievements.map((userAchievement) => (
              <div
                key={userAchievement.id}
                className="flex items-center space-x-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-700"
              >
                <div className="flex-shrink-0">
                  <Medal className="w-6 h-6 text-yellow-600" />
                </div>
                <div className="flex-1">
                  <h4 className="text-sm sm:text-base font-medium text-foreground">
                    {userAchievement.achievement.name}
                  </h4>
                  <p className="text-sm text-muted-foreground">
                    {userAchievement.achievement.description}
                  </p>
                  <div className="flex flex-col sm:flex-row sm:items-center sm:space-x-2 mt-1 space-y-1 sm:space-y-0">
                    <span className="text-xs bg-yellow-200 dark:bg-yellow-800 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded-full">
                      {userAchievement.achievement.points} points
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {new Date(userAchievement.unlocked_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}