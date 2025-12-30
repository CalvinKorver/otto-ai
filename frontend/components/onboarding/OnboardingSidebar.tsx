'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { useTheme } from 'next-themes';

interface OnboardingSidebarProps {
  currentStep: number;
}

export default function OnboardingSidebar({ currentStep }: OnboardingSidebarProps) {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Use dark logo in light mode, light logo in dark mode
  const logoSrc = mounted && resolvedTheme === 'light' 
    ? '/logo-dark-v2.png' 
    : '/logo-light-v2.png';

  const steps = [
    { label: 'Welcome!', step: 0 },
    { label: 'Quick Questions', step: 1 },
    { label: 'Lets go', step: 2 },
  ];

  return (
    <div className="w-full lg:w-64 bg-card border-r border-border p-6 flex lg:flex-col gap-6">
      <div className="hidden lg:block mb-8">
        <Image
          src={logoSrc}
          alt="Otto"
          width={200}
          height={60}
          className="h-12 w-auto"
        />
      </div>

      <nav className="flex lg:flex-col gap-4 lg:gap-6 flex-1 lg:flex-none">
        {steps.map((step, index) => {
          const isActive = currentStep >= step.step;
          const isCurrent = currentStep === step.step;

          return (
            <div key={step.label} className="flex items-center gap-3">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold transition-all ${
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'border-2 border-border text-muted-foreground'
                } ${isCurrent ? 'ring-4 ring-primary/20' : ''}`}
              >
                {isActive ? '✓' : index + 1}
              </div>
              <div
                className={`text-sm font-medium transition-colors ${
                  isActive ? 'text-foreground' : 'text-muted-foreground'
                }`}
              >
                {step.label}
              </div>
            </div>
          );
        })}
      </nav>
    </div>
  );
}
