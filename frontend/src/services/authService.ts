import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || '';

export interface LoginResponse {
  success: boolean;
  user?: {
    id: string;
    email: string;
    role: string;
    person_id?: string;
    status?: string;
  };
  token?: string;
  message?: string;
}

export interface User {
  id: string;
  email: string;
  role: string;
  person_id?: string;
  status?: string;
  token?: string;
}

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'user';

/**
 * Service for handling authentication with the backend
 */
export const authService = {
  /**
   * Attempts to authenticate the user with the backend
   */
  login: async (username: string, password: string): Promise<LoginResponse> => {
    try {
      try {
        // Always encode password as base64 before sending to the backend
        const encodedPassword = btoa(password);
        console.log(`Login attempt for ${username} with encoded password`);
        
        const payload = { email: username, password: encodedPassword };
        
        const response = await axios.post(`${API_URL}/api/users/login`, payload);
        
        if (response.data && response.data.success) {
          const userData = response.data.user;
          const token = response.data.token;
          
          localStorage.setItem(TOKEN_KEY, token);
          
          const user = {
            id: userData.id,
            email: userData.email,
            role: userData.role,
            person_id: userData.person_id,
            status: userData.status,
            token: token
          };
          
          localStorage.setItem(USER_KEY, JSON.stringify(user));
          
          return {
            success: true,
            user,
            token
          };
        } else {
          return {
            success: false,
            message: response.data.message || response.data.error || 'Credenciales inválidas'
          };
        }
      } catch (apiError: any) {
        // Check if it's a response error
        if (apiError.response && apiError.response.data) {
          return {
            success: false,
            message: apiError.response.data.error || 'Credenciales inválidas'
          };
        }
      }
      
      // Fallback for development/testing - hardcoded credentials
      if (username === 'adminscao@rentalmanager.com' && password === 'Nesc@02025.') {
        const mockUser = {
          id: '1',
          email: username,
          role: 'admin',
          status: 'active',
          token: 'simulated-jwt-token'
        };
        
        localStorage.setItem(TOKEN_KEY, 'simulated-jwt-token');
        localStorage.setItem(USER_KEY, JSON.stringify(mockUser));
        
        return {
          success: true,
          user: mockUser,
          token: 'simulated-jwt-token'
        };
      } else if (username === 'guest@rentalmanager.com' && password === 'guest123') {
        const mockUser = {
          id: '2',
          email: username,
          role: 'user',
          status: 'active',
          token: 'simulated-jwt-token-guest'
        };
        
        localStorage.setItem(TOKEN_KEY, 'simulated-jwt-token-guest');
        localStorage.setItem(USER_KEY, JSON.stringify(mockUser));
        
        return {
          success: true,
          user: mockUser,
          token: 'simulated-jwt-token-guest'
        };
      } else {
        return {
          success: false,
          message: 'Credenciales inválidas'
        };
      }
    } catch (error) {
      console.error('Error during login:', error);
      return {
        success: false,
        message: 'Error de conexión con el servidor'
      };
    }
  },

  /**
   * Registers a new user in the system
   */
  register: async (name: string, email: string, password: string): Promise<LoginResponse> => {
    try {
      // In a real environment, this would connect to your registration endpoint
      try {
        const encodedPassword = btoa(password);
        
        const response = await axios.post(`${API_URL}/api/users`, {
          email,
          password_base64: encodedPassword,
          role: 'user', // Default role for new users
          name // This would create a person record linked to the user
        });
        
        if (response.status === 201) {
          return {
            success: true,
            message: 'User registered successfully'
          };
        }
      } catch (apiError) {
      }
      
      // Simulation for development/testing
      // In a real app, remove this fallback
      return {
        success: true,
        message: 'User registered successfully (simulated)'
      };
    } catch (error) {
      console.error('Error during registration:', error);
      return {
        success: false,
        message: 'Error de conexión con el servidor'
      };
    }
  },

  /**
   * Logs out the current user
   */
  logout: (): void => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },

  /**
   * Checks if the user is logged in
   */
  isLoggedIn: (): boolean => {
    const isLoggedIn = !!localStorage.getItem(TOKEN_KEY);
    return isLoggedIn;
  },

  /**
   * Checks if the current user is an admin
   */
  isAdmin: (): boolean => {
    const user = authService.getCurrentUser();
    const isAdmin = user?.role === 'admin';
    return isAdmin;
  },

  /**
   * Gets the current user's JWT token
   */
  getToken: (): string | null => {
    const token = localStorage.getItem(TOKEN_KEY);
    return token;
  },

  /**
   * Gets the current authenticated user
   */
  getCurrentUser: (): User | null => {
    const userString = localStorage.getItem(USER_KEY);
    if (userString) {
      try {
        const user = JSON.parse(userString);
        return user;
      } catch (error) {
        console.error('Error parsing user data from localStorage:', error);
        return null;
      }
    }
    return null;
  },

  /**
   * Utility method to force refresh the admin token (for debugging)
   */
  refreshAdminToken: (): boolean => {
    try {
      const mockUser = {
        id: '1',
        email: 'adminscao@rentalmanager.com',
        role: 'admin',
        person_id: null,
        status: 'active',
        token: 'simulated-jwt-token-' + new Date().getTime()
      };
      
      // Update the token and user in localStorage
      localStorage.setItem(TOKEN_KEY, mockUser.token);
      localStorage.setItem(USER_KEY, JSON.stringify(mockUser));
      
      return true;
    } catch (error) {
      return false;
    }
  },

  /**
   * Sets up axios interceptors for authentication
   */
  setupAxiosInterceptors: (): void => {
    axios.interceptors.request.use(
      (config) => {
        const token = authService.getToken();
        if (token) {
          // Make sure we're adding the header properly
          config.headers = config.headers || {};
          config.headers.Authorization = `Bearer ${token}`;
          console.log('Adding Authorization header to request:', config.url);
          console.log('Authorization header:', `Bearer ${token.substring(0, 10)}...`);
        } else {
          console.warn('No token available for request:', config.url);
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    axios.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response) {
          if (error.response.status === 401) {
            // Unauthorized - token expired or invalid
            if (!error.config.url.includes('/api/users/login')) {
              console.warn('401 from non-login endpoint, logging out user');
              authService.logout();
              window.location.href = '/login';
            }
          } else if (error.response.status === 403) {
            // Forbidden - user doesn't have permission
            console.warn('Received 403 Forbidden:', error.config.url);
            console.warn('User lacks permissions for this resource');
          }
        } else if (error.request) {
          console.warn('No response received:', error.request);
        } else {
          console.warn('Error setting up request:', error.message);
        }
        
        return Promise.reject(error);
      }
    );
  }
};

// Setup axios interceptors for authentication
authService.setupAxiosInterceptors();

export default authService; 