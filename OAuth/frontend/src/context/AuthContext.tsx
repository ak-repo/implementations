/**
 * Authentication Context
 * 
 * Provides global authentication state management using React Context API.
 * Features:
 * - Automatic user session validation on app load
 * - Login/logout functionality
 * - Loading states and error handling
 * - Protected route support via useAuth hook
 * 
 * Security Note: JWT tokens are stored in httpOnly cookies by the backend,
 * so they're not accessible to JavaScript (XSS protection).
 */

import {
  createContext,
  useContext,
  useReducer,
  useEffect,
  ReactNode,
} from 'react';
import type { User, AuthState } from '../types/auth';
import { fetchUser, logoutUser, redirectToGoogleLogin } from '../api/auth';

// Auth context actions type
export type AuthAction =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: User }
  | { type: 'AUTH_FAILURE'; payload: string }
  | { type: 'LOGOUT' };

// Initial authentication state
const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,
};

// Reducer for auth state management
function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'AUTH_START':
      return { ...state, isLoading: true, error: null };
    case 'AUTH_SUCCESS':
      return {
        ...state,
        user: action.payload,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      };
    case 'AUTH_FAILURE':
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: action.payload,
      };
    case 'LOGOUT':
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      };
    default:
      return state;
  }
}

// Context type definition
interface AuthContextType extends AuthState {
  login: () => void;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

// Create the context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider component
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  /**
   * Check if user is already authenticated on mount
   * This runs when the app loads to validate the existing session
   */
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const user = await fetchUser();
        dispatch({ type: 'AUTH_SUCCESS', payload: user });
      } catch {
        // User is not authenticated - this is expected for logged-out users
        dispatch({ type: 'AUTH_FAILURE', payload: '' });
      }
    };

    checkAuth();
  }, []);

  /**
   * Initiate Google OAuth login
   * Redirects to backend /auth/google/login endpoint
   */
  const login = () => {
    dispatch({ type: 'AUTH_START' });
    redirectToGoogleLogin();
  };

  /**
   * Log out the current user
   * Clears the JWT cookie via backend API
   */
  const logout = async () => {
    try {
      await logoutUser();
      dispatch({ type: 'LOGOUT' });
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  /**
   * Refresh user data from the server
   * Useful after profile updates
   */
  const refreshUser = async () => {
    try {
      const user = await fetchUser();
      dispatch({ type: 'AUTH_SUCCESS', payload: user });
    } catch (error) {
      dispatch({
        type: 'AUTH_FAILURE',
        payload: error instanceof Error ? error.message : 'Failed to refresh user',
      });
    }
  };

  const value: AuthContextType = {
    ...state,
    login,
    logout,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Custom hook for consuming auth context
export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
