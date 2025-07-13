import { Coins, Leaf, LogOut, Settings, Trophy, User } from 'lucide-react';
import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { ThemeToggle } from './ThemeToggle';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';

export function Header() {
  const { user, logout } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  return (
    <header className="bg-card border-b border-border sticky top-0 z-50 transition-colors duration-200">
      <div className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center space-x-3">
            <div className="flex items-center space-x-2">
              <Leaf className="w-8 h-8 text-green-600" />
              <span className="text-lg sm:text-xl font-bold text-foreground">
                <span className="hidden sm:inline">Virtual Garden</span>
                <span className="sm:hidden">Garden</span>
              </span>
            </div>
          </div>

          {user && (
            <>
              <div className="hidden md:flex items-center space-x-4">
                <div className="flex items-center space-x-4 text-sm">
                  <Badge
                    variant="default"
                    className="flex items-center space-x-1 bg-yellow-100 dark:bg-yellow-900/20 px-3 py-1 rounded-full"
                  >
                    <Coins className="w-4 h-4 text-yellow-600" />
                    <span className="font-medium text-yellow-800 dark:text-yellow-200">{user.coins}</span>
                  </Badge>
                  <Badge
                    variant="default"
                    className="flex items-center space-x-1 bg-blue-100 dark:bg-blue-900/20 px-3 py-1 rounded-full">
                    <Trophy className="w-4 h-4 text-blue-600" />
                    <span className="font-medium text-blue-800 dark:text-blue-200">Level {user.level}</span>
                  </Badge>
                </div>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="outline"
                      className="flex items-center space-x-2 text-muted-foreground hover:text-foreground p-2 rounded-lg hover:bg-accent transition-colors"
                    >
                      <User className="w-5 h-5" />
                      <span className="font-medium">{user.username}</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-48">
                    <DropdownMenuLabel>My Account</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <Settings className="w-4 h-4 mr-2" />
                      <span>Settings</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={logout} className="text-destructive focus:text-destructive">
                      <LogOut className="w-4 h-4 mr-2" />
                      <span>Logout</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <ThemeToggle />
              </div>

              <div className="md:hidden flex items-center space-x-2">
                <ThemeToggle />
                <Button
                  onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                  className="p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-accent transition-colors"
                >
                  <User className="w-5 h-5" />
                </Button>
              </div>
            </>
          )}
          {!user && <ThemeToggle />}
        </div>

        {/* Mobile menu */}
        {user && isMobileMenuOpen && (
          <div className="md:hidden border-t border-border py-4">
            <div className="flex flex-col space-y-4">
              <div className="flex items-center justify-between">
                <span className="font-medium text-foreground">{user.username}</span>
                <div className="flex items-center space-x-3 text-sm">
                  <div className="flex items-center space-x-1 bg-yellow-100 dark:bg-yellow-900/20 px-2 py-1 rounded-full">
                    <Coins className="w-3 h-3 text-yellow-600" />
                    <span className="font-medium text-yellow-800 dark:text-yellow-200">{user.coins}</span>
                  </div>
                  <div className="flex items-center space-x-1 bg-blue-100 dark:bg-blue-900/20 px-2 py-1 rounded-full">
                    <Trophy className="w-3 h-3 text-blue-600" />
                    <span className="font-medium text-blue-800 dark:text-blue-200">Level {user.level}</span>
                  </div>
                </div>
              </div>
              <div className="flex flex-col space-y-2">
                <Button
                  variant="secondary"
                  className="flex items-center space-x-2 w-full px-3 py-2 text-sm text-muted-foreground hover:bg-accent rounded-lg text-left">
                  <Settings className="w-4 h-4" />
                  <span>Settings</span>
                </Button>
                <Button
                  onClick={logout}
                  variant="ghost"
                  className="flex items-center space-x-2 w-full px-3 py-2 text-sm text-destructive hover:bg-destructive/10 rounded-lg text-left"
                >
                  <LogOut className="w-4 h-4" />
                  <span>Logout</span>
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </header>
  );
}