import { Loader } from 'lucide-react';
import { useEffect, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation, useParams } from 'react-router-dom';
import { TooltipProvider } from './components/animate-ui/components/tooltip';
import { Dashboard } from './components/Dashboard';
import { GardenDetail } from './components/GardenDetail';
import { Header } from './components/Header';
import JoinGardenPage from './components/JoinGardenPage';
import { Login } from './components/Login';
import { Register } from './components/Register';
import { Store } from './components/Store';
import { AuditLogs } from './components/AuditLogs';
import { Marketplace } from './components/Marketplace';
import { Settings } from './components/Settings';
import { Sidebar } from './components/admin-panel/sidebar';
import { Toaster } from './components/ui/sonner';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { apiService } from './services/api';

function PrivateRoute({ children }: { children: JSX.Element }) {
  const { user, loading } = useAuth();
  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center transition-colors duration-200">
        <div className="text-center">
          <Loader className="w-8 h-8 text-green-600 animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Loading your garden...</p>
        </div>
      </div>
    );
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

function GardenDetailRoute() {
  const { id } = useParams();
  const [garden, setGarden] = useState<any>(null);
  const [permission, setPermission] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showStore, setShowStore] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    apiService.getGarden(id)
      .then((res) => {
        setGarden(res.garden);
        setPermission(res.permission);
        setError('');
      })
      .catch(() => {
        setError('Garden not found');
        setGarden(null);
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader className="w-8 h-8 text-green-600 animate-spin mx-auto mb-4" />
      </div>
    );
  }
  if (error || !garden) {
    return <div className="min-h-screen flex items-center justify-center text-2xl text-muted-foreground">Garden not found</div>;
  }
  return (
    <>
      <GardenDetail
        garden={garden}
        permission={permission}
        isShared={!!permission && permission !== 'manage'}
        onBack={() => window.history.back()}
        onOpenStore={() => setShowStore(true)}
      />
      <Store open={showStore} onClose={() => setShowStore(false)} />
    </>
  );
}

function AppContent() {
  const location = useLocation();
  const hideHeader = location.pathname === '/login' || location.pathname === '/register';
  const [showStore, setShowStore] = useState(false);

  return (
    <>
      {!hideHeader && <Header onOpenStore={() => setShowStore(true)} />}
      {!hideHeader && <Sidebar />}
      <div className={!hideHeader ? "lg:ml-72 transition-all duration-300 pt-16 lg:pt-0" : ""}>
        <Routes>
          <Route
            path="/"
            element={
              <PrivateRoute>
                <Dashboard />
              </PrivateRoute>
            }
          />
          <Route
            path="/garden"
            element={
              <PrivateRoute>
                <Dashboard />
              </PrivateRoute>
            }
          />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route
            path="/join-garden/:token"
            element={
              <PrivateRoute>
                <JoinGardenPage />
              </PrivateRoute>
            }
          />
          <Route
            path="/garden/:id"
            element={
              <PrivateRoute>
                <GardenDetailRoute />
              </PrivateRoute>
            }
          />
          <Route
            path="/marketplace"
            element={
              <PrivateRoute>
                <Marketplace />
              </PrivateRoute>
            }
          />
          <Route
            path="/audit"
            element={
              <PrivateRoute>
                <AuditLogs />
              </PrivateRoute>
            }
          />
          <Route
            path="/settings"
            element={
              <PrivateRoute>
                <Settings />
              </PrivateRoute>
            }
          />
          <Route path="*" element={<div className="min-h-screen flex items-center justify-center text-2xl text-muted-foreground">404 Not Found</div>} />
        </Routes>
      </div>

      <Store open={showStore} onClose={() => setShowStore(false)} />
    </>
  );
}

function App() {
  return (
    <ThemeProvider>
      <TooltipProvider openDelay={1} closeDelay={1}>
        <AuthProvider>
          <BrowserRouter>
            <AppContent />
            <Toaster richColors position='bottom-center' closeButton duration={1800} />
          </BrowserRouter>
        </AuthProvider>
      </TooltipProvider>
    </ThemeProvider>
  );
}

export default App;