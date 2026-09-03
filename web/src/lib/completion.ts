import type { Column } from '@/api/types';

export type SuggestionKind = 'keyword' | 'keyspace' | 'table' | 'column';

export interface Suggestion {
  label: string;
  detail?: string;
  kind: SuggestionKind;
}

export interface CompletionSources {
  keyspaces: string[];
  tablesFor: (keyspace: string) => string[];
  columnsFor: (keyspace: string, table: string) => Column[];
  defaultKeyspace?: string;
  defaultTable?: string;
}

const CQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'IN', 'CONTAINS', 'LIMIT',
  'ALLOW', 'FILTERING', 'ORDER', 'BY', 'ASC', 'DESC', 'NULL', 'TOKEN',
  'COUNT', 'MIN', 'MAX', 'SUM', 'AVG', 'DISTINCT', 'AS', 'JSON',
  'PER', 'PARTITION',
];

const WORD_BREAK = /[\s.,()=<>*;']/;

function tokenize(text: string): string[] {
  return text.split(WORD_BREAK).filter(Boolean);
}

function trailingWord(text: string): string {
  const match = /[^\s.,()=<>*;']*$/.exec(text);
  return match ? match[0] : '';
}

function hasToken(tokens: string[], word: string): boolean {
  return tokens.some((t) => t.toUpperCase() === word);
}

function matchPrefix(prefix: string, candidate: string): boolean {
  return candidate.toLowerCase().startsWith(prefix.toLowerCase());
}

function keywordSuggestions(prefix: string): Suggestion[] {
  return CQL_KEYWORDS.filter((kw) => !prefix || matchPrefix(prefix, kw)).map(
    (kw) => ({ label: kw, kind: 'keyword' as const })
  );
}

function resolveTargetTable(text: string): { ks: string; table: string } {
  const fromMatch = /\bfrom\s+([\w".]+)/i.exec(text);
  if (!fromMatch || !fromMatch[1]) {
    return { ks: '', table: '' };
  }
  const target = fromMatch[1];
  const dotIdx = target.indexOf('.');
  if (dotIdx >= 0) {
    return { ks: target.slice(0, dotIdx), table: target.slice(dotIdx + 1) };
  }
  return { ks: '', table: target };
}

export function complete(text: string, sources: CompletionSources): Suggestion[] {
  const lastChar = text.length > 0 ? text[text.length - 1] : undefined;
  const endsMidWord = lastChar !== undefined && !WORD_BREAK.test(lastChar);
  const trimmed = text.replace(/[\s]+$/, '');
  const tokens = tokenize(trimmed);

  if (!tokens.length) {
    return keywordSuggestions('');
  }

  const lastUpper = tokens[tokens.length - 1]?.toUpperCase() ?? '';
  const prefix = endsMidWord ? trailingWord(trimmed) : '';

  // "ks." → tables of keyspace
  if (trimmed.endsWith('.')) {
    const ks = tokens[tokens.length - 1] ?? '';
    return sources
      .tablesFor(ks)
      .map((tbl) => ({ label: `${ks}.${tbl}`, kind: 'table' as const }));
  }

  // Mid-word right after a dot: "app_data.u"
  if (endsMidWord) {
    const dotIdx = trimmed.lastIndexOf('.');
    if (dotIdx >= 0) {
      const ks = trimmed.slice(0, dotIdx).split(WORD_BREAK).pop() ?? '';
      const tblPrefix = trimmed.slice(dotIdx + 1);
      return sources
        .tablesFor(ks)
        .filter((tbl) => tbl.toLowerCase().startsWith(tblPrefix.toLowerCase()))
        .map((tbl) => ({ label: `${ks}.${tbl}`, kind: 'table' as const }));
    }
  }

  const hasFrom = hasToken(tokens, 'FROM');
  const hasWhere = hasToken(tokens, 'WHERE');

  // FROM clause: keyspaces (+ default-keyspace tables)
  if (lastUpper === 'FROM' || (hasFrom && !hasWhere)) {
    const out: Suggestion[] = sources.keyspaces
      .filter((ks) => !prefix || matchPrefix(prefix, ks))
      .map((ks) => ({ label: ks, kind: 'keyspace' as const }));
    if (sources.defaultKeyspace) {
      for (const tbl of sources.tablesFor(sources.defaultKeyspace)) {
        if (!prefix || matchPrefix(prefix, tbl)) {
          out.push({ label: tbl, kind: 'table' });
        }
      }
    }
    return out;
  }

  // Column context: after SELECT, in WHERE, after AND/OR/BY
  const inSelectList = !hasFrom && tokens[0]?.toUpperCase() === 'SELECT';
  if (
    lastUpper === 'SELECT' ||
    lastUpper === 'WHERE' ||
    lastUpper === 'AND' ||
    lastUpper === 'OR' ||
    lastUpper === 'BY' ||
    inSelectList ||
    (hasFrom && hasWhere)
  ) {
    let { ks, table } = resolveTargetTable(trimmed);
    if ((!ks || !table) && sources.defaultKeyspace && sources.defaultTable) {
      ks = sources.defaultKeyspace;
      table = sources.defaultTable;
    }

    const out: Suggestion[] = [];
    if (ks && table) {
      for (const col of sources.columnsFor(ks, table)) {
        if (!prefix || matchPrefix(prefix, col.name)) {
          out.push({ label: col.name, detail: col.type, kind: 'column' });
        }
      }
    }
    for (const kw of ['AND', 'OR', 'IN', 'CONTAINS', 'LIMIT', 'ORDER', 'ALLOW', 'FILTERING']) {
      if (!prefix || matchPrefix(prefix, kw)) {
        out.push({ label: kw, kind: 'keyword' });
      }
    }
    return out;
  }

  return keywordSuggestions(prefix);
}
