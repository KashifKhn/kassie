import type { ApiError } from './types';

export type ExportFormat = 'CSV' | 'JSON';

interface ExportChunk {
  data: string;
  rowsExported: number;
  done: boolean;
}

interface ExportOptions {
  keyspace: string;
  table: string;
  whereClause?: string;
  format: ExportFormat;
  onProgress?: (rowsExported: number) => void;
}

function handleExportError(error: unknown): ApiError {
  if (error instanceof Error) {
    return { code: 'EXPORT_ERROR', message: error.message, details: {} };
  }
  return { code: 'EXPORT_ERROR', message: 'Export failed', details: {} };
}

export async function streamExport(options: ExportOptions): Promise<number> {
  const token = readAuthToken();

  const response = await fetch(buildExportUrl(), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({
      keyspace: options.keyspace,
      table: options.table,
      whereClause: options.whereClause || '',
      format: options.format === 'CSV' ? 'EXPORT_FORMAT_CSV' : 'EXPORT_FORMAT_JSON',
      fetchSize: 1000,
    }),
  });

  if (!response.ok || !response.body) {
    let message = `Export failed (${response.status})`;
    try {
      const body = await response.json();
      if (body?.message) message = body.message;
    } catch {
      // keep default message
    }
    throw handleExportError(new Error(message));
  }

  const chunks: Uint8Array[] = [];
  let rowsExported = 0;
  let buffer = '';
  const decoder = new TextDecoder();

  const reader = response.body.getReader();
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() ?? '';

    for (const line of lines) {
      const chunk = parseChunk(line);
      if (!chunk) continue;
      if (chunk.rowsExported > rowsExported) {
        rowsExported = chunk.rowsExported;
        options.onProgress?.(rowsExported);
      }
      if (chunk.data) {
        chunks.push(base64ToBytes(chunk.data));
      }
    }
  }

  if (buffer.trim()) {
    const chunk = parseChunk(buffer);
    if (chunk?.data) {
      chunks.push(base64ToBytes(chunk.data));
      rowsExported = chunk.rowsExported;
    }
  }

  const blob = new Blob(chunks as BlobPart[], {
    type: options.format === 'CSV' ? 'text/csv' : 'application/x-ndjson',
  });
  triggerDownload(blob, exportFileName(options.keyspace, options.table, options.format));

  return rowsExported;
}

function buildExportUrl(): string {
  const base = import.meta.env.VITE_API_URL || '/api/v1';
  return `${base}/data/export`;
}

function readAuthToken(): string | null {
  const raw = localStorage.getItem('auth-storage');
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as { state?: { accessToken?: string } };
    return parsed.state?.accessToken ?? null;
  } catch {
    return null;
  }
}

function parseChunk(line: string): ExportChunk | null {
  const trimmed = line.trim();
  if (!trimmed) return null;
  try {
    const parsed = JSON.parse(trimmed) as Partial<ExportChunk>;
    return {
      data: parsed.data ?? '',
      rowsExported: parsed.rowsExported ?? 0,
      done: parsed.done ?? false,
    };
  } catch {
    return null;
  }
}

function base64ToBytes(b64: string): Uint8Array {
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

export function exportFileName(keyspace: string, table: string, format: ExportFormat): string {
  const stamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  return `kassie-${keyspace}-${table}-${stamp}.${format.toLowerCase()}`;
}

function triggerDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
