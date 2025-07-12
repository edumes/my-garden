import { Loader } from 'lucide-react';
import { AuthScreen } from './components/AuthScreen';
import { Dashboard } from './components/Dashboard';
import { Header } from './components/Header';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';

function AppContent() {
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
    return <AuthScreen />;
  }

  return (
    <div className="min-h-screen bg-background transition-colors duration-200">
      <Header />
      <Dashboard />
    </div>
  );
}

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;