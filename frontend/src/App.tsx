import { Loader } from 'lucide-react';
import { BrowserRouter, Routes, Route, Navigate, useParams, useLocation } from 'react-router-dom';
import { AuthScreen } from './components/AuthScreen';
import { Login } from './components/Login';
import { Register } from './components/Register';
import { Dashboard } from './components/Dashboard';
import { Header } from './components/Header';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { UserProfile } from './components/UserProfile';
import { GardenDetail } from './components/GardenDetail';
import { useEffect, useState } from 'react';
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    apiService.getGarden(id)
      .then((res) => {
        setGarden(res.garden);
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
  return <GardenDetail garden={garden} onBack={() => window.history.back()} />;
}

function AppContent() {
  const location = useLocation();
  const hideHeader = location.pathname === '/login' || location.pathname === '/register';
  return (
    <>
      {!hideHeader && <Header />}
      <Routes>
        <Route
          path="/"
          element={
            <PrivateRoute>
              <Dashboard />
            </PrivateRoute>
          }
        />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/garden/:id"
          element={
            <PrivateRoute>
              <GardenDetailRoute />
            </PrivateRoute>
          }
        />
        <Route path="*" element={<div className="min-h-screen flex items-center justify-center text-2xl text-muted-foreground">404 Not Found</div>} />
      </Routes>
    </>
  );
}

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <BrowserRouter>
          <AppContent />
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;