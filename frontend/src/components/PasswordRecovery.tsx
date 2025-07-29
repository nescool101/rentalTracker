import { useState } from 'react';
import { Modal, TextInput, Button, Stack, Text, Alert } from '@mantine/core';
import { IconAlertCircle, IconCheck } from '@tabler/icons-react';

interface PasswordRecoveryProps {
  opened: boolean;
  onClose: () => void;
}

export default function PasswordRecovery({ opened, onClose }: PasswordRecoveryProps) {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!email || !/^\S+@\S+$/.test(email)) {
      setError('Por favor, ingrese un email válido');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Simulate API call for password recovery
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // For now, just show success message
      setSuccess(true);
      
      // Reset form after 3 seconds
      setTimeout(() => {
        setSuccess(false);
        setEmail('');
        onClose();
      }, 3000);
      
    } catch (err) {
      setError('Error al enviar el email de recuperación. Intente nuevamente.');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setEmail('');
    setError('');
    setSuccess(false);
    onClose();
  };

  return (
    <Modal
      opened={opened}
      onClose={handleClose}
      title="Recuperar Contraseña"
      centered
    >
      {success ? (
        <Alert 
          icon={<IconCheck size={16} />} 
          title="Email enviado" 
          color="green"
        >
          Se ha enviado un email con instrucciones para recuperar su contraseña.
        </Alert>
      ) : (
        <form onSubmit={handleSubmit}>
          <Stack>
            <Text size="sm" c="dimmed">
              Ingrese su email y le enviaremos instrucciones para recuperar su contraseña.
            </Text>
            
            {error && (
              <Alert 
                icon={<IconAlertCircle size={16} />} 
                title="Error" 
                color="red"
              >
                {error}
              </Alert>
            )}
            
            <TextInput
              label="Email"
              placeholder="tu@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            
            <Button type="submit" loading={loading} fullWidth>
              Enviar instrucciones
            </Button>
          </Stack>
        </form>
      )}
    </Modal>
  );
} 