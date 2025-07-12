import { Eye, EyeOff, Leaf, Lock, User } from 'lucide-react';
import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

interface LoginProps {
  onToggleMode: () => void;
}

export function Login({ onToggleMode }: LoginProps) {
  const { login } = useAuth();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await login(formData.username, formData.password);
    } catch (err) {
      setError('Invalid username or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-3 sm:p-4 transition-colors duration-200">
      <div className="max-w-md w-full">
        <div className="bg-card rounded-2xl shadow-xl p-6 sm:p-8 transition-colors duration-200">
          <div className="text-center mb-6 sm:mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 dark:bg-green-900 rounded-full mb-4">
              <Leaf className="w-8 h-8 text-green-600" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold text-foreground">Welcome Back</h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-2">Sign in to your garden account</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4 sm:space-y-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Username
              </label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  type="text"
                  required
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  className="w-full pl-10 pr-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Enter your username"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  required
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full pl-10 pr-12 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Enter your password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            {error && (
              <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-2 sm:p-3">
                <p className="text-destructive text-sm">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-semibold py-2.5 sm:py-3 px-4 text-sm sm:text-base rounded-lg transition-colors"
            >
              {loading ? 'Signing In...' : 'Sign In'}
            </button>
          </form>

          <div className="mt-4 sm:mt-6 text-center">
            <p className="text-sm sm:text-base text-muted-foreground">
              Don't have an account?{' '}
              <button
                onClick={onToggleMode}
                className="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold"
              >
                Sign up
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}