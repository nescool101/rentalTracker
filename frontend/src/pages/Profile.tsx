import { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Paper,
  TextInput,
  Button,
  Group,
  Stack,
  PasswordInput,
  Divider,
  Text,
  Alert,
  LoadingOverlay
} from '@mantine/core';
import { useAuth } from '../contexts/AuthContext';
import { personApi } from '../api/apiService';
import { notifications } from '@mantine/notifications';
import { IconAlertCircle, IconCheck } from '@tabler/icons-react';

export default function Profile() {
  const { user, logout } = useAuth();
  const [loading, setLoading] = useState(false);
  const [formError, setFormError] = useState('');
  const [fullName, setFullName] = useState('');
  const [phone, setPhone] = useState('');
  const [address, setAddress] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  
  // Check if user has edit permissions (admin/manager only)
  const canEdit = user?.role === 'admin' || user?.role === 'manager';

  useEffect(() => {
    if (user?.person_id) {
      fetchPersonData();
    }
  }, [user]);

  const fetchPersonData = async () => {
    setLoading(true);
    try {
      if (!user?.person_id) return;
      
      const personData = await personApi.getById(user.person_id);
      if (personData) {
        setFullName(personData.full_name || '');
        setPhone(personData.phone || '');
        setAddress(personData.address || '');
      }
    } catch (error) {
      console.error('Error fetching person data:', error);
      setFormError('No se pudo cargar su información personal.');
    } finally {
      setLoading(false);
    }
  };

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setFormError('');

    try {
      if (!user?.person_id) {
        setFormError('No se encontró información de perfil asociada a su cuenta.');
        return;
      }

      // Update person information
      await personApi.update(user.person_id, {
        id: user.person_id,
        full_name: fullName,
        phone,
        address
      });

      notifications.show({
        title: 'Perfil actualizado',
        message: 'Su información personal ha sido actualizada exitosamente.',
        color: 'green',
        icon: <IconCheck size={16} />
      });
    } catch (error) {
      console.error('Error updating profile:', error);
      setFormError('Error al actualizar su información personal.');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate passwords
    if (!currentPassword) {
      setFormError('Debe ingresar su contraseña actual');
      return;
    }
    
    if (!newPassword || newPassword.length < 8) {
      setFormError('La nueva contraseña debe tener al menos 8 caracteres');
      return;
    }
    
    if (newPassword !== confirmPassword) {
      setFormError('Las contraseñas nuevas no coinciden');
      return;
    }
    
    setLoading(true);
    setFormError('');
    
    try {
      if (!user?.id) {
        setFormError('Información de usuario incompleta');
        return;
      }
      
      // Use the new change-password endpoint for better security
      const response = await fetch(`${import.meta.env.VITE_API_URL || ''}/api/users/${user.id}/change-password`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${user.token || localStorage.getItem('auth_token')}`
        },
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Error al actualizar la contraseña');
      }
      
      notifications.show({
        title: 'Contraseña actualizada',
        message: 'Su contraseña ha sido actualizada exitosamente. Por favor, inicie sesión nuevamente.',
        color: 'green',
        icon: <IconCheck size={16} />
      });
      
      // Clear password fields
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      
      // Force re-login with new password
      await logout();
      setTimeout(() => {
        window.location.href = '/login';
      }, 2000);
      
    } catch (error) {
      console.error('Error updating password:', error);
      setFormError(error instanceof Error ? error.message : 'Error al actualizar su contraseña. Por favor, inténtelo de nuevo.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container size="md">
      <Title order={2} mb="md">Mi Perfil</Title>
      <Text c="dimmed" mb="xl">Actualice su información personal y contraseña</Text>
      
      {formError && (
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Error"
          color="red"
          mb="md"
        >
          {formError}
        </Alert>
      )}
      
      <Paper withBorder p="xl" radius="md" mb="xl" pos="relative">
        <LoadingOverlay visible={loading} />
        
        <Title order={3} mb="lg">Información Personal</Title>
        {!canEdit && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            title="Solo lectura"
            color="blue"
            mb="md"
          >
            Su información personal es de solo lectura. Solo los administradores pueden modificar estos datos.
          </Alert>
        )}
        <form onSubmit={handleProfileUpdate}>
          <Stack>
            <TextInput
              label="Nombre completo"
              placeholder="Nombre completo"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              required
              readOnly={!canEdit}
              disabled={!canEdit}
            />
            
            <TextInput
              label="Teléfono"
              placeholder="Teléfono de contacto"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              readOnly={!canEdit}
              disabled={!canEdit}
            />
            
            <TextInput
              label="Dirección"
              placeholder="Dirección"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              readOnly={!canEdit}
              disabled={!canEdit}
            />
            
            {canEdit && (
              <Group justify="flex-end" mt="md">
                <Button type="submit">Guardar cambios</Button>
              </Group>
            )}
          </Stack>
        </form>
      </Paper>
      
      <Paper withBorder p="xl" radius="md" pos="relative">
        <LoadingOverlay visible={loading} />
        
        <Title order={3} mb="lg">Cambiar Contraseña</Title>
        <form onSubmit={handlePasswordUpdate}>
          <Stack>
            <PasswordInput
              label="Contraseña actual"
              placeholder="Ingrese su contraseña actual"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
            />
            
            <Divider my="md" />
            
            <PasswordInput
              label="Nueva contraseña"
              placeholder="Ingrese su nueva contraseña"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
            
            <PasswordInput
              label="Confirmar nueva contraseña"
              placeholder="Confirme su nueva contraseña"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
            
            <Text size="xs" c="dimmed">
              La contraseña debe tener al menos 8 caracteres.
            </Text>
            
            <Group justify="flex-end" mt="md">
              <Button type="submit" color="blue">Actualizar contraseña</Button>
            </Group>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
} 