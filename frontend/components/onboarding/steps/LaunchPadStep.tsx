import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import PollyAvatar from '../PollyAvatar';

interface LaunchPadStepProps {
  vehicleData: {
    make: string;
    model: string;
    year: number;
    trimName?: string | null;
  };
  onComplete: () => Promise<void>;
}

export default function LaunchPadStep({ vehicleData, onComplete }: LaunchPadStepProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPrivacyModal, setShowPrivacyModal] = useState(false);

  const handleGetStarted = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Check if token exists before making the API call
      const token = localStorage.getItem('token');
      if (!token) {
        setError('You must be logged in to continue. Redirecting to login...');
        setTimeout(() => {
          router.push('/login');
        }, 2000);
        return;
      }
      
      await onComplete();
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message || 'Failed to save preferences. Please try again.';
      
      // If it's an auth error, redirect to login
      if (errorMessage.includes('authorization') || errorMessage.includes('token') || err.response?.status === 401) {
        setError('Your session has expired. Redirecting to login...');
        setTimeout(() => {
          localStorage.removeItem('token');
          router.push('/login');
        }, 2000);
      } else {
        setError(errorMessage);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      {/* Otto's message - no bubble, just text like ChatPane */}
      <div className="flex items-start gap-1">
        <PollyAvatar />
        <div className="space-y-2 text-sm leading-relaxed text-foreground">
          <div className="font-semibold text-base">Ready to disrupt the status quo?</div>
          <div className="text-sm">
            Our AI is ready to help you save an average of $1,280 or more while shielding you from &quot;Communication Fatigue&quot;.
          </div>
          <div className="bg-muted/50 rounded-md p-3 mt-3">
            <div className="text-xs font-medium text-muted-foreground mb-1">Your Target Vehicle:</div>
            <div className="font-semibold text-sm">
              {vehicleData.year} {vehicleData.make} {vehicleData.model}
              {vehicleData.trimName && vehicleData.trimName !== 'Unspecified' && (
                <span className="font-normal"> {vehicleData.trimName}</span>
              )}
            </div>
          </div>
          <div className="text-sm text-muted-foreground mt-2">
            Before continuing, take some time to read our{' '}
            <button
              onClick={(e) => {
                e.preventDefault();
                setShowPrivacyModal(true);
              }}
              className="text-primary hover:underline underline-offset-2 font-medium cursor-pointer bg-transparent border-none p-0"
            >
              privacy notice & auto SMS policy
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-destructive/10 border border-destructive/20 text-destructive rounded-md p-3 text-sm">
          {error}
        </div>
      )}

      {/* Privacy Notice Modal */}
      <Dialog open={showPrivacyModal} onOpenChange={setShowPrivacyModal}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Privacy Notice & Auto SMS Policy</DialogTitle>
            <DialogDescription asChild>
              <div className="space-y-4 text-sm leading-relaxed">
                <p>
                  When you use our platform, we provide you with machine-generated phone numbers to
                  communicate with prospective sellers. This ensures your personal contact information
                  remains private throughout the car-buying process.
                </p>

                <div>
                  <p className="font-medium text-foreground mb-2">Key points:</p>
                  <ul className="list-disc pl-5 space-y-1">
                    <li>Your conversations are facilitated through our secure SMS system</li>
                    <li>Your real phone number is never shared with sellers</li>
                    <li>Messages are automatically processed to protect you from &quot;Communication Fatigue&quot;</li>
                    <li>All communications are stored securely and used to improve your car-buying experience</li>
                  </ul>
                </div>

                <p>
                  This platform should only be used for its intended purpose: buying a car. By using
                  our service, you agree to communicate with sellers in good faith for the purpose of
                  purchasing a vehicle.
                </p>

                <p className="text-xs text-muted-foreground">
                  We take your privacy seriously and will never sell your personal information to third parties.
                </p>
              </div>
            </DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>

      <div className="flex justify-center pt-4">
        <Button onClick={handleGetStarted} disabled={loading} size="lg" className="px-8">
          {loading ? 'Setting Up...' : 'Get Started'}
        </Button>
      </div>
    </div>
  );
}
