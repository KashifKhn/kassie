import { Moon, Sun, Monitor, Menu, PanelRight, LogOut } from 'lucide-react';
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
    <div className="h-full flex items-center justify-between px-4 bg-background border-b border-border">
      <div className="flex items-center gap-4">
        <button
          onClick={toggleSidebar}
          className="p-2 hover:bg-accent rounded-md transition-colors"
          title="Toggle Sidebar"
        >
          <Menu className="w-5 h-5" />
        </button>

        <div className="flex items-center gap-2">
          <h1 className="text-lg font-bold text-primary">Kassie</h1>
          {profile && (
            <span className="text-sm text-muted-foreground">
              {profile.name}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2">
        {profile && (
          <div className="text-sm text-muted-foreground hidden md:flex items-center gap-2">
            <span>{profile.hosts.join(', ')}</span>
            <span>·</span>
            <span>:{profile.port}</span>
          </div>
        )}

        <button
          onClick={cycleTheme}
          className="p-2 hover:bg-accent rounded-md transition-colors"
          title={`Theme: ${theme}`}
        >
          <ThemeIcon className="w-5 h-5" />
        </button>

        <button
          onClick={toggleInspector}
          className="p-2 hover:bg-accent rounded-md transition-colors"
          title="Toggle Inspector"
        >
          <PanelRight className="w-5 h-5" />
        </button>

        {onLogout && (
          <button
            onClick={onLogout}
            className="p-2 hover:bg-accent text-destructive rounded-md transition-colors"
            title="Logout"
          >
            <LogOut className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}
