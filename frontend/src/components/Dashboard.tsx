import { useState } from 'react';
import { Garden } from '../types/api';
import { GardenDetail } from './GardenDetail';
import { GardenList } from './GardenList';
import { UserProfile } from './UserProfile';
import { WeatherWidget } from './WeatherWidget';
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs';
import { Card } from './ui/card';
import { useNavigate } from 'react-router-dom';

export function Dashboard() {
  const navigate = useNavigate();
  // Remove selectedGarden state and GardenDetail rendering

  return (
    <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-8">
        <div className="lg:col-span-3 order-2 lg:order-1">
          <Tabs defaultValue="gardens" className="w-full">
            <TabsList className="mb-6">
              <TabsTrigger value="gardens">My Gardens</TabsTrigger>
              <TabsTrigger value="profile">Profile</TabsTrigger>
            </TabsList>
            <TabsContent value="gardens">
              <GardenList onSelectGarden={(garden) => navigate(`/garden/${garden.id}`)} />
            </TabsContent>
            <TabsContent value="profile">
              <UserProfile />
            </TabsContent>
          </Tabs>
        </div>

        <div className="space-y-4 sm:space-y-6 order-1 lg:order-2">
          <WeatherWidget />

          <Card className="rounded-xl shadow-sm border border-border p-6">
            <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Quick Tips</h3>
            <div className="space-y-3">
              <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
                <p className="text-sm text-blue-800 dark:text-blue-300">
                  💧 Water your plants regularly to keep them healthy
                </p>
              </div>
              <div className="bg-yellow-50 dark:bg-yellow-900/20 p-3 rounded-lg">
                <p className="text-sm text-yellow-800 dark:text-yellow-300">
                  ⚡ Use fertilizer to speed up growth
                </p>
              </div>
              <div className="bg-green-50 dark:bg-green-900/20 p-3 rounded-lg">
                <p className="text-sm text-green-500 dark:text-green-300">
                  🌾 Harvest plants when they're ready for maximum rewards
                </p>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}