import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/store/authSlice';
import type { ReactElement } from 'react';

export function PrivateRoute({ children }: Readonly<{ children: ReactElement }>) {
  const user = useAuthStore((s) => s.user);

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
