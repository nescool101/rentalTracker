import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import AccessDenied from '../pages/AccessDenied';
import { debugLog } from '../utils/debug';

// Enable this for debugging - uncomment when needed
// const DEBUG = true; // Temporarily enabled for debugging login issues

interface ProtectedRouteProps {
  requireAdmin?: boolean;
}

/**
 * A component that wraps routes which require authentication
 * It can also restrict access to admin users only
 */
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ requireAdmin = false }) => {
  const { isAuthenticated, user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const isNewUser = user?.status === 'newuser';
  const isPendingActivation = user?.status === 'activenopaid';
  const location = useLocation();
  const isDashboardPath = location.pathname === '/dashboard' || location.pathname === '/dashboard/';

  debugLog('ProtectedRoute check:', { 
    isAuthenticated, 
    requireAdmin, 
    isAdmin,
    userRole: user?.role,
    userStatus: user?.status,
    path: location.pathname,
    isPendingActivation,
    isDashboardPath
  });

  // If not authenticated, redirect to login
  if (!isAuthenticated) {
    debugLog('User not authenticated, redirecting to login');
    return <Navigate to="/login" replace />;
  }

  // If user is new and not already on the onboarding page, redirect to onboarding
  if (isNewUser && location.pathname !== '/onboarding') {
    debugLog('New user detected, redirecting to onboarding page');
    return <Navigate to="/onboarding" replace />;
  }

  // If user has pending activation status and not trying to access dashboard, redirect to dashboard
  if (isPendingActivation && !isDashboardPath) {
    debugLog('User with pending activation trying to access non-dashboard route, redirecting to dashboard');
    return <Navigate to="/dashboard" replace />;
  }

  // If admin access is required but user is not admin, show access denied
  if (requireAdmin && !isAdmin) {
    debugLog('Admin access required but user is not admin, showing AccessDenied');
    return <AccessDenied />;
  }

  // User is authenticated and has appropriate permissions
  debugLog('Access granted, rendering child routes');
  return <Outlet />;
};

export default ProtectedRoute; 