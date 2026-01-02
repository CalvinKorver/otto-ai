'use client';

import { useState, useEffect } from 'react';
import { IconMail, IconMessageCircle, IconPlus, IconCopy, IconCheck, IconCurrencyDollar, IconHelpCircle } from '@tabler/icons-react';
import { Thread, TrackedOffer, threadAPI } from '@/lib/api';
import { useAuth } from '@/contexts/AuthContext';
import { SidebarOfferItem } from './SidebarOfferItem';
import { NavUser } from './NavUser';
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarFooter,
  SidebarHeader,
} from '@/components/ui/sidebar';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import * as React from 'react';

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  threads: Thread[];
  selectedThreadId: string | null;
  offers: TrackedOffer[];
  selectedOfferId: string | null;
  totalUnreadCount?: number;
  onThreadSelect: (threadId: string) => void;
  onThreadCreated: (thread: Thread) => void;
  onOfferSelect: (offer: TrackedOffer) => void;
  onGoToDashboard?: () => void;
}

export function AppSidebar({
  threads,
  selectedThreadId,
  offers,
  selectedOfferId,
  totalUnreadCount = 0,
  onThreadSelect,
  onThreadCreated,
  onOfferSelect,
  onGoToDashboard,
  ...props
}: AppSidebarProps) {
  const { user } = useAuth();
  const [showEmailDialog, setShowEmailDialog] = useState(false);
  const [showNewThreadDialog, setShowNewThreadDialog] = useState(false);
  const [copied, setCopied] = useState(false);

  // New thread form state
  const [newSellerName, setNewSellerName] = useState('');
  const [newSellerType, setNewSellerType] = useState<'private' | 'dealership' | 'other'>('dealership');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCopyEmail = async () => {
    if (user?.inboxEmail) {
      try {
        await navigator.clipboard.writeText(user.inboxEmail);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (err) {
        console.error('Failed to copy email:', err);
      }
    }
  };

  const handleCreateThread = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!newSellerName.trim()) {
      setError('Seller name is required');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const newThread = await threadAPI.create({
        sellerName: newSellerName.trim(),
        sellerType: newSellerType,
      });

      onThreadCreated(newThread);
      setNewSellerName('');
      setNewSellerType('dealership');
      setShowNewThreadDialog(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create thread');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Sidebar {...props}>

        <SidebarHeader className="border-sidebar-border h-16 border-b">
          <NavUser 
            user={{
              name: user?.email?.split('@')[0] || 'User',
              email: user?.email || ''
            }}
            onGoToDashboard={onGoToDashboard}
          />
        </SidebarHeader>
        <SidebarContent>

          {/* Messages Section */}
          <SidebarGroup>
            <SidebarGroupLabel className="flex items-center justify-between text-semibold">
              <span className="flex items-center gap-2 ">
                <IconMail className="h-4 w-4" />
                Messages
                {totalUnreadCount > 0 && (
                  <span className="flex items-center justify-center min-w-[1.25rem] h-5 px-1.5 text-xs font-medium text-primary-foreground bg-primary rounded-full">
                    {totalUnreadCount > 99 ? '99+' : totalUnreadCount}
                  </span>
                )}
              </span>
              {user?.inboxEmail && (
                <IconHelpCircle
                  onClick={() => setShowEmailDialog(true)}
                  className="h-4 w-4 cursor-pointer hover:text-foreground text-muted-foreground transition-colors"
                />
              )}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <div className="flex flex-col gap-1 px-2">
                {threads.length === 0 ? (
                  <div className="px-2 py-1 text-sm text-muted-foreground italic">
                    No messages
                  </div>
                ) : (
                  threads.map((thread) => (
                    <div
                      key={thread.id}
                      onClick={() => onThreadSelect(thread.id)}
                      className={`cursor-pointer rounded-md p-1.5 hover:bg-accent transition-colors ${
                        selectedThreadId === thread.id ? 'bg-accent' : ''
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {thread.unreadCount > 0 && (
                          <span className="flex-shrink-0 w-2 h-2 rounded-full bg-primary" />
                        )}
                        <div className="flex-1 min-w-0">
                          <div className="font-medium text-sm truncate leading-tight">
                            {thread.displayName || thread.sellerName || thread.phone || 'Unknown'}
                          </div>
                          {thread.lastMessagePreview && (
                            <div className="text-xs text-muted-foreground truncate leading-tight mt-0.5">
                              {thread.lastMessagePreview}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </SidebarGroupContent>
          </SidebarGroup>

          {/* Offers Section */}
          <SidebarGroup>
            <SidebarGroupLabel className="flex items-center gap-2 text-semibold">
              <IconCurrencyDollar className="h-4 w-4" />
              Offers
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <div className="flex flex-col gap-1 px-2">
                {offers.length === 0 ? (
                  <div className="px-2 py-1 text-sm text-muted-foreground italic">
                    No tracked offers
                  </div>
                ) : (
                  offers.map((offer) => (
                    <SidebarOfferItem
                      key={offer.id}
                      offer={offer}
                      isActive={selectedOfferId === offer.id}
                      onSelect={onOfferSelect}
                    />
                  ))
                )}
              </div>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter className="border-t p-4">
          <Button
            onClick={() => setShowNewThreadDialog(true)}
            className="w-full"
          >
            <IconPlus className="h-4 w-4 mr-2" />
            New Seller Thread
          </Button>
        </SidebarFooter>
      </Sidebar>

      {/* Email Dialog */}
      <Dialog open={showEmailDialog} onOpenChange={setShowEmailDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Forward Emails</DialogTitle>
            <DialogDescription>
              Forward or BCC emails from sellers to this address and they'll appear in your inbox:
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2 p-3 bg-muted rounded-md">
            <code className="flex-1 text-sm break-all">{user?.inboxEmail}</code>
            <Button
              onClick={handleCopyEmail}
              variant="outline"
              size="sm"
            >
              {copied ? (
                <>
                  <IconCheck className="h-4 w-4 mr-2" />
                  Copied
                </>
              ) : (
                <>
                  <IconCopy className="h-4 w-4 mr-2" />
                  Copy
                </>
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* New Thread Dialog */}
      <Dialog open={showNewThreadDialog} onOpenChange={setShowNewThreadDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Seller Thread</DialogTitle>
            <DialogDescription>
              Start a thread to organize and track your communication with this seller.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateThread} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="sellerType">Seller Type</Label>
              <Select
                value={newSellerType}
                onValueChange={(value) => setNewSellerType(value as any)}
                disabled={loading}
              >
                <SelectTrigger id="sellerType">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="dealership">Dealership</SelectItem>
                  <SelectItem value="private">Private Seller</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="sellerName">Seller Name</Label>
              <Input
                id="sellerName"
                value={newSellerName}
                onChange={(e) => setNewSellerName(e.target.value)}
                placeholder="e.g., Subaru of Renton"
                disabled={loading}
              />
            </div>

            {error && (
              <div className="text-sm text-destructive">{error}</div>
            )}

            <div className="flex gap-2">
              <Button type="submit" disabled={loading} className="flex-1">
                {loading ? 'Creating...' : 'Create'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setShowNewThreadDialog(false);
                  setNewSellerName('');
                  setError('');
                }}
                disabled={loading}
                className="flex-1"
              >
                Cancel
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
