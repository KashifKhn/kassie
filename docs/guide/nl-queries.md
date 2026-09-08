# Natural Language Queries (AI)

Kassie can translate plain-English questions into valid, partition-key-aware
CQL — without shipping API keys, talking to an LLM provider directly, or
adding AI dependencies to the binary.

## How it works

Instead of embedding an LLM client, Kassie uses the coding agent CLIs
already installed and authenticated on your machine as headless
text-in/text-out translators:

- `opencode` (`opencode run`)
- `claude` (`claude -p`)
- `codex` (`codex exec`)
- `gemini` (`gemini -p`)

Kassie builds a prompt containing your schema (partition keys, clustering
keys, columns), pipes it to the first available CLI, and treats the output
as untrusted text. Kassie never sees or stores API keys — your existing
agent CLI authentication is used as-is.

The pipeline:

```
question → schema-context prompt → agent CLI → SELECT-only validation
        → dry-run against Cassandra → preview for your approval
```

Nothing is ever auto-run. The generated CQL is shown for review first.

## Four safety gates

1. **SELECT-only regex** — anything that is not a single SELECT is rejected
2. **Forbidden keyword scan** — INSERT/UPDATE/DELETE/DDL never reach the database
3. **Dry-run** — the query is executed with `LIMIT 1` against your cluster; syntax or schema errors trigger one automatic retry with the error fed back to the agent
4. **Human approval** — you press Enter/Use query only after reviewing the CQL

## Using it

**TUI:** press `Ctrl+N`, type your question, `Enter` to translate. `Ctrl+P`
cycles the provider (auto → opencode → claude → codex → gemini); `Ctrl+E`
opens an inline model editor (empty = the CLI's default model). The
generated CQL appears in an amber preview box — `Enter` to load it into the
grid, `Esc` to discard.

**Web:** click the **NL** toggle in the query editor, pick a provider from
the dropdown and optionally type a model (empty = the CLI's default), enter
your question, and click Translate. Review the generated CQL in the amber
panel, then "Use query" or "Discard".

Examples:

- "critical events for device 123e4567 yesterday"
- "top 10 customers by signup date in the east region"
- "count of failed logins per hour"

## Learning from examples

When you accept a translation, the question/CQL pair is saved as a few-shot
example (up to 50 per profile, in `~/.config/kassie/nl-examples.json`). Future
prompts include your accepted pairs, so translations improve on repeated
usage patterns for your schema.

## Configuration

Add an `ai` block to any profile in `~/.config/kassie/config.json`:

```json
{
  "name": "production",
  "hosts": ["cassandra.example.com"],
  "port": 9042,
  "ai": {
    "provider": "auto",
    "model": "",
    "timeout_seconds": 60,
    "privacy": "names+types"
  }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `provider` | `auto` | `auto`, `opencode`, `claude`, `codex`, or `gemini`. `auto` picks the first found in that order |
| `model` | *(CLI default)* | Passed as `--model`/`-m` to the agent CLI (e.g. `sonnet` for claude, `provider/model` for opencode). Empty = the CLI's own default model |
| `timeout_seconds` | `60` | Agent CLI timeout (max 600) |
| `privacy` | `names+types` | `names` sends column names only; `names+types` also includes CQL types |

Runtime overrides: the web panel has a provider dropdown and model input;
the TUI cycles providers with `Ctrl+P`. Per-request choices take precedence
over the profile config for that session.

## Privacy

- Only schema metadata for the selected keyspace is sent (table names,
  partition/clustering keys, column names, and optionally types — per the
  `privacy` setting). **Row data is never sent.**
- Prompts are capped at 50 tables; larger schemas are truncated.
- Communication happens between the kassie server process and the agent CLI
  on the same machine. What the agent CLI sends to its own provider depends
  on that tool's configuration — check its privacy policy if this matters
  for your environment.
- Accepted examples are stored locally in
  `~/.config/kassie/nl-examples.json` and are included in future prompts.

## Troubleshooting

- **"no ai provider available"** — none of the CLIs are in the server's
  `PATH`. Install one (e.g. `npm i -g opencode-ai`) and authenticate it.
- **Timeouts** — first runs can be slow (model warm-up); raise
  `timeout_seconds`.
- **Bad translations** — check that your question references real tables
  and include concrete values; accept good translations to build better
  few-shot examples.
