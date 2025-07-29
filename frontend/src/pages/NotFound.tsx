import { Container, Title, Text, Button, Group } from '@mantine/core';
import { Link } from 'react-router-dom';

export default function NotFound() {
  return (
    <Container size="md" py={80} style={{ textAlign: 'center' }}>
      <Title order={1} size={120} mb={0} fw={900} lh={1}>404</Title>
      <Title order={2} mb="xl">Página no encontrada</Title>
      <Text size="lg" mb="xl">
        La página que estás buscando no existe o ha sido movida.
      </Text>
      <Group justify="center">
        <Button component={Link} to="/" size="md">
          Volver al inicio
        </Button>
      </Group>
    </Container>
  );
} 