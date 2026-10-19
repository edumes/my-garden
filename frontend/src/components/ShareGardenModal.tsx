import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { DateField, DateInput } from '@/components/ui/datefield-rac';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { apiService } from '@/services/api';
import { Garden, GardenAccessLink, GardenPermission } from '@/types/api';
import { CalendarDate } from '@internationalized/date';
import { Check, Copy } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { toast } from 'sonner';

interface ShareGardenModalProps {
  garden: Garden;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onShareCreated?: () => void;
}

const ShareGardenModal: React.FC<ShareGardenModalProps> = ({ garden, open, onOpenChange, onShareCreated }) => {
  const [permissions, setPermissions] = useState<GardenPermission[]>(['view']);
  const [maxUses, setMaxUses] = useState<string>('');
  const [expiresAt, setExpiresAt] = useState<CalendarDate | null>(null);
  const [accessLinks, setAccessLinks] = useState<GardenAccessLink[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [copiedLinkId, setCopiedLinkId] = useState<string | null>(null);

  const permissionOptions = [
    { value: 'view', label: 'View', description: 'Can view the garden and plants' },
    { value: 'plant', label: 'Plant', description: 'Can plant new seeds' },
    { value: 'harvest', label: 'Harvest', description: 'Can harvest mature plants' },
    { value: 'manage', label: 'Manage', description: 'Full access to garden management' },
  ];

  useEffect(() => {
    if (open) {
      loadAccessLinks();
    }
  }, [open]);

  const loadAccessLinks = async () => {
    try {
      const response = await apiService.getAccessLinks(garden.id);
      setAccessLinks(response.access_links);
    } catch (error) {
      console.error('Failed to load access links:', error);
    }
  };

  const handlePermissionChange = (permission: GardenPermission, checked: boolean) => {
    if (checked) {
      setPermissions(prev => [...prev, permission]);
    } else {
      setPermissions(prev => prev.filter(p => p !== permission));
    }
  };

  const handleCreateLink = async () => {
    if (permissions.length === 0) {
      toast.info('Please select at least one permission');
      return;
    }

    setIsLoading(true);
    try {
      const request = {
        garden_id: garden.id,
        permissions,
        max_uses: maxUses ? parseInt(maxUses) : undefined,
        expires_at: expiresAt ? expiresAt.toString() : undefined,
      };

      const response = await apiService.createAccessLink(request);
      await loadAccessLinks();
      // onShareCreated?.();
    } catch (error) {
      console.error('Failed to create access link:', error);
      toast.error('Failed to create access link');
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = async (text: string, linkId: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedLinkId(linkId);
      setTimeout(() => setCopiedLinkId(null), 1500);
    } catch (err) {
      console.error('Failed to copy text: ', err);
      toast.error('Failed to copy to clipboard');
    }
  };

  const getShareUrl = (token: string) => {
    return `${window.location.origin}/join-garden/${token}`;
  };

  const handleDeactivateLink = async (linkId: string) => {
    try {
      await apiService.deactivateAccessLink(garden.id, linkId);
      await loadAccessLinks();
    } catch (error) {
      console.error('Failed to deactivate link:', error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Share Garden: {garden.name}</DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          <div>
            <h3 className="text-lg font-semibold mb-4">Create Access Link</h3>

            <div className="space-y-4">
              <div>
                <Label className="text-sm font-medium">Permissions</Label>
                <div className="grid grid-cols-2 gap-3 mt-2">
                  {permissionOptions.map((option) => (
                    <div key={option.value} className="flex items-start space-x-2">
                      <Checkbox
                        id={option.value}
                        checked={permissions.includes(option.value as GardenPermission)}
                        onCheckedChange={(checked: boolean) =>
                          handlePermissionChange(option.value as GardenPermission, checked)
                        }
                      />
                      <div className="space-y-1">
                        <Label htmlFor={option.value} className="text-sm font-medium">
                          {option.label}
                        </Label>
                        <p className="text-xs text-muted-foreground">{option.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="maxUses">Max Uses (optional)</Label>
                  <Input
                    id="maxUses"
                    type="number"
                    placeholder="Unlimited"
                    value={maxUses}
                    onChange={(e) => setMaxUses(e.target.value)}
                  />
                </div>
                <div>
                  <Label htmlFor="expiresAt">Expires At (optional)</Label>
                  <DateField value={expiresAt} onChange={setExpiresAt}>
                    <DateInput />
                  </DateField>
                </div>
              </div>

              <Button
                onClick={handleCreateLink}
                disabled={isLoading || permissions.length === 0}
                className="w-full"
              >
                {isLoading ? 'Creating...' : 'Create Access Link'}
              </Button>
            </div>
          </div>

          <div>
            <h3 className="text-lg font-semibold mb-4">Active Access Links</h3>
            <div className="space-y-3">
              {accessLinks.filter(link => link.is_active).map((link) => (
                <div key={link.id} className="p-3 border rounded-lg">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <Badge variant="secondary">{link.used_count} uses</Badge>
                        {link.max_uses && (
                          <Badge variant="outline">Max: {link.max_uses}</Badge>
                        )}
                        {link.expires_at && (
                          <Badge variant="outline">
                            Expires: {new Date(link.expires_at).toLocaleDateString()}
                          </Badge>
                        )}
                      </div>
                      <div className="flex items-center space-x-2">
                        <Input
                          value={getShareUrl(link.token)}
                          readOnly
                          className="flex-1 text-sm"
                        />
                        <TooltipProvider delayDuration={0}>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="outline"
                                size="sm"
                                className="disabled:opacity-100 relative"
                                onClick={() => copyToClipboard(getShareUrl(link.token), link.id)}
                                aria-label={copiedLinkId === link.id ? "Copied" : "Copy to clipboard"}
                                disabled={copiedLinkId === link.id}
                              >
                                <div
                                  className={cn(
                                    "transition-all",
                                    copiedLinkId === link.id ? "scale-100 opacity-100" : "scale-0 opacity-0",
                                  )}
                                >
                                  <Check className="stroke-emerald-500 w-4 h-4" strokeWidth={2} aria-hidden="true" />
                                </div>
                                <div
                                  className={cn(
                                    "absolute transition-all",
                                    copiedLinkId === link.id ? "scale-0 opacity-0" : "scale-100 opacity-100",
                                  )}
                                >
                                  <Copy className="w-4 h-4" strokeWidth={2} aria-hidden="true" />
                                </div>
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent className="px-2 py-1 text-xs">Click to copy</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDeactivateLink(link.id)}
                        >
                          Deactivate
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
              {accessLinks.filter(link => link.is_active).length === 0 && (
                <p className="text-muted-foreground text-center py-4">
                  No active access links
                </p>
              )}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default ShareGardenModal; 