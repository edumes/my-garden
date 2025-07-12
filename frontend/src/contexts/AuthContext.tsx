import { createContext, ReactNode, useContext, useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { User } from '../types/api';

interface AuthContextType {
  user: User | null;
  login: (username: string, password: string) => Promise<void>;
  register: (userData: {
    username: string;
    email: string;
    password: string;
    first_name?: string;
    last_name?: string;
  }) => Promise<void>;
  logout: () => Promise<void>;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      loadUserProfile();
    } else {
      setLoading(false);
    }
  }, []);

  const loadUserProfile = async () => {
    try {
      const response = await apiService.getUserProfile();
      setUser(response.user);
    } catch (error) {
      localStorage.removeItem('auth_token');
    } finally {
      setLoading(false);
    }
  };

  const login = async (username: string, password: string) => {
    const response = await apiService.login({ username, password });
    setUser(response.user);
  };

  const register = async (userData: {
    username: string;
    email: string;
    password: string;
    first_name?: string;
    last_name?: string;
  }) => {
    const response = await apiService.register(userData);
    setUser(response.user);
  };

  const logout = async () => {
    await apiService.logout();
    setUser(null);
  };

  const value = {
    user,
    login,
    register,
    logout,
    loading,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}