import { useState } from 'react';
import { Search, X, Terminal } from 'lucide-react';

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
    <div 
      style={{ 
        borderBottom: '1px solid var(--border-primary)',
        background: 'var(--bg-secondary)',
      }}
    >
      <form onSubmit={handleSubmit} className="p-4">
        <div className="flex items-center gap-3">
          <div 
            className="flex-1 relative"
            style={{
              borderRadius: '8px',
              overflow: 'hidden',
            }}
          >
            <div 
              className="absolute left-3 top-1/2 -translate-y-1/2 flex items-center gap-2"
              style={{ color: 'var(--accent-primary)' }}
            >
              <Terminal className="w-4 h-4" />
              <span className="text-xs font-mono">WHERE</span>
            </div>
            <input
              type="text"
              value={whereClause}
              onChange={(e) => setWhereClause(e.target.value)}
              placeholder="id = 123 AND status = 'active'"
              className="w-full pl-24 pr-10 py-2.5 text-sm font-mono transition-all"
              style={{
                background: 'var(--bg-tertiary)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border-primary)',
                borderRadius: '8px',
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = 'var(--accent-primary)';
                e.currentTarget.style.boxShadow = '0 0 0 2px var(--accent-subtle)';
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = 'var(--border-primary)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            />
            {whereClause && (
              <button
                type="button"
                onClick={handleClear}
                className="absolute right-3 top-1/2 -translate-y-1/2 transition-all"
                style={{ color: 'var(--text-tertiary)' }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = 'var(--error)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = 'var(--text-tertiary)';
                }}
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <button
            type="submit"
            disabled={!whereClause.trim()}
            className="flex items-center gap-2 px-5 py-2.5 text-sm font-mono font-medium transition-all"
            style={{
              background: whereClause.trim() ? 'var(--accent-primary)' : 'var(--bg-tertiary)',
              color: whereClause.trim() ? 'var(--text-inverse)' : 'var(--text-tertiary)',
              border: '1px solid var(--border-primary)',
              borderRadius: '8px',
              cursor: whereClause.trim() ? 'pointer' : 'not-allowed',
              opacity: whereClause.trim() ? '1' : '0.5',
            }}
            onMouseEnter={(e) => {
              if (whereClause.trim()) {
                e.currentTarget.style.background = 'var(--accent-hover)';
                e.currentTarget.style.transform = 'translateY(-1px)';
                e.currentTarget.style.boxShadow = 'var(--shadow-glow)';
              }
            }}
            onMouseLeave={(e) => {
              if (whereClause.trim()) {
                e.currentTarget.style.background = 'var(--accent-primary)';
                e.currentTarget.style.transform = 'translateY(0)';
                e.currentTarget.style.boxShadow = 'none';
              }
            }}
          >
            <Search className="w-4 h-4" />
            Filter
          </button>
        </div>

        <div 
          className="mt-3 text-xs font-mono flex items-center gap-2"
          style={{ color: 'var(--text-tertiary)' }}
        >
          <span style={{ color: 'var(--accent-primary)' }}>▸</span>
          Enter a CQL WHERE clause to filter rows
          <span style={{ color: 'var(--text-tertiary)', opacity: '0.5' }}>|</span>
          Example: user_id = 123
        </div>
      </form>
    </div>
  );
}
