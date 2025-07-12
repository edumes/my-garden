import { Plus } from 'lucide-react';
import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';

interface CreateGardenModalProps {
  onClose: () => void;
  onSubmit: (data: { name: string; description?: string }) => void;
}

export function CreateGardenModal({ onClose, onSubmit }: CreateGardenModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });
  const [open, setOpen] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      name: formData.name,
      description: formData.description || undefined,
    });
    setOpen(false);
    setFormData({ name: '', description: '' });
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      onClose();
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="flex items-center justify-center space-x-2 bg-green-600 hover:bg-green-700 text-white px-3 sm:px-4 py-2 text-sm sm:text-base rounded-lg transition-colors w-full sm:w-auto">
          <Plus className="w-4 h-4" />
          New Garden
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create New Garden</DialogTitle>
          <DialogDescription>
            Create a new garden to start growing your plants. Give it a name and optional description.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="garden-name">Garden Name *</Label>
              <Input
                id="garden-name"
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Beautiful Garden"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="garden-description">Description</Label>
              <Textarea
                id="garden-description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="A place to grow beautiful plants..."
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button type="submit" className="bg-green-600 hover:bg-green-700">
              Create Garden
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}