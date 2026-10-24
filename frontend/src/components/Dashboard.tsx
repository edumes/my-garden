import { useNavigate } from 'react-router';
import { GardenList } from './GardenList';
import { WeatherWidget } from './WeatherWidget';

export function Dashboard() {
  const navigate = useNavigate();

  return (
    <div className="garden-page max-w-7xl mx-auto px-3 sm:px-4 lg:px-8 py-4 sm:py-6 lg:py-8">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 sm:gap-6 lg:gap-8">
        <div className="lg:col-span-3 order-1">
          <GardenList onSelectGarden={garden => navigate(`/garden/${garden.id}`)} />
        </div>

        <div className="space-y-4 sm:space-y-6 order-2">
          <WeatherWidget />

        </div>
      </div>
    </div>
  );
}