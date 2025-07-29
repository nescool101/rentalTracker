import { Container, Title, Text, Button, Grid, Card, Image, List, ThemeIcon, Group, Badge, Box } from '@mantine/core';
import { Link, useLocation } from 'react-router-dom';
import { IconCheck, IconHome, IconCoin, IconUsers, IconArrowRight } from '@tabler/icons-react';
import { useEffect } from 'react';
import { notifications } from '@mantine/notifications';

export default function Landing() {
  const location = useLocation();
  
  useEffect(() => {
    // Check if we're coming from onboarding completion
    const urlParams = new URLSearchParams(location.search);
    const newManager = urlParams.get('newManager');
    
    if (newManager === 'true') {
      // Display a welcome message for new managers
      setTimeout(() => {
        notifications.show({
          title: '¡Bienvenido a Rental Manager!',
          message: 'Tu cuenta de administrador ha sido creada exitosamente. Inicia sesión para comenzar a gestionar tus propiedades.',
          color: 'green',
          icon: <IconCheck size={20} />,
          autoClose: 8000,
        });
      }, 500);
    }
  }, [location]);
  
  return (
    <Box>
      {/* Hero Section */}
      <Box 
        style={{ 
          backgroundImage: 'linear-gradient(rgba(0, 0, 0, 0.7), rgba(0, 0, 0, 0.7)), url(https://images.unsplash.com/photo-1560518883-ce09059eeffa?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1773&q=80)',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          padding: '80px 0',
          color: 'white'
        }}
      >
        <Container size="xl">
          <Grid>
            <Grid.Col span={{ base: 12, md: 7 }}>
              <Title order={1} size={46} mb="md">Sistema de Gestión de Propiedades de Alquiler</Title>
              <Text size="xl" mb="xl">Administre fácilmente sus propiedades, inquilinos y contratos con nuestra plataforma completa. Optimizada para propietarios y administradores de propiedades.</Text>
              <Group>
                <Button component={Link} to="/login" size="lg" variant="filled" rightSection={<IconArrowRight size={18} />}>
                  Iniciar Sesión
                </Button>
                <Button component={Link} to="/contact" size="lg" variant="outline" color="white">
                  Contáctenos
                </Button>
              </Group>
            </Grid.Col>
          </Grid>
        </Container>
      </Box>

      {/* Features Section */}
      <Container size="xl" py={80}>
        <Title order={2} ta="center" mb={50}>Funcionalidades Principales</Title>
        
        <Grid>
          <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
            <Card withBorder p="xl" radius="md" h="100%">
              <ThemeIcon size={60} radius="md" mb="md">
                <IconHome size={30} />
              </ThemeIcon>
              <Title order={3} mb="sm">Gestión de Propiedades</Title>
              <Text>Mantenga un registro detallado de todas sus propiedades, incluyendo ubicación, características y mantenimiento.</Text>
            </Card>
          </Grid.Col>
          
          <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
            <Card withBorder p="xl" radius="md" h="100%">
              <ThemeIcon size={60} radius="md" mb="md" color="orange">
                <IconUsers size={30} />
              </ThemeIcon>
              <Title order={3} mb="sm">Gestión de Inquilinos</Title>
              <Text>Administre los perfiles de sus inquilinos, su historial y su información de contacto en un solo lugar.</Text>
            </Card>
          </Grid.Col>
          
          <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
            <Card withBorder p="xl" radius="md" h="100%">
              <ThemeIcon size={60} radius="md" mb="md" color="green">
                <IconCoin size={30} />
              </ThemeIcon>
              <Title order={3} mb="sm">Seguimiento de Pagos</Title>
              <Text>Registre y monitoree todos los pagos de alquiler y genere reportes detallados de ingresos.</Text>
            </Card>
          </Grid.Col>
          
          <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
            <Card withBorder p="xl" radius="md" h="100%">
              <ThemeIcon size={60} radius="md" mb="md" color="grape">
                <IconCheck size={30} />
              </ThemeIcon>
              <Title order={3} mb="sm">Contratos Digitales</Title>
              <Text>Cree, almacene y gestione contratos de manera digital con recordatorios automáticos de renovación.</Text>
            </Card>
          </Grid.Col>
        </Grid>
      </Container>

      {/* Benefits Section */}
      <Box style={{ backgroundColor: '#f8f9fa', padding: '80px 0' }}>
        <Container size="xl">
          <Grid align="center">
            <Grid.Col span={{ base: 12, md: 6 }} order={{ base: 2, md: 1 }}>
              <Title order={2} mb="xl">Beneficios para Propietarios</Title>
              
              <List
                spacing="md"
                size="lg"
                center
                icon={
                  <ThemeIcon color="blue" size={28} radius="xl">
                    <IconCheck size={18} />
                  </ThemeIcon>
                }
              >
                <List.Item>Ahorre tiempo con la gestión centralizada de propiedades</List.Item>
                <List.Item>Mejore la comunicación con los inquilinos</List.Item>
                <List.Item>Reduzca las vacantes con un seguimiento eficiente</List.Item>
                <List.Item>Obtenga análisis detallados de sus ingresos por alquileres</List.Item>
                <List.Item>Reciba alertas sobre contratos por vencer</List.Item>
              </List>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, md: 6 }} order={{ base: 1, md: 2 }}>
              <Image
                src="https://images.unsplash.com/photo-1556155092-490a1ba16284?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=2070&q=80"
                radius="md"
                alt="Beneficios para propietarios"
              />
            </Grid.Col>
          </Grid>
        </Container>
      </Box>

      {/* Call to Action */}
      <Container size="xl" py={80} ta="center">
        <Title order={2} mb="md">¿Listo para optimizar la gestión de sus propiedades?</Title>
        <Text size="xl" mb="xl" maw={700} mx="auto">
          Únase a cientos de propietarios que ya están gestionando sus propiedades de manera más eficiente con nuestra plataforma.
        </Text>
        <Button 
          component={Link} 
          to="/login" 
          size="xl" 
          rightSection={<IconArrowRight size={20} />}
        >
          Comenzar Ahora
        </Button>
      </Container>

      {/* Footer */}
      <Box style={{ backgroundColor: '#f1f1f1', padding: '50px 0' }}>
        <Container size="xl">
          <Grid>
            <Grid.Col span={{ base: 12, md: 4 }}>
              <Title order={3} mb="md">Rental Manager</Title>
              <Text mb="md">
                La solución completa para la gestión de propiedades de alquiler, diseñada para propietarios y administradores.
              </Text>
              <Group gap={10}>
                <Badge size="lg">Propiedades</Badge>
                <Badge size="lg">Inquilinos</Badge>
                <Badge size="lg">Contratos</Badge>
                <Badge size="lg">Pagos</Badge>
              </Group>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, md: 4 }}>
              <Title order={3} mb="md">Enlaces Rápidos</Title>
              <List spacing="sm">
                <List.Item>
                  <Link to="/" style={{ textDecoration: 'none', color: 'inherit' }}>Inicio</Link>
                </List.Item>
                <List.Item>
                  <Link to="/about" style={{ textDecoration: 'none', color: 'inherit' }}>Acerca de</Link>
                </List.Item>
                <List.Item>
                  <Link to="/contact" style={{ textDecoration: 'none', color: 'inherit' }}>Contacto</Link>
                </List.Item>
                <List.Item>
                  <Link to="/login" style={{ textDecoration: 'none', color: 'inherit' }}>Iniciar Sesión</Link>
                </List.Item>
              </List>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, md: 4 }}>
              <Title order={3} mb="md">Contacto</Title>
              <Text>Email: info@rentalmanager.com</Text>
              <Text>Teléfono: +57 300 123 4567</Text>
              <Text>Dirección: Calle 123 #45-67, Bogotá, Colombia</Text>
            </Grid.Col>
          </Grid>
          
          <Text ta="center" mt={50} c="dimmed">
            © {new Date().getFullYear()} Rental Manager. Todos los derechos reservados.
          </Text>
        </Container>
      </Box>
    </Box>
  );
} 