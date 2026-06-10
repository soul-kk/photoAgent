import React from 'react';
import * as SecureStore from 'expo-secure-store';

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://8.139.5.99:31335';
const KEY_TOKEN = 'auth_token';
const KEY_USER = 'auth_user';

// ─── Types ────────────────────────────────────────────────────────────────────

export type User = {
  id: number;
  username: string;
  email: string;
  role: string;
};

type AuthContextValue = {
  token: string | null;
  user: User | null;
  isLoaded: boolean;
  login: (account: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
};

// ─── Context ──────────────────────────────────────────────────────────────────

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = React.useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = React.useState<string | null>(null);
  const [user, setUser] = React.useState<User | null>(null);
  const [isLoaded, setIsLoaded] = React.useState(false);

  // Load persisted auth on mount
  React.useEffect(() => {
    async function load() {
      try {
        const [storedToken, storedUser] = await Promise.all([
          SecureStore.getItemAsync(KEY_TOKEN),
          SecureStore.getItemAsync(KEY_USER),
        ]);
        if (storedToken) setToken(storedToken);
        if (storedUser) setUser(JSON.parse(storedUser) as User);
      } catch {
        // corrupt storage — treat as logged out
      } finally {
        setIsLoaded(true);
      }
    }
    load();
  }, []);

  async function login(account: string, password: string) {
    const res = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ account, password }),
    });
    const json = await res.json();
    if (!res.ok || json.code !== 200) {
      throw new Error(json.message ?? '登录失败，请检查用户名或密码');
    }
    const { access_token, user: userData } = json.data as {
      access_token: string;
      user: User;
    };
    await Promise.all([
      SecureStore.setItemAsync(KEY_TOKEN, access_token),
      SecureStore.setItemAsync(KEY_USER, JSON.stringify(userData)),
    ]);
    setToken(access_token);
    setUser(userData);
  }

  async function register(username: string, email: string, password: string) {
    const res = await fetch(`${API_BASE}/api/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password }),
    });
    const json = await res.json();
    if (!res.ok || json.code !== 200) {
      throw new Error(json.message ?? '注册失败，请稍后重试');
    }
    // Auto-login after register
    await login(username, password);
  }

  async function logout() {
    await Promise.all([
      SecureStore.deleteItemAsync(KEY_TOKEN),
      SecureStore.deleteItemAsync(KEY_USER),
    ]);
    setToken(null);
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ token, user, isLoaded, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
