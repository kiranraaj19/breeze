'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '@/lib/api';

interface User {
  token: string;
  id: string;
  email: string;
  userPairId: string;
  fullName: string;
}

interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check localStorage for saved user
    const savedUser = localStorage.getItem('breeze_user');
    if (savedUser) {
      setUser(JSON.parse(savedUser));
    }
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string): Promise<boolean> => {
    try {
      const response = await api.login(email, password);
      const userData: User = {
        token: response.token,
        id: response.id,
        email: response.email,
        userPairId: response.user_pair_id,
        fullName: response.full_name,
      };
      setUser(userData);
      localStorage.setItem('breeze_user', JSON.stringify(userData));
      return true;
    } catch (err) {
      console.error('Login failed:', err);
      return false;
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('breeze_user');
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
