import { useEffect } from 'react';
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react';
import { useToastStore } from '@/stores/toastStore';

export function ToastContainer() {
  const { toasts, removeToast } = useToastStore();

  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-3" style={{ maxWidth: '400px' }}>
      {toasts.map((toast, index) => (
        <Toast
          key={toast.id}
          id={toast.id}
          type={toast.type}
          message={toast.message}
          onClose={() => removeToast(toast.id)}
          index={index}
        />
      ))}
    </div>
  );
}

interface ToastProps {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  onClose: () => void;
  index: number;
}

function Toast({ type, message, onClose, index }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, 3000);
    return () => clearTimeout(timer);
  }, [onClose]);

  const config = {
    success: {
      icon: CheckCircle,
      bgColor: 'var(--success)',
      iconColor: 'var(--success)',
    },
    error: {
      icon: AlertCircle,
      bgColor: 'var(--error)',
      iconColor: 'var(--error)',
    },
    info: {
      icon: Info,
      bgColor: 'var(--info)',
      iconColor: 'var(--info)',
    },
    warning: {
      icon: AlertTriangle,
      bgColor: 'var(--warning)',
      iconColor: 'var(--warning)',
    },
  }[type];

  const Icon = config.icon;

  return (
    <div
      className="flex items-center gap-3 px-4 py-3 rounded-lg glass font-mono animate-slide-up"
      style={{
        background: 'var(--bg-elevated)',
        border: `1px solid ${config.bgColor}`,
        boxShadow: `0 0 20px ${config.bgColor}40, var(--shadow-lg)`,
        color: 'var(--text-primary)',
        animationDelay: `${index * 100}ms`,
      }}
    >
      <Icon 
        className="h-5 w-5 flex-shrink-0 animate-pulse" 
        style={{ 
          color: config.iconColor,
          filter: `drop-shadow(0 0 10px ${config.iconColor})`,
        }}
      />
      <p className="flex-1 text-sm font-medium">{message}</p>
      <button
        onClick={onClose}
        className="flex-shrink-0 transition-all"
        style={{ 
          color: 'var(--text-tertiary)',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = 'var(--error)';
          e.currentTarget.style.transform = 'scale(1.1)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = 'var(--text-tertiary)';
          e.currentTarget.style.transform = 'scale(1)';
        }}
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
