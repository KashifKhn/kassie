import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Database, Loader2 } from 'lucide-react';
import { sessionApi } from '@/api/queries';
import { useAuthStore } from '@/stores/authStore';
import type { ProfileInfo } from '@/api/types';

export function LoginPage() {
  const navigate = useNavigate();
  const { setTokens, setProfile } = useAuthStore();
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
      navigate('/explorer');
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
      <div className="flex h-screen items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Loading profiles...
          </p>
        </div>
      </div>
    );
  }

  const profiles = profilesData?.profiles ?? [];

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 dark:bg-gray-900">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
            <Database className="h-8 w-8 text-blue-600 dark:text-blue-400" />
          </div>
          <h1 className="mt-6 text-3xl font-bold text-gray-900 dark:text-white">
            Kassie
          </h1>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Select a connection profile to continue
          </p>
        </div>

        <div className="mt-8 space-y-3">
          {profiles.length === 0 ? (
            <div className="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-gray-700 dark:bg-gray-800">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                No connection profiles found. Please configure profiles in your
                config file.
              </p>
            </div>
          ) : (
            profiles.map((profile) => (
              <button
                key={profile.name}
                onClick={() => handleLogin(profile)}
                disabled={loginMutation.isPending}
                className="w-full rounded-lg border border-gray-200 bg-white p-4 text-left transition-colors hover:border-blue-500 hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:hover:border-blue-500 dark:hover:bg-gray-700"
              >
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      {profile.name}
                    </h3>
                    <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                      {profile.hosts.join(', ')}
                    </p>
                    {profile.keyspace && (
                      <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
                        Default keyspace: {profile.keyspace}
                      </p>
                    )}
                  </div>
                  {loginMutation.isPending &&
                    selectedProfile === profile.name && (
                      <Loader2 className="h-5 w-5 animate-spin text-blue-600" />
                    )}
                </div>
              </button>
            ))
          )}
        </div>

        {loginMutation.isError && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20">
            <p className="text-sm text-red-800 dark:text-red-400">
              {loginMutation.error instanceof Error
                ? loginMutation.error.message
                : 'Failed to connect. Please try again.'}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
