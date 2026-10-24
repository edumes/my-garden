import { Eye, EyeOff, Leaf, Lock, User } from 'lucide-react';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { toast } from 'sonner';
import { useAuth } from '../contexts/AuthContext';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';

interface LoginProps {
  onToggleMode?: () => void;
}

export function Login({ onToggleMode }: LoginProps) {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (loading) return;
    setError('');
    setLoading(true);

    try {
      await login(formData.username, formData.password);
      toast.success('Successfully signed in!');
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to sign in. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="auth-page">
      <div className="auth-shell">
        <section className="auth-hero" aria-label="Welcome to My Garden">
          <Link to="/login" className="auth-brand"><Leaf aria-hidden="true" /> My Garden</Link>
          <p className="auth-eyebrow">A LITTLE EVERY DAY. A LOT TO DISCOVER.</p>
          <h2>Your next<br />chapter grows here.</h2>
          <p className="auth-intro">Your plots, your plants, your pace. Come back and see what is ready to grow.</p>
          <div className="auth-plot" aria-hidden="true">
            {Array.from({ length: 9 }, (_, i) => <div key={i} className={i === 4 ? 'auth-plot-ripe' : ''}><Leaf /></div>)}
            <span className="auth-harvest">Ready for your first harvest?</span>
          </div>
          <div className="auth-starter"><span className="auth-coin">100</span><div><strong>Your adventure starts here</strong><p>Starting coins to buy your first seeds</p></div></div>
        </section>
        <div className="auth-form-panel">
          <div className="text-center mb-3 sm:mb-4">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 dark:bg-green-900 rounded-full mb-4">
              <Leaf className="w-8 h-8 text-green-600" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold text-foreground">Welcome Back</h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-2">Sign in to your garden account</p>
          </div>

          {error && <p role="alert" className="auth-error">{error}</p>}
          <form aria-label="Sign in" onSubmit={handleSubmit} className="space-y-4 sm:space-y-3">
            <div>
              <Label htmlFor="login-username" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Username
              </Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input
                  type="text"
                  required
                  id="login-username"
                  autoComplete="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  className="w-full pl-10 pr-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Enter your username"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="login-password" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Password
              </Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input
                  type={showPassword ? 'text' : 'password'}
                  required
                  id="login-password"
                  autoComplete="current-password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full pl-10 pr-12 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Enter your password"
                />
                <button
                  type="button"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  aria-pressed={showPassword}
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
            </div>

            <Button
              type="submit"
              disabled={loading}
              className="w-full bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-semibold py-2.5 sm:py-3 px-4 text-sm sm:text-base rounded-lg transition-colors"
            >
              {loading ? 'Signing In...' : 'Sign In'}
            </Button>
          </form>

          <div className="mt-4 sm:mt-6 text-center">
            <p className="text-sm sm:text-base text-muted-foreground">
              Don't have an account?{' '}
              {onToggleMode ? (
                <button
                  onClick={onToggleMode}
                  className="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold"
                >
                  Sign up
                </button>
              ) : (
                <Link
                  to="/register"
                  className="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold"
                >
                  Sign up
                </Link>
              )}
            </p>
          </div>
        </div>
      </div>
    </main>
  );
}