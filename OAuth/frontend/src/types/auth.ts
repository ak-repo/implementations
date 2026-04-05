/**
 * Type definitions for the authentication system
 */

// User type representing the authenticated user
export interface User {
  id: number;
  email: string;
  name: string;
  picture: string;
  created_at: string;
}

// Auth state for the context
export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

// Auth context actions
export type AuthAction =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: User }
  | { type: 'AUTH_FAILURE'; payload: string }
  | { type: 'LOGOUT' };

// API response types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
}
