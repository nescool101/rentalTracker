import React from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Box, Text, Title, Code, Paper, Button, Group } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import authService from '../services/authService';

/**
 * Debug component to show auth status and admin role info
 */
const DebugAuth: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  
  const handleTokenTest = () => {
    // Try an admin-protected route
    fetch('/api/persons', {
      headers: {
        'Authorization': `Bearer ${authService.getToken()}`
      }
    })
      .then(response => {
        alert(`API Response status: ${response.status}`);
        if (!response.ok) {
          throw new Error(`API returned ${response.status}`);
        }
        return response.json();
      })
      .then(data => {
        console.log('API data:', data);
        alert('API call successful! Check console for data.');
      })
      .catch(error => {
        console.error('API call failed:', error);
        alert(`API call failed: ${error.message}`);
      });
  };
  
  const handleForceAdminToken = () => {
    if (authService.refreshAdminToken()) {
      alert('Admin token refreshed. Reloading page...');
      window.location.reload();
    } else {
      alert('Failed to refresh admin token.');
    }
  };
  
  return (
    <Box p="xl" maw={600} mx="auto" my="xl">
      <Title order={2} mb="md">Authentication Debug</Title>
      
      <Paper p="md" withBorder mb="lg">
        <Title order={3} mb="sm">Auth Status</Title>
        <Text mb="xs"><b>Is Authenticated:</b> {isAuthenticated ? 'Yes' : 'No'}</Text>
        
        <Title order={3} mt="md" mb="sm">User Info</Title>
        {user ? (
          <>
            <Text mb="xs"><b>User ID:</b> {user.id}</Text>
            <Text mb="xs"><b>Email:</b> {user.email}</Text>
            <Text mb="xs"><b>Role:</b> {user.role}</Text>
            <Text mb="xs"><b>Is Admin:</b> {user.role === 'admin' ? 'Yes' : 'No'}</Text>
            <Text mb="xs"><b>Person ID:</b> {user.person_id || 'None'}</Text>
            <Text mb="md"><b>Token Available:</b> {authService.getToken() ? 'Yes' : 'No'}</Text>
            
            <Title order={4} mt="md" mb="sm">Token</Title>
            <Code block mb="md">{authService.getToken() || 'No token'}</Code>
            
            <Title order={4} mt="md" mb="sm">Full User Object</Title>
            <Code block mb="md">{JSON.stringify(user, null, 2)}</Code>
          </>
        ) : (
          <Text c="dimmed">No user logged in</Text>
        )}
      </Paper>
      
      <Group>
        <Button onClick={handleTokenTest} color="blue">
          Test API Access
        </Button>
        
        {isAuthenticated && (
          <Button onClick={logout} color="red">
            Logout
          </Button>
        )}
        
        <Button onClick={() => navigate('/login')} variant="outline">
          Go to Login
        </Button>
        
        {isAuthenticated && user?.role === 'admin' && (
          <Button onClick={() => navigate('/dashboard')} color="green">
            Go to Admin Dashboard
          </Button>
        )}
        
        {isAuthenticated && user?.role === 'admin' && (
          <Button onClick={handleForceAdminToken} color="yellow">
            Force Refresh Admin Token
          </Button>
        )}
        
        {!isAuthenticated || user?.role !== 'admin' ? (
          <Button onClick={handleForceAdminToken} color="orange">
            Force Admin Access
          </Button>
        ) : null}
      </Group>
    </Box>
  );
};

export default DebugAuth; 