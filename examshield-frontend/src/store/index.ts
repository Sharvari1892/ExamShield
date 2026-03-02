import { create } from "zustand";

type AppState = {
  initialized: boolean;
  setInitialized: (val: boolean) => void;
};

export const useAppStore = create<AppState>((set) => ({
  initialized: false,
  setInitialized: (val) => set({ initialized: val }),
}));
