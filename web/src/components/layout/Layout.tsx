import type { ReactNode } from 'react';
import { Group, Panel, Separator } from 'react-resizable-panels';
import { useUiStore } from '@/stores/uiStore';

interface LayoutProps {
  header: ReactNode;
  sidebar: ReactNode;
  main: ReactNode;
  inspector: ReactNode;
}

export function Layout({ header, sidebar, main, inspector }: LayoutProps) {
  const { sidebarCollapsed, inspectorCollapsed } = useUiStore();

  return (
    <div className="h-screen w-screen flex flex-col overflow-hidden bg-background">
      <header className="h-14 border-b border-border flex-shrink-0">
        {header}
      </header>

      <div className="flex-1 overflow-hidden">
        <Group orientation="horizontal">
          {!sidebarCollapsed && (
            <>
              <Panel
                defaultSize={20}
                minSize={15}
                maxSize={30}
                className="bg-background"
              >
                {sidebar}
              </Panel>
              <Separator className="w-1 bg-border hover:bg-primary transition-colors" />
            </>
          )}

          <Panel defaultSize={60} minSize={30}>
            {main}
          </Panel>

          {!inspectorCollapsed && (
            <>
              <Separator className="w-1 bg-border hover:bg-primary transition-colors" />
              <Panel
                defaultSize={20}
                minSize={15}
                maxSize={40}
                className="bg-background"
              >
                {inspector}
              </Panel>
            </>
          )}
        </Group>
      </div>
    </div>
  );
}
