import { Title, Text, Card, RingProgress, Group, Stack, Container, ThemeIcon, Anchor, Paper, Button, Alert } from '@mantine/core';
import { useQuery } from '@tanstack/react-query';
import { personApi, propertyApi, rentalApi } from '../api/apiService';
import { useAuth } from '../contexts/AuthContext';
import { Link } from 'react-router-dom';
import { IconUserCircle, IconHomeCog, IconCreditCard, IconHistory, IconBuildingCommunity, IconAlertCircle, IconSettings, IconCloudUpload } from '@tabler/icons-react';
import styles from './Dashboard.module.css';

export default function Dashboard() {
  const { user } = useAuth();
  const isAdminOrManager = user?.role === 'admin' || user?.role === 'manager';
  const isPendingActivation = user?.status === 'activenopaid';

  const { data: persons = [] } = useQuery({
    queryKey: ['persons', 'all', 'dashboard'],
    queryFn: personApi.getAll,
    enabled: isAdminOrManager && !isPendingActivation,
  });

  const { data: properties = [] } = useQuery({
    queryKey: ['properties', 'all', 'dashboard'],
    queryFn: propertyApi.getAll,
    enabled: isAdminOrManager && !isPendingActivation,
  });

  const { data: rentals = [] } = useQuery({
    queryKey: ['rentals', 'all', 'dashboard'],
    queryFn: rentalApi.getAll,
    enabled: isAdminOrManager && !isPendingActivation,
  });

  let activeRentals: any[] = [];
  let pendingRenewals: any[] = [];
  let stats: any[] = [];

  if (isAdminOrManager && !isPendingActivation) {
    activeRentals = rentals.filter(rental => new Date(rental.end_date) > new Date());
    pendingRenewals = rentals.filter(rental => {
      const endDate = new Date(rental.end_date);
      const today = new Date();
      const diff = endDate.getTime() - today.getTime();
      const diffDays = Math.ceil(diff / (1000 * 3600 * 24));
      return diffDays <= 30 && diffDays > 0;
    });

    stats = [
      { title: 'Total Propiedades', value: properties.length, color: 'blue' },
      { title: 'Alquileres Activos', value: activeRentals.length, color: 'green' },
      { title: 'Renovaciones Pendientes', value: pendingRenewals.length, color: 'orange' },
      { title: 'Total Personas', value: persons.length, color: 'grape' },
    ];
  }

  // Special view for pending activation users
  if (isPendingActivation) {
    return (
      <Container size="xl" className={styles.dashboardContainer}>
        <Title order={1} mb="lg">Panel Principal</Title>
        
        <div className={styles.alertContainer}>
          <Alert 
            icon={<IconAlertCircle size={24} />} 
            title="Cuenta pendiente de activación" 
            color="yellow"
            variant="filled"
          >
            Tu cuenta está pendiente de activación por el administrador. Mientras tanto, tienes acceso limitado al sistema.
          </Alert>
        </div>
        
        <Paper withBorder radius="md" shadow="sm" className={styles.pendingCard}>
          <Title order={3} mb="lg">Acciones Disponibles</Title>
          <div className={styles.quickAccessGrid}>
            <Card withBorder radius="md" className={styles.quickAccessCard}>
              <Stack align="center">
                <IconSettings size={32} color="blue" />
                <Title order={4}>Administrar Perfil</Title>
                <Text size="sm" ta="center">Puedes cambiar tu contraseña y verificar tus datos personales</Text>
                <Button 
                  component={Link} 
                  to="/profile" 
                  variant="light" 
                  color="blue" 
                  mt="auto"
                  fullWidth
                >
                  Ir a Mi Perfil
                </Button>
              </Stack>
            </Card>
            
            <Card withBorder radius="md" className={styles.quickAccessCard}>
              <Stack align="center">
                <IconAlertCircle size={32} color="orange" />
                <Title order={4}>Contactar Soporte</Title>
                <Text size="sm" ta="center">Si necesitas ayuda o quieres consultar el estado de tu activación</Text>
                <Button 
                  component="a"
                  href="mailto:nescool101@gmail.com?subject=Activación%20de%20cuenta" 
                  variant="light" 
                  color="orange" 
                  mt="auto"
                  fullWidth
                >
                  Enviar Correo
                </Button>
              </Stack>
            </Card>
          </div>
        </Paper>
        
        <Text mt="xl" size="sm" c="dimmed" ta="center">
          Una vez que tu cuenta sea activada, tendrás acceso a todas las funcionalidades del sistema.
        </Text>
      </Container>
    );
  }

  return (
    <Container size="xl" className={styles.dashboardContainer}>
      <Title order={1} mb="lg">Panel Principal</Title>
      
      {isAdminOrManager ? (
        <>
          <Text mb="md">Bienvenido a tu Panel de Gestión de Alquileres. Aquí tienes un resumen rápido:</Text>
          <div className={styles.statsGrid}>
            {stats.map((stat) => (
              <Card key={stat.title} withBorder radius="md" shadow="sm" className={styles.statCard}>
                <Group justify="space-between" align="flex-start">
                  <Stack gap={0}>
                    <Text fz="lg" fw={500}>{stat.title}</Text>
                    <Title order={3}>{stat.value}</Title>
                  </Stack>
                  <RingProgress
                    size={80}
                    thickness={8}
                    roundCaps
                    sections={[{ value: stat.value > 0 ? 100 : 0, color: stat.color }]}
                  />
                </Group>
              </Card>
            ))}
          </div>
        </>
      ) : (
        <>
          <Text mb="md" c="dimmed">Bienvenido a tu portal personal. Desde aquí puedes acceder a tu información y servicios:</Text>
          <Paper withBorder radius="md" shadow="sm" className={styles.quickAccessList}>
            <Title order={3} mb="lg" style={{ padding: '24px 24px 0 24px' }}>Accesos Rápidos</Title>
            <div style={{ padding: '0 24px 24px 24px' }}>
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="blue" size={32} radius="xl">
                    <IconUserCircle size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/profile" className={styles.listItemTitle}>Mi Perfil</Anchor>
                  <Text className={styles.listItemDescription}>Ver y actualizar tu información personal.</Text>
                </div>
              </div>
              
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="teal" size={32} radius="xl">
                    <IconBuildingCommunity size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/properties" className={styles.listItemTitle}>Detalles de mi Propiedad</Anchor>
                  <Text className={styles.listItemDescription}>Información sobre tu unidad alquilada.</Text>
                </div>
              </div>
              
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="cyan" size={32} radius="xl">
                    <IconHomeCog size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/maintenance" className={styles.listItemTitle}>Solicitudes de Mantenimiento</Anchor>
                  <Text className={styles.listItemDescription}>Enviar y rastrear solicitudes de reparaciones.</Text>
                </div>
              </div>
              
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="grape" size={32} radius="xl">
                    <IconCreditCard size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/payments" className={styles.listItemTitle}>Mis Pagos</Anchor>
                  <Text className={styles.listItemDescription}>Ver tu historial de pagos y próximos vencimientos.</Text>
                </div>
              </div>
              
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="orange" size={32} radius="xl">
                    <IconHistory size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/rental-history" className={styles.listItemTitle}>Historial de Alquiler</Anchor>
                  <Text className={styles.listItemDescription}>Detalles de tus contratos pasados y actuales.</Text>
                </div>
              </div>
              
              <div className={styles.listItem}>
                <div className={styles.listItemIcon}>
                  <ThemeIcon color="indigo" size={32} radius="xl">
                    <IconCloudUpload size="1.2rem" />
                  </ThemeIcon>
                </div>
                <div className={styles.listItemContent}>
                  <Anchor component={Link} to="/file-upload" className={styles.listItemTitle}>Subir Archivos</Anchor>
                  <Text className={styles.listItemDescription}>Subir documentos y archivos relacionados con tu alquiler.</Text>
                </div>
              </div>
            </div>
          </Paper>
        </>
      )}
    </Container>
  );
} 