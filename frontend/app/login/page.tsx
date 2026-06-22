import { Suspense } from 'react';

import { AuthFlowScreen } from '@/components/auth/AuthFlowScreen';

export default function LoginPage() {
  return (
    <Suspense fallback={<main className="auth-shell"><div className="auth-card glass-card shimmer-card"><p>Preparing your login flow…</p></div></main>}>
      <AuthFlowScreen mode="login" />
    </Suspense>
  );
}
