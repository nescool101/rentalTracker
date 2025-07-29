import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { notifications } from '@mantine/notifications';
import authService, { User } from '../services/authService';
import { IconCheck, IconX } from '@tabler/icons-react';
import { useNavigate } from 'react-router-dom';

// Enable this for debugging
const DEBUG = false; // Disable debug logs in production

const debugLog = (...args: any[]) => {
  if (DEBUG) {
    console.log('[AuthContext]', ...args);
  }
};

// Make sure these match the keys in authService.ts
const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  register: (name: string, email: string, password: string) => Promise<boolean>;
  logout: () => void;
  redirectToUserHome: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // Check if user is already logged in
    debugLog('Checking for existing user on initial load');
    const currentUser = authService.getCurrentUser();
    if (currentUser) {
      debugLog('Found existing user:', currentUser);
      setUser(currentUser);
    } else {
      debugLog('No existing user found');
    }
    setIsLoading(false);
  }, []);

  // Redirect based on user role
  const redirectToUserHome = () => {
    debugLog('redirectToUserHome called, user:', user);
    if (!user) {
      debugLog('Cannot redirect, no user is authenticated');
      return;
    }
    
    // Consistent redirect target for this function as well
    debugLog(`Redirecting ${user.role} user via redirectToUserHome to /dashboard`);
    navigate('/dashboard');
  };

  const login = async (email: string, password: string): Promise<boolean> => {
    debugLog('Login initiated for email:', email);
    setIsLoading(true);
    
    try {
      debugLog('Calling authService.login');
      const response = await authService.login(email, password);
      debugLog('Login response:', response);
      
      if (response.success && response.user) {
        debugLog('Login successful, setting user state');
        setUser(response.user);
        
        // Store user data and token - use the same keys as in authService
        localStorage.setItem(USER_KEY, JSON.stringify(response.user));
        if (response.token) {
          localStorage.setItem(TOKEN_KEY, response.token);
          debugLog('Token saved to localStorage');
        }
        
        notifications.show({
          title: 'Inicio de sesión exitoso',
          message: 'Bienvenido al sistema',
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        
        // Check if user is a new user (status === 'newuser')
        if (response.user.status === 'newuser') {
          debugLog('New user detected, redirecting to onboarding');
          navigate('/onboarding');
        } else if (response.user.status === 'activenopaid') {
          // For users with pending activation, still redirect to dashboard
          // The dashboard component will display the pending activation notice
          debugLog('User with pending activation, redirecting to dashboard');
          navigate('/dashboard');
        } else {
          // Auto-redirect based on role for existing users
          debugLog(`Redirecting ${response.user.role} user to /dashboard`);
          navigate('/dashboard');
        }
        
        return true;
      } else {
        debugLog('Login failed:', response.message);
        
        // Store the last error message for retrieval by other components
        const lastError = response.message || 'Credenciales inválidas';
        localStorage.setItem('last_auth_error', lastError);
        
        notifications.show({
          title: 'Error de autenticación',
          message: lastError,
          color: 'red',
          icon: <IconX size={16} />,
        });
        return false;
      }
    } catch (error) {
      debugLog('Login error:', error);
      console.error('Login error:', error);
      notifications.show({
        title: 'Error de conexión',
        message: 'No se pudo conectar con el servidor. Intente de nuevo más tarde.',
        color: 'red',
        icon: <IconX size={16} />,
      });
      return false;
    } finally {
      debugLog('Login process completed, setting isLoading to false');
      setIsLoading(false);
    }
  };

  const register = async (name: string, email: string, password: string): Promise<boolean> => {
    debugLog('Register initiated for email:', email);
    setIsLoading(true);
    
    try {
      const response = await authService.register(name, email, password);
      debugLog('Register response:', response);
      
      if (response.success) {
        debugLog('Registration successful');
        notifications.show({
          title: 'Registro exitoso',
          message: 'Ahora puedes iniciar sesión',
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        return true;
      } else {
        debugLog('Registration failed:', response.message);
        notifications.show({
          title: 'Error de registro',
          message: response.message || 'No se pudo completar el registro',
          color: 'red',
          icon: <IconX size={16} />,
        });
        return false;
      }
    } catch (error) {
      debugLog('Registration error:', error);
      console.error('Registration error:', error);
      notifications.show({
        title: 'Error de conexión',
        message: 'No se pudo conectar con el servidor',
        color: 'red',
        icon: <IconX size={16} />,
      });
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    debugLog('Logout initiated');
    setUser(null);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem(TOKEN_KEY);
    debugLog('Navigating to /login');
    navigate('/login');
    notifications.show({
      title: 'Sesión cerrada',
      message: 'Has cerrado sesión correctamente',
      color: 'blue',
    });
  };

  const contextValue = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
    redirectToUserHome,
  };

  debugLog('Rendering AuthProvider with context:', contextValue);

  return (
    <AuthContext.Provider value={contextValue}>
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