import { Container, Paper, Title, Text, Button, Group, ThemeIcon, Stack, Divider } from '@mantine/core';
import { IconAlertCircle, IconMail, IconArrowRight } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { Link } from 'react-router-dom';

export function PendingActivationNotice() {
  const { user } = useAuth();

  const sendEmail = () => {
    window.location.href = 'mailto:nescool101@gmail.com?subject=Activación%20de%20cuenta&body=Hola,%0A%0ASolicito%20la%20activación%20de%20mi%20cuenta%20en%20el%20sistema%20de%20gestión%20de%20propiedades.%0A%0AMi%20correo%20es:%20' + 
      (user?.email || '') + '%0A%0AGracias.';
  };

  return (
    <Container size="lg" mt={30}>
      <Paper radius="md" p="xl" withBorder>
        <Stack align="center" gap="lg">
          <ThemeIcon size={60} radius={60} color="yellow">
            <IconAlertCircle size={40} />
          </ThemeIcon>
          
          <Title order={2} ta="center">BIENVENIDO AL SISTEMA DE GESTIÓN</Title>
          
          <Text ta="center" size="lg" mb={10}>
            ESTAMOS ESPERANDO TU PAGO O QUE EL ADMINISTRADOR CAMBIE TU ESTADO A ACTIVO
          </Text>
          
          <Divider w="100%" my="sm" />
          
          <Text ta="center" mb={5} fw={500} size="md">
            Mientras tanto, puedes:
          </Text>
          
          <Group grow w="100%" mb={10}>
            <Button 
              component={Link} 
              to="/profile" 
              variant="light" 
              rightSection={<IconArrowRight size={16} />}
            >
              Cambiar tu contraseña
            </Button>
            
            <Button 
              component={Link} 
              to="/dashboard" 
              variant="light" 
              rightSection={<IconArrowRight size={16} />}
            >
              Ver tu panel principal
            </Button>
          </Group>
          
          <Divider w="100%" my="sm" />
          
          <Text ta="center" mb={5}>
            Para más información, escríbele a:
          </Text>
          
          <Text ta="center" fw={700} size="lg" color="blue">
            NESCOOL101@gmail.com
          </Text>
          
          <Group mt={20}>
            <Button 
              variant="filled" 
              color="blue" 
              leftSection={<IconMail size={18} />}
              onClick={sendEmail}
            >
              Enviar correo ahora
            </Button>
          </Group>
        </Stack>
      </Paper>
    </Container>
  );
} 