'use client';

import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';

import { beginLoginFlow, beginRegistrationFlow, fetchLoginFlow, fetchRegistrationFlow } from '@/lib/ory';
import type { OryFlow } from '@/types/ory';
import { OryFlowForm } from './OryFlowForm';

export function AuthFlowScreen({ mode }: { mode: 'login' | 'registration' }) {
  const searchParams = useSearchParams();
  const flowId = searchParams.get('flow');
  const [flow, setFlow] = useState<OryFlow | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!flowId) {
      if (mode === 'login') beginLoginFlow();
      else beginRegistrationFlow();
      return;
    }

    const load = async () => {
      try {
        setError(null);
        const nextFlow =
          mode === 'login'
            ? await fetchLoginFlow(flowId)
            : await fetchRegistrationFlow(flowId);
        setFlow(nextFlow);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Could not load authentication flow.');
      }
    };

    void load();
  }, [flowId, mode]);

  return (
    <main className="auth-shell">
      <div className="orb orb--pink" />
      <div className="orb orb--violet" />
      <div className="orb orb--gold" />

      <div className="auth-shell__content">
        <div className="hero-copy">
          <span className="pill">Sukoon vision studio</span>
          <h1>Bring your bucket to life with color, calm, and AI storytelling.</h1>
          <p>
            Sign in with Ory, step into your own image space, and let the backend weave every picture into a beautiful description.
          </p>
          <div className="hero-links">
            <Link className="ghost-link" href="/">
              ← Back home
            </Link>
            <Link className="ghost-link" href={mode === 'login' ? '/registration' : '/login'}>
              {mode === 'login' ? 'Need an account?' : 'Already have an account?'}
            </Link>
          </div>
        </div>

        {error ? (
          <div className="auth-card glass-card">
            <div className="toast toast--error">{error}</div>
            <button className="candy-button" onClick={() => (mode === 'login' ? beginLoginFlow() : beginRegistrationFlow())}>
              Try again
            </button>
          </div>
        ) : flow ? (
          <OryFlowForm
            flow={flow}
            title={mode === 'login' ? 'Welcome back ✨' : 'Create your joyful corner 🌈'}
            subtitle={
              mode === 'login'
                ? 'Your private image bucket, your AI summaries, your calm workspace.'
                : 'Make an account and start transforming your image collection into searchable stories.'
            }
          />
        ) : (
          <div className="auth-card glass-card shimmer-card">
            <p>Loading your secure Ory flow…</p>
          </div>
        )}
      </div>
    </main>
  );
}
