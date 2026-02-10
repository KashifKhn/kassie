import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Database, Loader2, Server, Shield } from 'lucide-react';
import { sessionApi } from '@/api/queries';
import { useAuthStore } from '@/stores/authStore';
import { useToastStore } from '@/stores/toastStore';
import type { ProfileInfo } from '@/api/types';

export function LoginPage() {
  const navigate = useNavigate();
  const { setTokens, setProfile } = useAuthStore();
  const { success, error } = useToastStore();
  const [selectedProfile, setSelectedProfile] = useState<string>('');

  const { data: profilesData, isLoading: loadingProfiles } = useQuery({
    queryKey: ['profiles'],
    queryFn: sessionApi.getProfiles,
  });

  const loginMutation = useMutation({
    mutationFn: sessionApi.login,
    onSuccess: (data) => {
      setTokens(data.accessToken, data.refreshToken, data.expiresAt);
      setProfile(data.profile);
      success(`Connected to ${data.profile.name}`);
      
      queueMicrotask(() => {
        navigate('/explorer');
      });
    },
    onError: (err) => {
      error(err instanceof Error ? err.message : 'Failed to connect');
    },
  });

  const handleLogin = (profile: ProfileInfo) => {
    setSelectedProfile(profile.name);
    loginMutation.mutate({
      profile: profile.name,
    });
  };

  if (loadingProfiles) {
    return (
      <div className="flex h-screen items-center justify-center" style={{ background: 'var(--bg-primary)' }}>
        <div className="flex flex-col items-center gap-6 animate-fade-in">
          <div className="relative">
            <div className="absolute inset-0 animate-ping" style={{ 
              background: 'radial-gradient(circle, var(--accent-primary) 0%, transparent 70%)',
              opacity: 0.2 
            }}/>
            <Loader2 className="h-10 w-10 animate-spin relative" style={{ color: 'var(--accent-primary)' }} />
          </div>
          <p className="font-mono text-sm tracking-wide" style={{ color: 'var(--text-secondary)' }}>
            INITIALIZING_CONNECTION...
          </p>
        </div>
      </div>
    );
  }

  const profiles = profilesData?.profiles ?? [];

  return (
    <div 
      className="relative flex min-h-screen items-center justify-center px-4 noise-bg overflow-hidden"
      style={{ background: 'var(--bg-primary)' }}
    >
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 rounded-full blur-3xl opacity-20 animate-pulse" 
          style={{ 
            background: 'radial-gradient(circle, var(--accent-primary) 0%, transparent 70%)',
            animationDuration: '4s'
          }}
        />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 rounded-full blur-3xl opacity-15 animate-pulse" 
          style={{ 
            background: 'radial-gradient(circle, var(--accent-primary) 0%, transparent 70%)',
            animationDuration: '6s',
            animationDelay: '2s'
          }}
        />
      </div>

      <div className="relative w-full max-w-lg space-y-10 animate-scale-in">
        {/* Header */}
        <div className="text-center space-y-6">
          <div className="inline-flex items-center justify-center relative group">
            <div 
              className="absolute inset-0 rounded-2xl blur-xl opacity-50 group-hover:opacity-75 transition-opacity duration-500"
              style={{ background: 'linear-gradient(135deg, var(--accent-primary), var(--accent-hover))' }}
            />
            <div 
              className="relative flex items-center justify-center w-20 h-20 rounded-2xl"
              style={{ 
                background: 'var(--bg-elevated)',
                border: '2px solid var(--border-primary)'
              }}
            >
              <Database className="w-10 h-10" style={{ color: 'var(--accent-primary)' }} />
            </div>
          </div>
          
          <div className="space-y-3">
            <h1 
              className="font-mono text-5xl font-bold tracking-tight"
              style={{ color: 'var(--text-primary)' }}
            >
              KASSIE
            </h1>
            <div className="flex items-center justify-center gap-2">
              <div className="h-px w-8" style={{ background: 'var(--border-secondary)' }} />
              <p 
                className="font-mono text-sm tracking-wider uppercase"
                style={{ color: 'var(--text-tertiary)' }}
              >
                Database Explorer
              </p>
              <div className="h-px w-8" style={{ background: 'var(--border-secondary)' }} />
            </div>
            <p 
              className="text-sm"
              style={{ color: 'var(--text-secondary)' }}
            >
              Select a connection profile to continue
            </p>
          </div>
        </div>

        {/* Profiles */}
        <div className="space-y-4">
          {profiles.length === 0 ? (
            <div 
              className="rounded-xl p-8 text-center backdrop-blur-sm"
              style={{ 
                background: 'var(--bg-secondary)',
                border: '1px solid var(--border-primary)'
              }}
            >
              <Shield className="w-12 h-12 mx-auto mb-4 opacity-50" style={{ color: 'var(--text-tertiary)' }} />
              <p className="font-mono text-sm" style={{ color: 'var(--text-secondary)' }}>
                No connection profiles found.
              </p>
              <p className="text-xs mt-2" style={{ color: 'var(--text-tertiary)' }}>
                Configure profiles in ~/.config/kassie/config.json
              </p>
            </div>
          ) : (
            profiles.map((profile, index) => (
              <button
                key={profile.name}
                onClick={() => handleLogin(profile)}
                disabled={loginMutation.isPending}
                className="group w-full rounded-xl p-6 text-left transition-all duration-300 disabled:cursor-not-allowed disabled:opacity-50 relative overflow-hidden"
                style={{
                  background: 'var(--bg-elevated)',
                  border: '1px solid var(--border-primary)',
                  animationDelay: `${index * 100}ms`
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = 'var(--accent-primary)';
                  e.currentTarget.style.boxShadow = 'var(--shadow-glow)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'var(--border-primary)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                {/* Hover gradient effect */}
                <div 
                  className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
                  style={{
                    background: 'linear-gradient(135deg, transparent 0%, var(--accent-subtle) 100%)'
                  }}
                />
                
                <div className="relative flex items-start justify-between gap-4">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-3">
                      <Server className="w-5 h-5 flex-shrink-0" style={{ color: 'var(--accent-primary)' }} />
                      <h3 
                        className="font-mono text-lg font-semibold tracking-wide"
                        style={{ color: 'var(--text-primary)' }}
                      >
                        {profile.name}
                      </h3>
                    </div>
                    
                    <div className="space-y-1 pl-8">
                      <p 
                        className="font-mono text-sm"
                        style={{ color: 'var(--text-secondary)' }}
                      >
                        {profile.hosts.join(', ')}:{profile.port}
                      </p>
                      
                      {profile.keyspace && (
                        <p 
                          className="text-xs font-mono"
                          style={{ color: 'var(--text-tertiary)' }}
                        >
                          keyspace: {profile.keyspace}
                        </p>
                      )}
                      
                      {profile.sslEnabled && (
                        <span 
                          className="inline-flex items-center gap-1 text-xs font-mono px-2 py-0.5 rounded"
                          style={{ 
                            background: 'var(--success)',
                            color: 'white',
                            opacity: 0.9
                          }}
                        >
                          <Shield className="w-3 h-3" />
                          SSL
                        </span>
                      )}
                    </div>
                  </div>
                  
                  {loginMutation.isPending && selectedProfile === profile.name && (
                    <div className="flex-shrink-0">
                      <Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--accent-primary)' }} />
                    </div>
                  )}
                </div>
              </button>
            ))
          )}
        </div>

        {/* Error Message */}
        {loginMutation.isError && (
          <div 
            className="rounded-xl p-4 backdrop-blur-sm animate-slide-down"
            style={{ 
              background: 'rgb(239 68 68 / 0.1)',
              border: '1px solid var(--error)'
            }}
          >
            <p className="text-sm font-mono" style={{ color: 'var(--error)' }}>
              ERROR: {loginMutation.error instanceof Error
                ? loginMutation.error.message
                : 'Connection failed'}
            </p>
          </div>
        )}

        {/* Footer hint */}
        <p 
          className="text-center text-xs font-mono opacity-50"
          style={{ color: 'var(--text-tertiary)' }}
        >
          &gt;_ Cassandra & ScyllaDB Explorer
        </p>
      </div>
    </div>
  );
}
