import type { ReactNode } from "react";
import { Group, Panel, Separator } from "react-resizable-panels";
import { useUiStore } from "@/stores/uiStore";

interface LayoutProps {
  header: ReactNode;
  sidebar: ReactNode;
  main: ReactNode;
  inspector: ReactNode;
}

export function Layout({ header, sidebar, main, inspector }: LayoutProps) {
  const { sidebarCollapsed, inspectorCollapsed } = useUiStore();

  return (
    <div
      className="h-screen w-screen flex flex-col overflow-hidden"
      style={{ background: "var(--bg-primary)" }}
    >
      <header className="h-14 flex-shrink-0">{header}</header>

      <div className="flex-1 overflow-hidden">
        <Group orientation="horizontal">
          {!sidebarCollapsed && (
            <>
              <Panel
                defaultSize={300}
                minSize={50}
                maxSize={400}
                style={{ background: "var(--bg-secondary)" }}
              >
                {sidebar}
              </Panel>
              <Separator
                className="w-1 transition-colors duration-200 hover:cursor-col-resize"
                style={{
                  background: "var(--border-primary)",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = "var(--accent-primary)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = "var(--border-primary)";
                }}
              />
            </>
          )}

          <Panel defaultSize={60} minSize={100}>
            {main}
          </Panel>

          {!inspectorCollapsed && (
            <>
              <Separator
                className="w-1 transition-colors duration-200 hover:cursor-col-resize"
                style={{
                  background: "var(--border-primary)",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = "var(--accent-primary)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = "var(--border-primary)";
                }}
              />
              <Panel
                defaultSize={300}
                minSize={50}
                maxSize={400}
                style={{ background: "var(--bg-secondary)" }}
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
