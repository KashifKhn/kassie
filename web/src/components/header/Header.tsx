import { Moon, Sun, Monitor, Menu, PanelRight, LogOut, Database } from 'lucide-react';
import { useUiStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';

interface HeaderProps {
  onLogout?: () => void;
}

export function Header({ onLogout }: HeaderProps) {
  const { theme, setTheme, toggleSidebar, toggleInspector } = useUiStore();
  const { profile } = useAuthStore();

  const cycleTheme = () => {
    const themes: Array<'light' | 'dark' | 'system'> = [
      'light',
      'dark',
      'system',
    ];
    const currentIndex = themes.indexOf(theme);
    const nextIndex = (currentIndex + 1) % themes.length;
    const nextTheme = themes[nextIndex];
    if (nextTheme) {
      setTheme(nextTheme);
    }
  };

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor;

  return (
    <div 
      className="h-full flex items-center justify-between px-6"
      style={{
        background: 'var(--bg-elevated)',
        borderBottom: '1px solid var(--border-primary)'
      }}
    >
      <div className="flex items-center gap-6">
        <button
          onClick={toggleSidebar}
          className="p-2 rounded-lg transition-all duration-200 hover:scale-105 group"
          style={{
            color: 'var(--text-secondary)'
          }}
          title="Toggle Sidebar"
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-tertiary)';
            e.currentTarget.style.color = 'var(--text-primary)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
            e.currentTarget.style.color = 'var(--text-secondary)';
          }}
        >
          <Menu className="w-5 h-5" />
        </button>

        <div className="flex items-center gap-3">
          <Database className="w-5 h-5" style={{ color: 'var(--accent-primary)' }} />
          <h1 className="font-mono text-lg font-bold tracking-wide" style={{ color: 'var(--text-primary)' }}>
            KASSIE
          </h1>
          {profile && (
            <>
              <div className="w-px h-4" style={{ background: 'var(--border-secondary)' }} />
              <span className="font-mono text-sm" style={{ color: 'var(--text-tertiary)' }}>
                {profile.name}
              </span>
            </>
          )}
        </div>
      </div>

      <div className="flex items-center gap-3">
        {profile && (
          <div className="hidden md:flex items-center gap-2 font-mono text-xs px-3 py-1.5 rounded-md" 
            style={{ 
              background: 'var(--bg-tertiary)',
              color: 'var(--text-secondary)'
            }}
          >
            <span>{profile.hosts.join(', ')}</span>
            <span style={{ color: 'var(--text-tertiary)' }}>:</span>
            <span>{profile.port}</span>
          </div>
        )}

        <button
          onClick={cycleTheme}
          className="p-2 rounded-lg transition-all duration-200 hover:scale-105 group relative"
          style={{
            color: 'var(--text-secondary)'
          }}
          title={`Theme: ${theme}`}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-tertiary)';
            e.currentTarget.style.color = 'var(--accent-primary)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
            e.currentTarget.style.color = 'var(--text-secondary)';
          }}
        >
          <ThemeIcon className="w-5 h-5 transition-transform group-hover:rotate-12" />
        </button>

        <button
          onClick={toggleInspector}
          className="p-2 rounded-lg transition-all duration-200 hover:scale-105"
          style={{
            color: 'var(--text-secondary)'
          }}
          title="Toggle Inspector"
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-tertiary)';
            e.currentTarget.style.color = 'var(--text-primary)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
            e.currentTarget.style.color = 'var(--text-secondary)';
          }}
        >
          <PanelRight className="w-5 h-5" />
        </button>

        {onLogout && (
          <button
            onClick={onLogout}
            className="p-2 rounded-lg transition-all duration-200 hover:scale-105 group"
            style={{
              color: 'var(--text-secondary)'
            }}
            title="Logout"
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'var(--error)';
              e.currentTarget.style.color = 'white';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'transparent';
              e.currentTarget.style.color = 'var(--text-secondary)';
            }}
          >
            <LogOut className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}
