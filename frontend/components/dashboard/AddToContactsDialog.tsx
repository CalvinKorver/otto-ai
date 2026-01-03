import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface AddToContactsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (name: string) => Promise<void>;
  saving: boolean;
  currentName: string;
}

export default function AddToContactsDialog({
  open,
  onOpenChange,
  onConfirm,
  saving,
  currentName,
}: AddToContactsDialogProps) {
  const [name, setName] = useState(currentName);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate name is not empty
    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    setError('');
    await onConfirm(name.trim());
    // Reset on successful save (parent will close dialog)
    setName('');
  };

  const handleDiscard = () => {
    setName(currentName); // Reset to current name
    setError('');
    onOpenChange(false);
  };

  // Reset name when dialog opens with new current name
  const handleOpenChange = (newOpen: boolean) => {
    if (newOpen) {
      setName(currentName);
      setError('');
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Add to Contacts</DialogTitle>
          </DialogHeader>

          <div className="py-4">
            <Label htmlFor="contact-name" className="text-sm font-medium">
              Name
            </Label>
            <Input
              id="contact-name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                setError(''); // Clear error on input
              }}
              placeholder="Enter contact name"
              className="mt-2"
              disabled={saving}
              autoFocus
            />
            {error && (
              <p className="text-sm text-destructive mt-1">{error}</p>
            )}
          </div>

          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleDiscard}
              disabled={saving}
            >
              Discard
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
