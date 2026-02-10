import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';

interface ProtectedRouteProps {
  children: ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isTokenExpired } = useAuthStore();

  if (!isAuthenticated || isTokenExpired()) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
