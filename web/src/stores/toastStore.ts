import { create } from 'zustand';

type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
  id: string;
  type: ToastType;
  message: string;
  duration?: number;
}

interface ToastState {
  toasts: Toast[];
}

interface ToastActions {
  addToast: (type: ToastType, message: string, duration?: number) => void;
  removeToast: (id: string) => void;
  success: (message: string, duration?: number) => void;
  error: (message: string, duration?: number) => void;
  info: (message: string, duration?: number) => void;
  warning: (message: string, duration?: number) => void;
}

export const useToastStore = create<ToastState & ToastActions>((set) => ({
  toasts: [],

  addToast: (type, message, duration = 3000) => {
    const id = Math.random().toString(36).substring(7);
    const toast: Toast = { id, type, message, duration };

    set((state) => ({
      toasts: [...state.toasts, toast],
    }));

    if (duration > 0) {
      setTimeout(() => {
        set((state) => ({
          toasts: state.toasts.filter((t) => t.id !== id),
        }));
      }, duration);
    }
  },

  removeToast: (id) => {
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    }));
  },

  success: (message, duration) => {
    useToastStore.getState().addToast('success', message, duration);
  },

  error: (message, duration) => {
    useToastStore.getState().addToast('error', message, duration);
  },

  info: (message, duration) => {
    useToastStore.getState().addToast('info', message, duration);
  },

  warning: (message, duration) => {
    useToastStore.getState().addToast('warning', message, duration);
  },
}));
