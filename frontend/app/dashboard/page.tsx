'use client';

import { useEffect, useState, Suspense, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { AppSidebar } from '@/components/dashboard/AppSidebar';
import ChatPane from '@/components/dashboard/ChatPane';
import TopNavBar from '@/components/dashboard/TopNavBar';
import OffersPane from '@/components/dashboard/OffersPane';
import { Thread, TrackedOffer, Dealer, dashboardAPI, dealersAPI } from '@/lib/api';
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar';
import { toast } from 'sonner';

type ViewMode = 'chat' | 'offers';

function DashboardContent() {
  const { user, loading, refreshUser } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [threads, setThreads] = useState<Thread[]>([]);
  const [offers, setOffers] = useState<TrackedOffer[]>([]);
  const [dealers, setDealers] = useState<Dealer[]>([]);
  const [selectedThreadId, setSelectedThreadId] = useState<string | null>(null);
  const [selectedOfferId, setSelectedOfferId] = useState<string | null>(null);
  const [loadingDashboard, setLoadingDashboard] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('chat');

  // Check if user just connected Gmail
  useEffect(() => {
    const gmailConnected = searchParams.get('gmail_connected');
    if (gmailConnected === 'true') {
      // Refresh user data to get updated Gmail connection status
      refreshUser();
      // Show success toast
      toast.success('Gmail connected successfully!');
      // Clean up URL
      router.replace('/dashboard');
    }
  }, [searchParams, router, refreshUser]);

  useEffect(() => {
    if (!loading) {
      if (!user) {
        router.push('/login');
      } else if (!user.preferences) {
        router.push('/onboarding');
      }
    }
  }, [user, loading, router]);

  useEffect(() => {
    if (user && user.preferences) {
      loadDashboard();
    }
  }, [user]);

  // Polling for new threads
  const threadsRef = useRef<Thread[]>([]);
  const offersRef = useRef<TrackedOffer[]>([]);

  // Keep refs in sync with state
  useEffect(() => {
    threadsRef.current = threads;
  }, [threads]);

  useEffect(() => {
    offersRef.current = offers;
  }, [offers]);

  useEffect(() => {
    if (!user || !user.preferences) return;

    const POLL_INTERVAL = 20000; // 20 seconds
    let pollInterval: NodeJS.Timeout | null = null;
    let isPolling = true;

    const pollForThreads = async () => {
      // Only poll if tab is visible
      if (document.visibilityState === 'hidden') {
        return;
      }

      try {
        const dashboardData = await dashboardAPI.getDashboard();
        
        // Update threads if they changed
        setThreads(prev => {
          const prevIds = new Set(prev.map(t => t.id));
          const hasChanges = dashboardData.threads.some(t => !prevIds.has(t.id)) ||
            dashboardData.threads.length !== prev.length ||
            dashboardData.threads.some(t => {
              const prevThread = prev.find(p => p.id === t.id);
              return prevThread && (prevThread.unreadCount !== t.unreadCount || prevThread.lastMessageAt !== t.lastMessageAt);
            });
          return hasChanges ? dashboardData.threads : prev;
        });

        setOffers(prev => {
          const prevIds = new Set(prev.map(o => o.id));
          const hasChanges = dashboardData.offers.some(o => !prevIds.has(o.id)) ||
            dashboardData.offers.length !== prev.length;
          return hasChanges ? dashboardData.offers : prev;
        });
      } catch (error) {
        console.error('Failed to poll for new threads:', error);
      }
    };

    const startPolling = () => {
      if (isPolling && document.visibilityState === 'visible') {
        pollInterval = setInterval(pollForThreads, POLL_INTERVAL);
      }
    };

    const stopPolling = () => {
      if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    };

    // Handle visibility changes
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        // Poll immediately when tab becomes visible, then start interval
        pollForThreads();
        startPolling();
      } else {
        stopPolling();
      }
    };

    // Start polling
    startPolling();
    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      isPolling = false;
      stopPolling();
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [user]);


  const loadDashboard = async () => {
    setLoadingDashboard(true);
    try {
      const [dashboardData, dealersData] = await Promise.all([
        dashboardAPI.getDashboard(),
        dealersAPI.getDealers().catch(() => []), // Don't fail if dealers endpoint fails
      ]);
      setThreads(dashboardData.threads);
      setOffers(dashboardData.offers);
      setDealers(dealersData);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoadingDashboard(false);
    }
  };


  const handleThreadArchived = (threadId: string) => {
    // Immediately remove the archived thread from the threads list
    setThreads(prev => prev.filter(thread => thread.id !== threadId));

    // Clear selection if this was the selected thread
    if (selectedThreadId === threadId) {
      setSelectedThreadId(null);
    }

    // Refresh dashboard to ensure consistency with server
    loadDashboard();
  };

  const handleThreadUpdated = (threadId: string, newName: string) => {
    // Update the thread in the threads list
    setThreads(prev => prev.map(thread =>
      thread.id === threadId
        ? {
            ...thread,
            sellerName: newName,
            displayName: newName === thread.phone || newName === '' ? thread.phone : newName
          }
        : thread
    ));
  };

  const handleThreadCreated = (newThread: Thread) => {
    setThreads([...threads, newThread]);
    setSelectedThreadId(newThread.id);
    setSelectedOfferId(null);
  };

  const handleThreadSelect = (threadId: string) => {
    setSelectedThreadId(threadId);
    setSelectedOfferId(null);
    setViewMode('chat');
  };

  const handleOfferSelect = (offer: TrackedOffer) => {
    setSelectedOfferId(offer.id);
    setSelectedThreadId(offer.threadId);
    setViewMode('chat');
  };

  const handleViewOffers = () => {
    setViewMode('offers');
    setSelectedThreadId(null);
    setSelectedOfferId(null);
  };

  const handleNavigateToThread = (threadId: string) => {
    setSelectedThreadId(threadId);
    setSelectedOfferId(null);
    setViewMode('chat');
  };

  const handleGoToDashboard = () => {
    setSelectedThreadId(null);
    setSelectedOfferId(null);
    setViewMode('chat');
  };

  // Calculate total unread count across all threads
  const totalUnreadCount = threads.reduce((sum, thread) => sum + (thread.unreadCount || 0), 0);

  if (loading || loadingDashboard) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (!user || !user.preferences) {
    return null;
  }

  return (
    <SidebarProvider
      defaultOpen
      style={
        {
          '--sidebar-width': '22rem',
          '--header-height': 'calc(var(--spacing) * 12)',
        } as React.CSSProperties
      }
    >
      <AppSidebar
        variant="inset"
        threads={threads}
        selectedThreadId={selectedThreadId}
        offers={offers}
        selectedOfferId={selectedOfferId}
        totalUnreadCount={totalUnreadCount}
        onThreadSelect={handleThreadSelect}
        onThreadCreated={handleThreadCreated}
        onOfferSelect={handleOfferSelect}
        onGoToDashboard={handleGoToDashboard}
      />
      <SidebarInset className="overflow-x-hidden">
        {/* <TopNavBar onViewOffers={handleViewOffers} /> */}
        <div className="flex flex-1 flex-col overflow-x-hidden">
          {viewMode === 'offers' ? (
            <OffersPane onNavigateToThread={handleNavigateToThread} />
          ) : (
            <ChatPane
              selectedThreadId={selectedThreadId}
              threads={threads}
              offers={offers}
              dealers={dealers}
              onThreadArchived={handleThreadArchived}
              onThreadUpdated={handleThreadUpdated}
              onNavigateToThread={handleNavigateToThread}
              onOfferDeleted={loadDashboard}
              onDealersUpdated={loadDashboard}
            />
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default function DashboardPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-600">Loading...</div>
      </div>
    }>
      <DashboardContent />
    </Suspense>
  );
}
