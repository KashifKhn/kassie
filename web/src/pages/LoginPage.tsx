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
      className="relative flex min-h-screen items-center justify-center px-6 noise-bg overflow-hidden"
      style={{ background: 'var(--bg-primary)' }}
    >
      {/* Animated background elements - MUCH MORE VISIBLE */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/3 left-1/3 w-[600px] h-[600px] rounded-full blur-3xl opacity-30 animate-pulse" 
          style={{ 
            background: 'radial-gradient(circle, var(--accent-primary) 0%, transparent 60%)',
            animationDuration: '3s'
          }}
        />
        <div className="absolute bottom-1/3 right-1/3 w-[500px] h-[500px] rounded-full blur-3xl opacity-25 animate-pulse" 
          style={{ 
            background: 'radial-gradient(circle, var(--accent-hover) 0%, transparent 60%)',
            animationDuration: '4s',
            animationDelay: '1.5s'
          }}
        />
      </div>

      <div className="relative w-full max-w-2xl space-y-12 animate-scale-in">
        {/* Header - BIGGER AND BOLDER */}
        <div className="text-center space-y-8">
          <div className="inline-flex items-center justify-center relative group">
            <div 
              className="absolute inset-0 rounded-3xl blur-2xl opacity-60 group-hover:opacity-90 transition-opacity duration-500"
              style={{ background: 'linear-gradient(135deg, var(--accent-primary), var(--accent-hover))' }}
            />
            <div 
              className="relative flex items-center justify-center w-28 h-28 rounded-3xl"
              style={{ 
                background: 'var(--bg-elevated)',
                border: '3px solid var(--accent-primary)',
                boxShadow: '0 0 40px rgba(59, 130, 246, 0.3)'
              }}
            >
              <Database className="w-14 h-14" style={{ color: 'var(--accent-primary)' }} />
            </div>
          </div>
          
          <div className="space-y-4">
            <h1 
              className="font-mono text-7xl font-bold tracking-tight"
              style={{ color: 'var(--text-primary)' }}
            >
              KASSIE
            </h1>
            <div className="flex items-center justify-center gap-3">
              <div className="h-px w-12" style={{ background: 'var(--accent-primary)', opacity: 0.5 }} />
              <p 
                className="font-mono text-base tracking-widest uppercase"
                style={{ color: 'var(--accent-primary)' }}
              >
                Database Explorer
              </p>
              <div className="h-px w-12" style={{ background: 'var(--accent-primary)', opacity: 0.5 }} />
            </div>
            <p 
              className="text-base font-sans"
              style={{ color: 'var(--text-secondary)' }}
            >
              Select a connection profile to continue
            </p>
          </div>
        </div>

        {/* Profiles - MUCH LARGER CARDS */}
        <div className="space-y-5">
          {profiles.length === 0 ? (
            <div 
              className="rounded-2xl p-12 text-center backdrop-blur-sm"
              style={{ 
                background: 'var(--bg-secondary)',
                border: '2px solid var(--border-primary)'
              }}
            >
              <Shield className="w-16 h-16 mx-auto mb-6 opacity-50" style={{ color: 'var(--text-tertiary)' }} />
              <p className="font-mono text-base" style={{ color: 'var(--text-secondary)' }}>
                No connection profiles found.
              </p>
              <p className="text-sm mt-3" style={{ color: 'var(--text-tertiary)' }}>
                Configure profiles in ~/.config/kassie/config.json
              </p>
            </div>
          ) : (
            profiles.map((profile, index) => (
              <button
                key={profile.name}
                onClick={() => handleLogin(profile)}
                disabled={loginMutation.isPending}
                className="group w-full rounded-2xl text-left transition-all duration-300 disabled:cursor-not-allowed disabled:opacity-50 relative overflow-hidden animate-fade-in"
                style={{
                  background: 'var(--bg-elevated)',
                  border: '2px solid var(--border-primary)',
                  animationDelay: `${index * 150}ms`,
                  padding: '32px 40px'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = 'var(--accent-primary)';
                  e.currentTarget.style.boxShadow = '0 0 40px rgba(59, 130, 246, 0.4)';
                  e.currentTarget.style.transform = 'translateY(-4px)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'var(--border-primary)';
                  e.currentTarget.style.boxShadow = 'none';
                  e.currentTarget.style.transform = 'translateY(0)';
                }}
              >
                {/* Hover gradient effect - MORE VISIBLE */}
                <div 
                  className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
                  style={{
                    background: 'linear-gradient(135deg, transparent 0%, var(--accent-subtle) 100%)'
                  }}
                />
                
                <div className="relative flex items-center justify-between gap-8">
                  <div className="flex items-center gap-6">
                    {/* Icon - Much Bigger and Separated */}
                    <div 
                      className="flex-shrink-0 w-16 h-16 rounded-2xl flex items-center justify-center"
                      style={{
                        background: 'var(--accent-subtle)',
                        border: '2px solid var(--accent-primary)'
                      }}
                    >
                      <Server className="w-8 h-8" style={{ color: 'var(--accent-primary)' }} />
                    </div>
                    
                    {/* Content - Vertical Stack with Proper Spacing */}
                    <div className="flex-1 space-y-3">
                      <h3 
                        className="font-mono text-3xl font-bold tracking-wide"
                        style={{ color: 'var(--text-primary)' }}
                      >
                        {profile.name}
                      </h3>
                      
                      <div className="space-y-2">
                        <p 
                          className="font-mono text-base"
                          style={{ color: 'var(--text-primary)' }}
                        >
                          {profile.hosts.join(', ')}:{profile.port}
                        </p>
                        
                        {profile.keyspace && (
                          <p 
                            className="text-base font-mono flex items-center gap-2"
                            style={{ color: 'var(--text-secondary)' }}
                          >
                            <span style={{ color: 'var(--accent-primary)' }}>▸</span>
                            keyspace: {profile.keyspace}
                          </p>
                        )}
                        
                        {profile.sslEnabled && (
                          <div className="pt-1">
                            <span 
                              className="inline-flex items-center gap-2 text-sm font-mono px-3 py-1.5 rounded-lg"
                              style={{ 
                                background: 'var(--success)',
                                color: 'white',
                                boxShadow: '0 0 20px rgba(34, 197, 94, 0.3)'
                              }}
                            >
                              <Shield className="w-4 h-4" />
                              SSL Enabled
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                  
                  {loginMutation.isPending && selectedProfile === profile.name && (
                    <div className="flex-shrink-0">
                      <Loader2 
                        className="w-8 h-8 animate-spin" 
                        style={{ 
                          color: 'var(--accent-primary)',
                          filter: 'drop-shadow(0 0 10px var(--accent-primary))'
                        }} 
                      />
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
            className="rounded-2xl p-6 backdrop-blur-sm animate-slide-down"
            style={{ 
              background: 'rgba(239, 68, 68, 0.1)',
              border: '2px solid var(--error)',
              boxShadow: '0 0 30px rgba(239, 68, 68, 0.2)'
            }}
          >
            <p className="text-base font-mono font-bold" style={{ color: 'var(--error)' }}>
              ✗ CONNECTION ERROR: {loginMutation.error instanceof Error
                ? loginMutation.error.message
                : 'Connection failed'}
            </p>
          </div>
        )}

        {/* Footer hint */}
        <p 
          className="text-center text-sm font-mono"
          style={{ color: 'var(--text-tertiary)' }}
        >
          <span style={{ color: 'var(--accent-primary)' }}>&gt;_</span> Cassandra & ScyllaDB Explorer
        </p>
      </div>
    </div>
  );
}
