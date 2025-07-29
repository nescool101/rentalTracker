import { useState, useEffect } from 'react';
import { 
  Container, 
  Paper, 
  Title, 
  PasswordInput, 
  Button, 
  Group, 
  Anchor,
  Stack,
  Checkbox,
  TextInput,
  Alert,
  Text
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { useAuth } from '../contexts/AuthContext';
import { notifications } from '@mantine/notifications';
import { IconAlertCircle } from '@tabler/icons-react';
import PasswordRecovery from '../components/PasswordRecovery';

const MAX_LOGIN_ATTEMPTS = 3;

export default function Login() {
  const [loading, setLoading] = useState(false);
  const [isRegister, setIsRegister] = useState(false);
  const [loginAttempts, setLoginAttempts] = useState(0);
  const [loginError, setLoginError] = useState('');
  const [recoveryModalOpen, setRecoveryModalOpen] = useState(false);
  const { login, register } = useAuth();
  
  // Reset login error when switching between login and register
  useEffect(() => {
    setLoginError('');
  }, [isRegister]);
  
  // Get the last authentication error from localStorage
  useEffect(() => {
    const lastError = localStorage.getItem('last_auth_error');
    if (lastError) {
      setLoginError(lastError);
      localStorage.removeItem('last_auth_error');
    }
  }, [loading]);
  
  const form = useForm({
    initialValues: {
      email: '',
      password: '',
      confirmPassword: '',
      name: '',
      terms: false,
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Email inválido'),
      password: (value) => (value.length < 6 ? 'La contraseña debe tener al menos 6 caracteres' : null),
      confirmPassword: (value, values) => 
        isRegister && value !== values.password ? 'Las contraseñas no coinciden' : null,
      name: (value) => (isRegister && value.trim().length < 2 ? 'El nombre es requerido' : null),
      terms: (value) => (isRegister && !value ? 'Debes aceptar los términos y condiciones' : null),
    },
  });

  const handleSubmit = async (values: typeof form.values) => {
    setLoading(true);
    setLoginError('');
    
    try {
      let success = false;
      
      if (isRegister) {
        // Registro de usuario
        success = await register(values.name, values.email, values.password);
        if (success) {
          // Mostrar mensaje y cambiar a modo login
          setIsRegister(false);
          form.setValues({ 
            email: values.email, 
            password: '',
            confirmPassword: '',
            name: '',
            terms: false
          });
        }
      } else {
        // Login de usuario
        success = await login(values.email, values.password);
        
        if (!success) {
          // Get the last error message from authContext via notification
          const errorEvent = new CustomEvent('get-last-auth-error');
          window.dispatchEvent(errorEvent);
          
          // Increment failed attempts counter
          const newAttempts = loginAttempts + 1;
          setLoginAttempts(newAttempts);
          
          // Set custom error message based on attempts
          if (newAttempts >= MAX_LOGIN_ATTEMPTS) {
            setLoginError(`Demasiados intentos fallidos (${newAttempts}/${MAX_LOGIN_ATTEMPTS}). Serás redirigido a la página principal.`);
            
            // Redirect to home page after 3 failed attempts
            setTimeout(() => {
              notifications.show({
                title: 'Demasiados intentos',
                message: 'Redirigiendo a la página principal',
                color: 'red'
              });
              window.location.href = '/';
            }, 2000);
          }
        }
        // Note: Successful login redirection is handled in the AuthContext
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container size={420} my={40}>
      <Paper radius="md" p="xl" withBorder>
        <Title order={2} ta="center" mb="lg">
          {isRegister ? 'Crear una cuenta' : 'Iniciar sesión'}
        </Title>

        {loginError && (
          <Alert 
            icon={<IconAlertCircle size={16} />} 
            title="Error de autenticación" 
            color="red" 
            mb="md"
          >
            {loginError}
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            {isRegister && (
              <TextInput
                label="Nombre completo"
                placeholder="Tu nombre"
                required
                {...form.getInputProps('name')}
              />
            )}

            <TextInput
              label="Email"
              placeholder="tu@email.com"
              required
              {...form.getInputProps('email')}
            />

            <PasswordInput
              label="Contraseña"
              placeholder="Tu contraseña"
              required
              {...form.getInputProps('password')}
            />

            {isRegister && (
              <>
                <PasswordInput
                  label="Confirmar contraseña"
                  placeholder="Confirma tu contraseña"
                  required
                  {...form.getInputProps('confirmPassword')}
                />

                <Checkbox
                  label="Acepto los términos y condiciones"
                  {...form.getInputProps('terms', { type: 'checkbox' })}
                />
              </>
            )}
            
            {!isRegister && (
              <Anchor 
                component="button" 
                type="button" 
                c="dimmed" 
                size="sm"
                onClick={() => setRecoveryModalOpen(true)} 
                style={{ alignSelf: 'flex-start' }}
              >
                ¿Olvidó su contraseña?
              </Anchor>
            )}
          </Stack>

          <Group justify="space-between" mt="md">
            {isRegister ? (
              <Anchor component="button" type="button" c="dimmed" onClick={() => setIsRegister(false)} size="sm">
                ¿Ya tienes una cuenta? Inicia sesión
              </Anchor>
            ) : (
              <Text size="sm" c="dimmed">
                El registro solo está disponible para usuarios autorizados
              </Text>
            )}
            <Button type="submit" loading={loading} disabled={loginAttempts >= MAX_LOGIN_ATTEMPTS}>
              {isRegister ? 'Registrarse' : 'Iniciar sesión'}
            </Button>
          </Group>
        </form>
      </Paper>
      
      {/* Password Recovery Modal */}
      <PasswordRecovery 
        opened={recoveryModalOpen}
        onClose={() => setRecoveryModalOpen(false)}
      />
    </Container>
  );
} 