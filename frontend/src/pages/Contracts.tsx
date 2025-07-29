import { 
  Title, 
  Container, 
  Text, 
  Paper, 
  Group, 
  ThemeIcon,
  List,
  Box,
  Divider,
  Button,
  Badge,
  Stack
} from '@mantine/core';
import { IconFileDescription, IconHistory, IconCheckbox, IconAlertCircle } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';

export default function Contracts() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';

  return (
    <Container size="xl">
      <Title order={1} mb="xl">Gestión de Contratos</Title>
      
      <Paper shadow="sm" p="lg" radius="md" withBorder mb="xl">
        <Group mb="md">
          <ThemeIcon size="xl" color="indigo" radius="md">
            <IconFileDescription size={24} />
          </ThemeIcon>
          <Title order={2}>Próximamente</Title>
        </Group>
        
        <Text mb="lg">
          El sistema de gestión de Contratos está actualmente en desarrollo y estará disponible pronto. 
          Este módulo proporcionará herramientas completas para crear, seguir y gestionar contratos de alquiler.
        </Text>

        <Divider my="xl" label="Características Planeadas" labelPosition="center" />
        
        <Group grow align="flex-start" mb="xl">
          <Paper p="md" radius="md" withBorder>
            <Title order={3} mb="md" size="h4">Contratos de Alquiler</Title>
            <Text size="sm" mb="md" c="dimmed">
              Basado en la tabla <Badge>rental</Badge> en la base de datos
            </Text>
            <List spacing="xs" size="sm" center icon={
              <ThemeIcon color="violet" size={20} radius="xl">
                <IconCheckbox size={14} />
              </ThemeIcon>
            }>
              <List.Item>Crear y gestionar contratos de alquiler</List.Item>
              <List.Item>Seguimiento de fechas de inicio y finalización</List.Item>
              <List.Item>Definir términos y condiciones de pago</List.Item>
              <List.Item>Asociar propiedades con inquilinos</List.Item>
              <List.Item>Monitorear meses impagos y estado de pagos</List.Item>
              <List.Item>Vinculación con información de cuentas bancarias para pagos</List.Item>
            </List>
          </Paper>
          
          <Paper p="md" radius="md" withBorder>
            <Title order={3} mb="md" size="h4">Historial de Alquileres</Title>
            <Text size="sm" mb="md" c="dimmed">
              Basado en la tabla <Badge>rental_history</Badge> en la base de datos
            </Text>
            <List spacing="xs" size="sm" center icon={
              <ThemeIcon color="green" size={20} radius="xl">
                <IconHistory size={14} />
              </ThemeIcon>
            }>
              <List.Item>Seguimiento completo del historial de alquileres</List.Item>
              <List.Item>Registro de cambios de estado a lo largo del tiempo</List.Item>
              <List.Item>Documentar razones de terminación de contratos</List.Item>
              <List.Item>Mantener registros de todas las relaciones pasadas con inquilinos</List.Item>
              <List.Item>Generar informes de historial de alquileres</List.Item>
            </List>
          </Paper>
        </Group>
        
        {isAdmin && (
          <Box>
            <Divider my="lg" />
            <Group justify="space-between">
              <Stack gap="xs">
                <Title order={4}>Características de Administrador</Title>
                <Text size="sm" color="dimmed">Estas características adicionales estarán disponibles para usuarios administradores</Text>
              </Stack>
              <ThemeIcon size="xl" color="yellow" radius="md">
                <IconAlertCircle size={24} />
              </ThemeIcon>
            </Group>
            <List spacing="xs" size="sm" mt="md">
              <List.Item>Crear y modificar contratos para cualquier propiedad</List.Item>
              <List.Item>Anular términos y condiciones de contratos según sea necesario</List.Item>
              <List.Item>Configurar notificaciones automáticas para eventos de contratos</List.Item>
              <List.Item>Generar informes sobre contratos activos e inactivos</List.Item>
              <List.Item>Gestionar plantillas legales y generación de documentos de contrato</List.Item>
            </List>
          </Box>
        )}
      </Paper>
      
      <Group justify="center">
        <Button variant="outline" size="md" disabled>Vuelva Pronto</Button>
      </Group>
    </Container>
  );
} 