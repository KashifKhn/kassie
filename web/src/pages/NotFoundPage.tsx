import { Link } from 'react-router-dom';
import { Home } from 'lucide-react';

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 dark:bg-gray-900">
      <div className="text-center">
        <h1 className="text-9xl font-bold text-gray-200 dark:text-gray-800">
          404
        </h1>
        <h2 className="mt-4 text-3xl font-bold text-gray-900 dark:text-white">
          Page not found
        </h2>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          The page you are looking for does not exist.
        </p>
        <Link
          to="/"
          className="mt-8 inline-flex items-center gap-2 rounded-lg bg-blue-600 px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          <Home className="h-4 w-4" />
          Go back home
        </Link>
      </div>
    </div>
  );
}
