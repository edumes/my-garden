import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiService } from '@/services/api';
import { Garden } from '@/types/api';
import { ExternalLink, Users } from 'lucide-react';
import React, { useState } from 'react';

interface JoinGardenModalProps {
  onGardenJoined?: (garden: Garden) => void;
}

const JoinGardenModal: React.FC<JoinGardenModalProps> = ({ onGardenJoined }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [token, setToken] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleJoinGarden = async () => {
    if (!token.trim()) {
      setError('Please enter an access token');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await apiService.joinGarden({ token: token.trim() });
      setToken('');
      setIsOpen(false);
      onGardenJoined?.(response.garden);
    } catch (error: any) {
      setError(error.message || 'Failed to join garden');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasteFromUrl = () => {
    const url = window.location.href;
    const tokenMatch = url.match(/\/join-garden\/([a-zA-Z0-9-]+)/);
    if (tokenMatch) {
      setToken(tokenMatch[1]);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <Users className="w-4 h-4 mr-2" />
          Join Garden
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Join a Garden</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label htmlFor="token">Access Token</Label>
            <div className="flex space-x-2">
              <Input
                id="token"
                placeholder="Enter access token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleJoinGarden()}
              />
              <Button
                variant="outline"
                size="sm"
                onClick={handlePasteFromUrl}
                title="Extract token from URL"
              >
                <ExternalLink className="w-4 h-4" />
              </Button>
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              Enter the access token from the garden owner's share link
            </p>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <Button
            onClick={handleJoinGarden}
            disabled={isLoading || !token.trim()}
            className="w-full"
          >
            {isLoading ? 'Joining...' : 'Join Garden'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default JoinGardenModal; 