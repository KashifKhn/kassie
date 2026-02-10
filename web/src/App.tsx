import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { LoginPage } from '@/pages/LoginPage';
import { ExplorerPage } from '@/pages/ExplorerPage';
import { NotFoundPage } from '@/pages/NotFoundPage';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ToastContainer } from '@/components/toast/ToastContainer';
import { ThemeProvider } from '@/components/theme/ThemeProvider';

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/explorer"
            element={
              <ProtectedRoute>
                <ExplorerPage />
              </ProtectedRoute>
            }
          />
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
        <ToastContainer />
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
