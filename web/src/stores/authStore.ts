import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ProfileInfo } from '@/api/types';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: number | null;
  profile: ProfileInfo | null;
  isAuthenticated: boolean;
}

interface AuthActions {
  setTokens: (accessToken: string, refreshToken: string, expiresAt: number) => void;
  setProfile: (profile: ProfileInfo) => void;
  clearAuth: () => void;
  isTokenExpired: () => boolean;
}

const initialState: AuthState = {
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
  profile: null,
  isAuthenticated: false,
};

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setTokens: (accessToken, refreshToken, expiresAt) => {
        set({
          accessToken,
          refreshToken,
          expiresAt,
          isAuthenticated: true,
        });
      },

      setProfile: (profile) => {
        set({ profile });
      },

      clearAuth: () => {
        set({
          ...initialState,
        });
      },

      isTokenExpired: () => {
        const { expiresAt } = get();
        if (!expiresAt) return true;
        return Date.now() >= expiresAt;
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        expiresAt: state.expiresAt,
        profile: state.profile,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
