import { Calendar, Share2, TreePine, Users } from 'lucide-react';
import { Garden, GardenPermission } from '../types/api';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Card } from './ui/card';

interface GardenCardProps {
  garden: Garden;
  permission?: GardenPermission;
  isShared?: boolean;
  sharedBy?: string;
  onClick: () => void;
  onEdit: () => void;
  onShare?: () => void;
}

export function GardenCard({
  garden,
  permission,
  isShared,
  sharedBy,
  onClick,
  // onEdit,
  onShare
}: GardenCardProps) {
  const plantCount = garden.plants?.length || 0;

  return (
    <Card className="rounded-xl shadow-sm border border-border hover:shadow-md transition-shadow cursor-pointer">
      <div onClick={onClick} className="p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <h3 className="text-lg font-semibold text-foreground">{garden.name}</h3>
            {isShared && (
              <Badge variant="secondary" className="text-xs">
                <Users className="w-3 h-3 mr-1" />
                Shared
              </Badge>
            )}
            {permission && (
              <Badge variant="outline" className="text-xs">
                {permission}
              </Badge>
            )}
          </div>
          <div className="flex items-center space-x-1">
            {onShare && (
              <Button
                onClick={(e) => {
                  e.stopPropagation();
                  onShare();
                }}
                variant="ghost"
                className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
              >
                <Share2 className="w-4 h-4" />
              </Button>
            )}
            {/* <Button
              onClick={(e) => {
                e.stopPropagation();
                onEdit();
              }}
              variant="ghost"
              className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
            >
              <Settings className="w-4 h-4" />
            </Button> */}
          </div>
        </div>

        {garden.description && (
          <p className="text-muted-foreground text-sm mb-4">{garden.description}</p>
        )}

        {isShared && sharedBy && (
          <p className="text-muted-foreground text-sm mb-4">
            Shared by: {sharedBy}
          </p>
        )}

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div className="flex items-center space-x-2">
            <TreePine className="w-4 h-4 text-green-500" />
            <span className="text-sm text-muted-foreground">{plantCount} plants</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {new Date(garden.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>
    </Card>
  );
}