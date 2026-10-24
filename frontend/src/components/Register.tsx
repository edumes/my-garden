import { Eye, EyeOff, Leaf, Lock, Mail, User } from 'lucide-react';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { toast } from 'sonner';
import { useAuth } from '../contexts/AuthContext';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';

interface RegisterProps {
  onToggleMode?: () => void;
}

export function Register({ onToggleMode }: RegisterProps) {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    first_name: '',
    last_name: '',
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
      await register(formData);
      toast.success('Account created successfully!');
      navigate('/', { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed. Please try again.');
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
          <h2>Small seeds.<br />Big adventures.</h2>
          <p className="auth-intro">Build your own little world. Plant a seed, collect your harvest and unlock something new.</p>
          <div className="auth-plot" aria-hidden="true">
            {Array.from({ length: 9 }, (_, i) => <div key={i} className={i === 4 ? 'auth-plot-ripe' : ''}><Leaf /></div>)}
            <span className="auth-harvest">Ready for your first harvest?</span>
          </div>
          <div className="auth-starter"><span className="auth-coin">100</span><div><strong>Your adventure starts here</strong><p>Starting coins to buy your first seeds</p></div></div>
        </section>
        <div className="auth-form-panel">
          <div className="text-center mb-6 sm:mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 dark:bg-green-900 rounded-full mb-4">
              <Leaf className="w-8 h-8 text-green-600" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold text-foreground">Create Account</h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-2">Start your gardening journey</p>
          </div>

          {error && <p role="alert" className="auth-error">{error}</p>}
          <form aria-label="Create account" onSubmit={handleSubmit} className="space-y-2 sm:space-y-3">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
              <div>
                <Label htmlFor="register-first_name" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                  First Name (optional)
                </Label>
                <Input
                  type="text"
                  id="register-first_name"
                  autoComplete="given-name"
                  value={formData.first_name}
                  onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
                  className="w-full px-3 sm:px-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="John"
                />
              </div>
              <div>
                <Label htmlFor="register-last_name" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                  Last Name (optional)
                </Label>
                <Input
                  type="text"
                  id="register-last_name"
                  autoComplete="family-name"
                  value={formData.last_name}
                  onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
                  className="w-full px-3 sm:px-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Doe"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="register-username" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Username
              </Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input
                  type="text"
                  required
                  id="register-username"
                  autoComplete="username"
                  minLength={3}
                  maxLength={20}
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  className="w-full pl-10 pr-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Choose a username"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="register-email" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Email
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input
                  type="email"
                  required
                  id="register-email"
                  autoComplete="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full pl-10 pr-4 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="your@email.com"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="register-password" className="block text-sm font-medium text-foreground mb-1 sm:mb-2">
                Password
              </Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <Input
                  type={showPassword ? 'text' : 'password'}
                  required
                  minLength={6}
                  aria-describedby="password-hint"
                  id="register-password"
                  autoComplete="new-password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full pl-10 pr-12 py-2.5 sm:py-3 text-sm sm:text-base border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent transition-colors"
                  placeholder="Create a password"
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

            <p id="password-hint" className="text-xs text-muted-foreground">Use at least 6 characters for your password.</p>
            <Button
              type="submit"
              disabled={loading}
              className="w-full bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-semibold py-2.5 sm:py-3 px-4 text-sm sm:text-base rounded-lg transition-colors"
            >
              {loading ? 'Creating Account...' : 'Create Account'}
            </Button>
          </form>

          <div className="mt-2 sm:mt-3 text-center">
            <p className="text-sm sm:text-base text-muted-foreground">
              Already have an account?{' '}
              {onToggleMode ? (
                <button
                  onClick={onToggleMode}
                  className="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold"
                >
                  Sign in
                </button>
              ) : (
                <Link
                  to="/login"
                  className="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold"
                >
                  Sign in
                </Link>
              )}
            </p>
          </div>
        </div>
      </div>
    </main>
  );
}