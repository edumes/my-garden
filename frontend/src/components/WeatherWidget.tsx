import { Cloud, CloudRain, Droplets, Sun, Thermometer, Wind } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Card } from './ui/card';

interface Weather {
  temperature: number;
  humidity: number;
  condition: string;
  season: string;
  timestamp: string;
}

const weatherIcons = {
  sunny: Sun,
  cloudy: Cloud,
  rainy: CloudRain,
  windy: Wind,
};

export function WeatherWidget() {
  const [weather, setWeather] = useState<Weather | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadWeather();
  }, []);

  const loadWeather = async () => {
    try {
      const response = await apiService.getCurrentWeather();
      setWeather(response.weather);
    } catch (error) {
      console.error('Failed to load weather:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6 animate-pulse">
        <div className="h-6 bg-muted rounded w-24 mb-4"></div>
        <div className="h-16 bg-muted rounded"></div>
      </div>
    );
  }

  if (!weather) {
    return (
      <div className="bg-card rounded-xl shadow-sm border border-border p-4 sm:p-6">
        <h3 className="text-base sm:text-lg font-semibold text-foreground mb-4">Weather</h3>
        <p className="text-muted-foreground">Weather data unavailable</p>
      </div>
    );
  }

  const WeatherIcon = weatherIcons[weather.condition as keyof typeof weatherIcons] || Cloud;

  return (
    <Card className="rounded-xl shadow-sm border border-border p-4 sm:p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-base sm:text-lg font-semibold text-foreground">Weather</h3>
        <span className="text-sm text-muted-foreground capitalize">{weather.season}</span>
      </div>

      <div className="flex items-center space-x-4">
        <div className="flex-shrink-0">
          <WeatherIcon className="w-10 h-10 sm:w-12 sm:h-12 text-blue-500" />
        </div>
        <div className="flex-1">
          <div className="flex items-center space-x-2 mb-2">
            <Thermometer className="w-4 h-4 text-red-500" />
            <span className="text-lg sm:text-xl font-bold text-foreground">{weather.temperature.toFixed(0)}°C</span>
          </div>
          <div className="flex items-center space-x-2">
            <Droplets className="w-4 h-4 text-blue-500" />
            <span className="text-sm text-muted-foreground">{weather.humidity}% humidity</span>
          </div>
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-border">
        <p className="text-sm text-muted-foreground capitalize">
          Current conditions: {weather.condition}
        </p>
      </div>
    </Card>
  );
}