'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import OnboardingContainer from '@/components/onboarding/OnboardingContainer';

export default function OnboardingPage() {
  const router = useRouter();
  const { user, loading, refreshUser } = useAuth();

  useEffect(() => {
    if (!loading && !user) {
      router.push('/login');
    }
  }, [user, loading, router]);

  const handleSuccess = async () => {
    await refreshUser();
    router.push('/dashboard');
  };

  if (loading || !user) {
    return null; // or a loading spinner
  }

  return <OnboardingContainer onSuccess={handleSuccess} />;
}
