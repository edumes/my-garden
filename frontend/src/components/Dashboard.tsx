import { useNavigate } from 'react-router-dom';
import { GardenList } from './GardenList';
import { UserProfile } from './UserProfile';
import { WeatherWidget } from './WeatherWidget';
import { Tabs, TabsContent, TabsContents, TabsList, TabsTrigger } from './animate-ui/components/tabs';
import { Card } from './ui/card';

export function Dashboard() {
  const navigate = useNavigate();

  return (
    <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-8">
        <div className="lg:col-span-3 order-2 lg:order-1">
          <Tabs defaultValue="gardens" className="w-full">
            <TabsList className="mb-6 bg-muted/50 backdrop-blur-sm border border-border/50">
              <TabsTrigger value="gardens" className="text-sm font-medium">
                🌱 My Gardens
              </TabsTrigger>
              <TabsTrigger value="profile" className="text-sm font-medium">
                👤 Profile
              </TabsTrigger>
            </TabsList>
            <TabsContents className="mt-6">
              <TabsContent value="gardens">
                <GardenList onSelectGarden={(garden) => navigate(`/garden/${garden.id}`)} />
              </TabsContent>
              <TabsContent value="profile">
                <UserProfile />
              </TabsContent>
            </TabsContents>
          </Tabs>
        </div>

        <div className="space-y-4 sm:space-y-6 order-1 lg:order-2">
          <WeatherWidget />

          <Card className="rounded-xl shadow-sm border border-border p-6">
            <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Quick Tips</h3>
            <div className="space-y-3">
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