import { create } from 'zustand';
import type { User } from '@/features/auth/types';
import { loginRequest, refreshRequest } from '@/features/auth/auth.api';

type AuthState = {
  user: User | null;
  accessToken: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refresh: () => Promise<void>;
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,

  login: async (email, password) => {
    const data = await loginRequest(email, password);
    set({
      user: data.user,
      accessToken: data.access_token,
    });
  },

  logout: () => {
    set({ user: null, accessToken: null });
  },

  refresh: async () => {
    const data = await refreshRequest();
    set({
      user: data.user,
      accessToken: data.access_token,
    });
  },
}));
