import { Suspense } from 'react';

import { AuthFlowScreen } from '@/components/auth/AuthFlowScreen';

export default function RegistrationPage() {
  return (
    <Suspense fallback={<main className="auth-shell"><div className="auth-card glass-card shimmer-card"><p>Preparing your registration flow…</p></div></main>}>
      <AuthFlowScreen mode="registration" />
    </Suspense>
  );
}
