import { useEffect, useState } from 'react';
import JsonView from '@uiw/react-json-view';
import { X, Copy, Check, Binary, Braces, Type } from 'lucide-react';
import type { CellValue } from '@/api/types';

interface CellDetailModalProps {
  columnName: string;
  cell: CellValue | null;
  onClose: () => void;
}

export function CellDetailModal({ columnName, cell, onClose }: CellDetailModalProps) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose]);

  if (!cell) {
    return null;
  }

  const kind = valueKind(cell);
  const display = fullValueText(cell);
  const jsonValue = tryParseJSON(display);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(display);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-8"
      style={{ background: 'rgba(0, 0, 0, 0.6)' }}
      onClick={onClose}
    >
      <div
        className="flex flex-col max-w-3xl w-full max-h-[80vh] rounded-xl glass animate-scale-in overflow-hidden"
        style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--border-primary)',
          boxShadow: 'var(--shadow-lg)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div
          className="flex-shrink-0 px-5 py-4 flex items-center justify-between"
          style={{ borderBottom: '1px solid var(--border-primary)' }}
        >
          <div className="flex items-center gap-3">
            <h3
              className="text-base font-mono font-bold"
              style={{ color: 'var(--text-primary)' }}
            >
              {columnName}
            </h3>
            {cell.cqlType && (
              <span
                className="flex items-center gap-1.5 text-xs font-mono px-2 py-1 rounded"
                style={{
                  background: 'var(--accent-subtle)',
                  color: 'var(--accent-primary)',
                }}
              >
                <Type className="w-3 h-3" />
                {cell.cqlType}
              </span>
            )}
            <span
              className="text-xs font-mono px-2 py-1 rounded"
              style={{
                background: 'var(--bg-tertiary)',
                color: 'var(--text-tertiary)',
              }}
            >
              {kind}
            </span>
          </div>
          <div className="flex items-center gap-2">
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
                  Copy
                </>
              )}
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-lg transition-all hover:scale-105"
              style={{ color: 'var(--text-secondary)' }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--bg-tertiary)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent';
              }}
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-auto p-5">
          {cell.isNull ? (
            <p
              className="text-sm font-mono italic"
              style={{ color: 'var(--text-tertiary)' }}
            >
              NULL
            </p>
          ) : 'bytesVal' in cell ? (
            <div className="space-y-3">
              <div
                className="flex items-center gap-2 text-xs font-mono"
                style={{ color: 'var(--text-secondary)' }}
              >
                <Binary className="w-3.5 h-3.5" />
                {cell.bytesVal.length} bytes
              </div>
              <pre
                className="text-xs font-mono p-4 rounded-lg overflow-auto whitespace-pre"
                style={{
                  background: 'var(--bg-primary)',
                  border: '1px solid var(--border-primary)',
                  color: 'var(--text-primary)',
                  lineHeight: '1.6',
                }}
              >
                {formatHexDump(cell.bytesVal)}
              </pre>
            </div>
          ) : jsonValue !== null ? (
            <div className="space-y-3">
              <div
                className="flex items-center gap-2 text-xs font-mono"
                style={{ color: 'var(--text-secondary)' }}
              >
                <Braces className="w-3.5 h-3.5" />
                JSON collection
              </div>
              <div
                className="rounded-lg p-4 overflow-auto"
                style={{
                  background: 'var(--bg-primary)',
                  border: '1px solid var(--border-primary)',
                }}
              >
                <JsonView
                  value={jsonValue}
                  collapsed={2}
                  displayDataTypes={false}
                  displayObjectSize={false}
                  style={{
                    fontSize: '13px',
                    fontFamily: 'var(--font-mono)',
                    background: 'transparent',
                    '--w-rjv-color': 'var(--text-primary)',
                    '--w-rjv-key-string': 'var(--accent-primary)',
                    '--w-rjv-background-color': 'transparent',
                    '--w-rjv-border-left': '1px solid var(--border-primary)',
                    '--w-rjv-line-color': 'var(--border-primary)',
                    '--w-rjv-arrow-color': 'var(--text-tertiary)',
                    '--w-rjv-info-color': 'var(--text-tertiary)',
                    '--w-rjv-type-string-color': '#22c55e',
                    '--w-rjv-type-int-color': '#f59e0b',
                    '--w-rjv-type-float-color': '#f59e0b',
                    '--w-rjv-type-boolean-color': '#a855f7',
                    '--w-rjv-type-null-color': '#ef4444',
                  } as React.CSSProperties}
                />
              </div>
            </div>
          ) : (
            <pre
              className="text-sm font-mono whitespace-pre-wrap break-all"
              style={{ color: 'var(--text-primary)', lineHeight: '1.6' }}
            >
              {display}
            </pre>
          )}
        </div>

        <div
          className="flex-shrink-0 px-5 py-3 text-xs font-mono"
          style={{
            borderTop: '1px solid var(--border-primary)',
            color: 'var(--text-tertiary)',
          }}
        >
          esc to close
        </div>
      </div>
    </div>
  );
}

function valueKind(cell: CellValue): string {
  if (cell.isNull) return 'null';
  if ('stringVal' in cell) return 'string';
  if ('intVal' in cell) return 'int';
  if ('doubleVal' in cell) return 'double';
  if ('boolVal' in cell) return 'bool';
  if ('bytesVal' in cell) return 'blob';
  return 'unknown';
}

function fullValueText(cell: CellValue): string {
  if (cell.isNull) return '';
  if ('stringVal' in cell) return cell.stringVal;
  if ('intVal' in cell) return cell.intVal.toString();
  if ('doubleVal' in cell) return cell.doubleVal.toString();
  if ('boolVal' in cell) return cell.boolVal ? 'true' : 'false';
  if ('bytesVal' in cell) return formatHexDump(cell.bytesVal);
  return '';
}

function tryParseJSON(text: string): unknown {
  if (!text) return null;
  const trimmed = text.trim();
  if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) {
    return null;
  }
  try {
    return JSON.parse(trimmed);
  } catch {
    return null;
  }
}

function formatHexDump(bytes: Uint8Array): string {
  const lines: string[] = [];
  const bytesPerLine = 16;

  for (let offset = 0; offset < bytes.length; offset += bytesPerLine) {
    const chunk = Array.from(bytes.slice(offset, offset + bytesPerLine));
    const hex = chunk
      .map((b) => b.toString(16).padStart(2, '0'))
      .join(' ')
      .padEnd(bytesPerLine * 3 - 1, ' ');
    const ascii = chunk
      .map((b) => (b >= 0x20 && b < 0x7f ? String.fromCharCode(b) : '.'))
      .join('');
    lines.push(`${offset.toString(16).padStart(8, '0')}  ${hex}  |${ascii}|`);
  }

  return lines.join('\n');
}
