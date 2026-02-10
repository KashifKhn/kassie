import { useState } from 'react';
import { Search, X } from 'lucide-react';

interface FilterBarProps {
  onFilter?: (whereClause: string) => void;
  onClear?: () => void;
}

export function FilterBar({ onFilter, onClear }: FilterBarProps) {
  const [whereClause, setWhereClause] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (whereClause.trim()) {
      onFilter?.(whereClause.trim());
    }
  };

  const handleClear = () => {
    setWhereClause('');
    onClear?.();
  };

  return (
    <div className="border-b border-border bg-background">
      <form onSubmit={handleSubmit} className="p-4">
        <div className="flex items-center gap-2">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              value={whereClause}
              onChange={(e) => setWhereClause(e.target.value)}
              placeholder="WHERE clause (e.g., id = 123 AND status = 'active')"
              className="w-full pl-10 pr-10 py-2 text-sm border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
            />
            {whereClause && (
              <button
                type="button"
                onClick={handleClear}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <button
            type="submit"
            disabled={!whereClause.trim()}
            className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Filter
          </button>
        </div>

        <div className="mt-2 text-xs text-muted-foreground">
          Enter a CQL WHERE clause to filter rows. Example: user_id = 123
        </div>
      </form>
    </div>
  );
}
