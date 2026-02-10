import { useState } from 'react';
import JsonView from '@uiw/react-json-view';
import { Copy, Check, FileJson } from 'lucide-react';
import type { Row, CellValue } from '@/api/types';

interface InspectorProps {
  row: Row | null;
}

export function Inspector({ row }: InspectorProps) {
  const [copied, setCopied] = useState(false);

  if (!row) {
    return (
      <div 
        className="h-full flex items-center justify-center p-4 text-center"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div className="animate-fade-in">
          <FileJson 
            className="w-12 h-12 mx-auto mb-3" 
            style={{ color: 'var(--text-tertiary)' }}
          />
          <p 
            className="text-sm font-mono"
            style={{ color: 'var(--text-secondary)' }}
          >
            No row selected
          </p>
          <p 
            className="text-xs font-mono mt-2"
            style={{ color: 'var(--text-tertiary)' }}
          >
            Click a row in the table to inspect
          </p>
        </div>
      </div>
    );
  }

  const handleCopy = async () => {
    const jsonData = convertRowToJSON(row);
    await navigator.clipboard.writeText(JSON.stringify(jsonData, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div 
      className="h-full flex flex-col overflow-hidden"
      style={{ background: 'var(--bg-primary)' }}
    >
      <div 
        className="flex-shrink-0 px-4 py-3 flex items-center justify-between"
        style={{ 
          borderBottom: '1px solid var(--border-primary)',
          background: 'var(--bg-secondary)',
        }}
      >
        <h3 
          className="text-sm font-mono font-bold"
          style={{ color: 'var(--text-primary)' }}
        >
          Row Inspector
        </h3>
        <button
          onClick={handleCopy}
          className="flex items-center gap-2 px-3 py-1.5 text-xs font-mono rounded-md transition-all"
          style={{
            background: 'var(--accent-subtle)',
            color: 'var(--accent-primary)',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--accent-primary)';
            e.currentTarget.style.color = 'var(--text-inverse)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'var(--accent-subtle)';
            e.currentTarget.style.color = 'var(--accent-primary)';
          }}
        >
          {copied ? (
            <>
              <Check className="w-3 h-3" />
              Copied
            </>
          ) : (
            <>
              <Copy className="w-3 h-3" />
              Copy JSON
            </>
          )}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="space-y-6 animate-fade-in">
          <div>
            <h4 
              className="text-xs font-mono font-bold tracking-wider uppercase mb-3"
              style={{ color: 'var(--text-tertiary)' }}
            >
              Key-Value Pairs
            </h4>
            <div className="space-y-2">
              {Object.entries(row.cells).map(([key, value]) => (
                <div
                  key={key}
                  className="flex gap-3 text-sm py-2 transition-all"
                  style={{
                    borderBottom: '1px solid var(--border-primary)',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'var(--bg-secondary)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'transparent';
                  }}
                >
                  <span 
                    className="font-mono font-bold min-w-[120px]"
                    style={{ color: 'var(--accent-primary)' }}
                  >
                    {key}
                  </span>
                  <span 
                    className="flex-1 font-mono"
                    style={{ color: 'var(--text-primary)' }}
                  >
                    {formatCellValue(value)}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h4 
              className="text-xs font-mono font-bold tracking-wider uppercase mb-3"
              style={{ color: 'var(--text-tertiary)' }}
            >
              JSON View
            </h4>
            <div 
              className="rounded-md p-3 overflow-auto terminal"
              style={{
                background: 'var(--bg-tertiary)',
                border: '1px solid var(--border-primary)',
              }}
            >
              <JsonView
                value={convertRowToJSON(row)}
                collapsed={1}
                displayDataTypes={false}
                displayObjectSize={false}
                style={{
                  fontSize: '12px',
                  fontFamily: 'var(--font-mono)',
                  background: 'transparent',
                }}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function formatCellValue(value: CellValue): string {
  if (value.isNull) {
    return 'NULL';
  }

  if ('stringVal' in value) return `"${value.stringVal}"`;
  if ('intVal' in value) return value.intVal.toString();
  if ('doubleVal' in value) return value.doubleVal.toString();
  if ('boolVal' in value) return value.boolVal.toString();
  if ('bytesVal' in value) return '<bytes>';

  return '';
}

function convertRowToJSON(row: Row): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(row.cells)) {
    if (value.isNull) {
      result[key] = null;
    } else if ('stringVal' in value) {
      result[key] = value.stringVal;
    } else if ('intVal' in value) {
      result[key] = value.intVal;
    } else if ('doubleVal' in value) {
      result[key] = value.doubleVal;
    } else if ('boolVal' in value) {
      result[key] = value.boolVal;
    } else if ('bytesVal' in value) {
      result[key] = '<bytes>';
    }
  }

  return result;
}
