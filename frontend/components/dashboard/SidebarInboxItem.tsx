'use client';

import { Trash2 } from 'lucide-react';
import { InboxMessage } from '@/lib/api';
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item';
import { Button } from '@/components/ui/button';

interface SidebarInboxItemProps {
  message: InboxMessage;
  isActive: boolean;
  onSelect: (message: InboxMessage) => void;
  onDelete?: (messageId: string) => void;
}

export function SidebarInboxItem({
  message,
  isActive,
  onSelect,
}: SidebarInboxItemProps) { 

  return (
    <Item
      variant="muted"
      size="default"
      className={`cursor-pointer ${isActive ? 'bg-accent' : ''}`}
      onClick={() => onSelect(message)}
    >
      <ItemContent>
        <ItemTitle>{message.subject || 'No Subject'}</ItemTitle>
        <ItemDescription>
          {message.senderEmail}
        </ItemDescription>
      </ItemContent>

    </Item>
  );
}
