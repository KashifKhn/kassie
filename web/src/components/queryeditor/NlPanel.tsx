import { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Sparkles, Loader2, AlertCircle, Check, X } from 'lucide-react';
import { aiApi, queryClient } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';

interface NlPanelProps {
  onAccept: (cql: string, question: string) => void;
}

export function NlPanel({ onAccept }: NlPanelProps) {
  const [question, setQuestion] = useState('');
  const [provider, setProvider] = useState('');
  const [model, setModel] = useState('');
  const { selectedKeyspace, selectedTable } = useUiStore();
  const keyspace = selectedKeyspace ?? '';
  const table = selectedTable ?? '';

  const providersQuery = useQuery({
    queryKey: ['ai', 'providers'],
    queryFn: aiApi.detectProviders,
    staleTime: 60 * 1000,
  });

  const effectiveProvider = provider || providersQuery.data?.configured || 'auto';

  const mutation = useMutation({
    mutationFn: () =>
      aiApi.nlToCql({
        question,
        keyspace,
        table: table || undefined,
        provider: provider || undefined,
        model: model.trim() || undefined,
      }),
  });

  const accept = async () => {
    if (!mutation.data?.cql || !question) return;
    try {
      await aiApi.acceptExample({
        question,
        cql: mutation.data.cql,
        keyspace,
      });
      queryClient.invalidateQueries({ queryKey: ['ai', 'examples'] });
    } catch {
      // example saving is best-effort; still apply the query
    }
    onAccept(mutation.data.cql, question);
    mutation.reset();
    setQuestion('');
  };

  const discard = () => {
    mutation.reset();
    setQuestion('');
  };

  return (
    <div className="border-border space-y-2 border-t pt-2">
      <div className="flex items-center gap-2">
        <Sparkles className="text-muted-foreground h-4 w-4" />
        <input
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey && question.trim()) {
              e.preventDefault();
              mutation.mutate();
            }
          }}
          placeholder={`Ask in plain English about ${table || keyspace || 'your data'}...`}
          className="bg-background placeholder:text-muted-foreground focus:ring-ring w-full rounded-md border px-3 py-1.5 text-sm outline-none focus:ring-2"
          disabled={mutation.isPending}
        />
        <button
          onClick={() => mutation.mutate()}
          disabled={!question.trim() || mutation.isPending}
          className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-3 py-1.5 text-sm font-medium disabled:opacity-50"
        >
          {mutation.isPending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            'Translate'
          )}
        </button>
      </div>

      {providersQuery.data && providersQuery.data.available.length === 0 && (
        <p className="text-destructive px-1 text-xs">
          No agent CLI found on the server. Install opencode, claude, codex,
          or gemini and authenticate it to use natural language queries.
        </p>
      )}

      {providersQuery.data && providersQuery.data.available.length > 0 && (
        <div className="flex items-center gap-2 px-1 text-xs">
          <select
            value={effectiveProvider}
            onChange={(e) => setProvider(e.target.value)}
            className="bg-background border-border rounded-md border px-2 py-1 text-xs"
            title="Agent CLI used for translation"
          >
            <option value="auto">auto</option>
            {providersQuery.data.available.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
          <input
            value={model}
            onChange={(e) => setModel(e.target.value)}
            placeholder="model (default)"
            className="bg-background border-border placeholder:text-muted-foreground w-40 rounded-md border px-2 py-1 font-mono text-xs"
            title="Model passed to the CLI (--model); empty = CLI default"
          />
          {providersQuery.data.available.length > 1 && (
            <span className="text-muted-foreground">
              detected: {providersQuery.data.available.join(', ')}
            </span>
          )}
        </div>
      )}

      {mutation.isPending && (
        <p className="text-muted-foreground px-1 text-xs">
          Asking the local agent CLI for a translation... this can take a few
          seconds.
        </p>
      )}

      {mutation.isError && (
        <div className="text-destructive flex items-start gap-2 rounded-md border px-3 py-2 text-xs">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            <p className="font-medium">Translation failed</p>
            <p className="text-muted-foreground">
              {String(mutation.error?.message ?? mutation.error)}
            </p>
            <p className="text-muted-foreground mt-1">
              Make sure an agent CLI (opencode, claude, codex, or gemini) is
              installed and authenticated.
            </p>
          </div>
        </div>
      )}

      {mutation.data && (
        <div className="space-y-2 rounded-md border border-amber-500/50 bg-amber-500/5 px-3 py-2">
          <div className="flex items-center justify-between">
            <p className="text-xs font-medium text-amber-600 dark:text-amber-400">
              AI-generated CQL via {mutation.data.provider} — review before
              running
            </p>
            <div className="flex gap-1">
              <button
                onClick={accept}
                className="flex items-center gap-1 rounded-md bg-emerald-600 px-2 py-1 text-xs font-medium text-white hover:bg-emerald-700"
              >
                <Check className="h-3 w-3" /> Use query
              </button>
              <button
                onClick={discard}
                className="text-muted-foreground hover:bg-muted flex items-center gap-1 rounded-md px-2 py-1 text-xs"
              >
                <X className="h-3 w-3" /> Discard
              </button>
            </div>
          </div>
          <pre className="overflow-x-auto text-xs whitespace-pre-wrap">
            {mutation.data.cql}
          </pre>
        </div>
      )}
    </div>
  );
}
