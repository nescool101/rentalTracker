import React from 'react';
import { Link } from 'react-router-dom';
import { Container, Text, Title, Button, Group, Box } from '@mantine/core';

const AccessDenied: React.FC = () => {
  return (
    <Container mt={80} mb={80}>
      <Box>
        <Title fw={900} fz={34} mb="md" c="red.6">
          Acceso Denegado
        </Title>
        <Text fz={24} fw={500}>
          No tienes permiso para acceder a esta página
        </Text>
        <Text fz={18} c="dimmed" mt="xl" mb={30}>
          La página a la que intentas acceder está restringida y requiere
          privilegios administrativos. Por favor contacta a un administrador si
          crees que deberías tener acceso a este recurso.
        </Text>
        <Group>
          <Button component={Link} to="/" variant="outline" size="md">
            Volver al Inicio
          </Button>
        </Group>
      </Box>
    </Container>
  );
};

export default AccessDenied; 