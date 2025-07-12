import { useState } from 'react';
import { Login } from './Login';
import { Register } from './Register';

export function AuthScreen() {
  const [isLogin, setIsLogin] = useState(true);

  return isLogin ? (
    <Login onToggleMode={() => setIsLogin(false)} />
  ) : (
    <Register onToggleMode={() => setIsLogin(true)} />
  );
}