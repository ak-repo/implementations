/**
 * API client for authentication endpoints
 * 
 * All requests automatically include credentials (cookies) for JWT authentication.
 * The backend uses httpOnly cookies, so we don't handle tokens in JavaScript.
 */

import type { User } from '../types/auth';

// API base URL from environment variables
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8090';

/**
 * Default fetch options for all API requests
 * - credentials: 'include' sends cookies with requests (required for httpOnly JWT cookies)
 * - Content-Type: application/json for POST requests
 */
const defaultOptions: RequestInit = {
  credentials: 'include', // Required to send/receive cookies
  headers: {
    'Content-Type': 'application/json',
  },
};

/**
 * Fetch the current authenticated user's profile
 * GET /api/me
 * 
 * This endpoint requires authentication via JWT cookie.
 * Returns 401 if user is not authenticated.
 */
export async function fetchUser(): Promise<User> {
  const response = await fetch(`${API_BASE_URL}/api/me`, {
    ...defaultOptions,
    method: 'GET',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error('Unauthorized');
    }
    throw new Error('Failed to fetch user');
  }

  return response.json();
}

/**
 * Log out the current user
 * POST /auth/logout
 * 
 * Clears the JWT cookie on the server side.
 * After logout, the user will need to re-authenticate.
 */
export async function logoutUser(): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/auth/logout`, {
    ...defaultOptions,
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error('Logout failed');
  }
}

/**
 * Redirect to Google OAuth login
 * GET /auth/google/login
 * 
 * This initiates the OAuth flow. The user is redirected to Google's
 * consent screen, then back to our callback endpoint.
 * 
 * Note: We don't use fetch for this - we directly redirect the browser.
 */
export function redirectToGoogleLogin(): void {
  window.location.href = `${API_BASE_URL}/auth/google/login`;
}
