import { Moon, Sun } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';

export function ThemeToggle() {
  const { isDarkMode, toggleDarkMode } = useTheme();

  return (
    <button
      onClick={toggleDarkMode}
      className="relative p-2 rounded-lg bg-secondary hover:bg-secondary/80 transition-colors duration-200"
      title={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      <div className="relative w-5 h-5">
        <Sun
          className={`absolute inset-0 w-5 h-5 text-yellow-500 transition-all duration-300 ${isDarkMode ? 'opacity-0 rotate-90 scale-0' : 'opacity-100 rotate-0 scale-100'
            }`}
        />
        <Moon
          className={`absolute inset-0 w-5 h-5 text-blue-400 transition-all duration-300 ${isDarkMode ? 'opacity-100 rotate-0 scale-100' : 'opacity-0 -rotate-90 scale-0'
            }`}
        />
      </div>
    </button>
  );
}