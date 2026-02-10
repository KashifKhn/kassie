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
        background: 'var(--bg-primary)',
        padding: '14px 20px'
      }}
    >
      <form onSubmit={handleSubmit}>
        <div className="flex items-stretch gap-3">
          <div className="flex-1 relative">
            {/* WHERE Label - Refined */}
            <div 
              className="absolute left-4 top-1/2 -translate-y-1/2 flex items-center gap-2 pointer-events-none"
              style={{ 
                color: 'var(--accent-primary)',
                zIndex: 10,
                opacity: 0.9
              }}
            >
              <Terminal className="w-4 h-4" />
              <span className="text-sm font-mono font-medium tracking-wide">WHERE</span>
            </div>
            
            {/* Input Field */}
            <input
              type="text"
              value={whereClause}
              onChange={(e) => setWhereClause(e.target.value)}
              placeholder="id = 123 AND status = 'active'"
              className="w-full font-mono transition-all placeholder:text-[var(--text-tertiary)] placeholder:opacity-60"
              style={{
                paddingLeft: '110px',
                paddingRight: whereClause ? '50px' : '20px',
                paddingTop: '12px',
                paddingBottom: '12px',
                fontSize: '14px',
                background: 'var(--bg-elevated)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border-primary)',
                borderRadius: '8px',
                outline: 'none',
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = 'var(--accent-primary)';
                e.currentTarget.style.boxShadow = '0 0 0 3px var(--accent-subtle)';
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = 'var(--border-primary)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            />
            
            {/* Clear Button */}
            {whereClause && (
              <button
                type="button"
                onClick={handleClear}
                className="absolute right-4 top-1/2 -translate-y-1/2 transition-all rounded-full p-1"
                style={{ 
                  color: 'var(--text-tertiary)',
                  background: 'transparent'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = 'var(--error)';
                  e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = 'var(--text-tertiary)';
                  e.currentTarget.style.background = 'transparent';
                }}
              >
                <X className="w-5 h-5" />
              </button>
            )}
          </div>
          
          {/* Filter Button */}
          <button
            type="submit"
            disabled={!whereClause.trim()}
            className="flex items-center gap-2 px-6 font-mono font-medium text-sm transition-all"
            style={{
              background: whereClause.trim() ? 'var(--accent-primary)' : 'var(--bg-tertiary)',
              color: whereClause.trim() ? 'var(--text-inverse)' : 'var(--text-tertiary)',
              border: '1px solid ' + (whereClause.trim() ? 'var(--accent-primary)' : 'var(--border-primary)'),
              borderRadius: '8px',
              cursor: whereClause.trim() ? 'pointer' : 'not-allowed',
              opacity: whereClause.trim() ? '1' : '0.6',
            }}
            onMouseEnter={(e) => {
              if (whereClause.trim()) {
                e.currentTarget.style.background = 'var(--accent-hover)';
                e.currentTarget.style.transform = 'translateY(-1px)';
                e.currentTarget.style.boxShadow = '0 4px 20px rgba(59, 130, 246, 0.3)';
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
            <span>Filter</span>
          </button>
        </div>

        {/* Helper Text */}
        <div 
          className="mt-3 text-xs font-mono flex items-center gap-2"
          style={{ color: 'var(--text-tertiary)' }}
        >
          <span style={{ color: 'var(--accent-primary)', fontSize: '14px' }}>▸</span>
          <span>Enter a CQL WHERE clause to filter rows</span>
          <span style={{ color: 'var(--border-primary)' }}>•</span>
          <span style={{ opacity: 0.7 }}>Example: user_id = 123</span>
        </div>
      </form>
    </div>
  );
}
