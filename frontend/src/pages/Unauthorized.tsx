import { Container, Title, Text, Button, Group } from '@mantine/core';
import { Link } from 'react-router-dom';
import { IconAlertTriangle } from '@tabler/icons-react';

export default function Unauthorized() {
  return (
    <Container size="md" py={80} style={{ textAlign: 'center' }}>
      <IconAlertTriangle size={120} color="orange" style={{ marginBottom: 20 }} />
      <Title order={1} mb="md">Acceso Denegado</Title>
      <Text size="xl" mb="xl">
        No tienes permisos para acceder a esta página. Por favor, contacta al administrador si crees que esto es un error.
      </Text>
      <Group justify="center">
        <Button component={Link} to="/" size="md">
          Volver al Inicio
        </Button>
        <Button component={Link} to="/dashboard" variant="outline" size="md">
          Ir al Dashboard
        </Button>
      </Group>
    </Container>
  );
} 