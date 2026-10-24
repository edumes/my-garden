import { useNavigate } from 'react-router';
import { GardenList } from './GardenList';
import { UserProfile } from './UserProfile';
import { WeatherWidget } from './WeatherWidget';
import { Tabs, TabsContent, TabsContents, TabsList, TabsTrigger } from './animate-ui/components/tabs';

export function Dashboard() {
  const navigate = useNavigate();

  return (
    <div className="garden-page max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-8">
        <div className="lg:col-span-3 order-1">
          <Tabs defaultValue="gardens" className="w-full">
            <TabsList className="mb-6 bg-muted/50 backdrop-blur-sm border border-border/50">
              <TabsTrigger value="gardens" className="text-sm font-medium">
                🌱 Gardens
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

        <div className="space-y-4 sm:space-y-6 order-2">
          <WeatherWidget />

        </div>
      </div>
    </div>
  );
}