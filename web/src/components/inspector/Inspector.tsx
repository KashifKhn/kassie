import { useState } from 'react';
import JsonView from '@uiw/react-json-view';
import { Copy, Check } from 'lucide-react';
import type { Row, CellValue } from '@/api/types';

interface InspectorProps {
  row: Row | null;
}

export function Inspector({ row }: InspectorProps) {
  const [copied, setCopied] = useState(false);

  if (!row) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground p-4 text-center">
        <div>
          <p className="text-sm">No row selected</p>
          <p className="text-xs mt-2">Click a row in the table to inspect</p>
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
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex-shrink-0 border-b border-border p-4 flex items-center justify-between">
        <h3 className="text-sm font-semibold">Row Inspector</h3>
        <button
          onClick={handleCopy}
          className="flex items-center gap-2 px-3 py-1 text-xs bg-secondary hover:bg-secondary/80 rounded-md transition-colors"
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
        <div className="space-y-4">
          <div>
            <h4 className="text-xs font-semibold text-muted-foreground mb-2">
              KEY-VALUE PAIRS
            </h4>
            <div className="space-y-1">
              {Object.entries(row.cells).map(([key, value]) => (
                <div
                  key={key}
                  className="flex gap-2 text-sm py-1 border-b border-border"
                >
                  <span className="font-medium text-primary min-w-[120px]">
                    {key}:
                  </span>
                  <span className="flex-1 text-foreground font-mono">
                    {formatCellValue(value)}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h4 className="text-xs font-semibold text-muted-foreground mb-2">
              JSON VIEW
            </h4>
            <div className="bg-muted rounded-md p-2 overflow-auto">
              <JsonView
                value={convertRowToJSON(row)}
                collapsed={1}
                displayDataTypes={false}
                displayObjectSize={false}
                style={{
                  fontSize: '12px',
                  fontFamily: 'monospace',
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
