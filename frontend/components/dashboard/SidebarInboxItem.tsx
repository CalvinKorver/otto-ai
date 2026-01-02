'use client';

// This component is deprecated - inbox messages are now consolidated into threads
// Keeping for reference but not actively used

interface InboxMessage {
  id: string;
  sender: 'user' | 'agent' | 'seller';
  senderEmail?: string;
  senderPhone?: string;
  subject?: string;
  content: string;
  timestamp: string;
  externalMessageId?: string;
  messageType?: 'EMAIL' | 'PHONE';
}

interface SidebarInboxItemProps {
  message: InboxMessage;
  isActive: boolean;
  onSelect: (message: InboxMessage) => void;
}

export function SidebarInboxItem({
  message,
  isActive,
  onSelect,
}: SidebarInboxItemProps) {
  const isSMS = message.messageType === 'PHONE';
  
  // For SMS, show content snippet; for email, show subject
  const displayText = isSMS 
    ? (message.content || 'No message')
    : (message.subject || 'No Subject');
  
  // For SMS, show phone number; for email, show email
  const senderInfo = isSMS
    ? (message.senderPhone || 'Unknown number')
    : (message.senderEmail || 'Unknown sender');

  return (
    <div
      className={`cursor-pointer rounded-md p-1.5 hover:bg-accent transition-colors ${
        isActive ? 'bg-accent' : ''
      }`}
      onClick={() => onSelect(message)}
    >
      <div className="font-medium text-sm truncate leading-tight">
        {displayText}
      </div>
      <div className="text-xs text-muted-foreground truncate leading-tight">
        {senderInfo}
      </div>
    </div>
  );
}
