import { Container, Title, Text, Timeline, Group, Avatar, Paper, Grid, Box, Divider } from '@mantine/core';
import { IconBuildingSkyscraper, IconUserCheck, IconCertificate } from '@tabler/icons-react';

export default function About() {
  const teamMembers = [
    {
      name: 'Nestor Alvarez',
      role: 'Fundador & CEO',
      bio: 'Con más de 15 años de experiencia en administración de propiedades, Nestor fundó Rental Manager para optimizar los procesos de gestión inmobiliaria.',
      avatar: 'https://images.unsplash.com/photo-1568602471122-7832951cc4c5?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=870&q=80'
    },
    {
      name: 'María González',
      role: 'Directora de Operaciones',
      bio: 'María tiene amplia experiencia en el sector inmobiliario y se asegura de que todos los aspectos operativos de la plataforma funcionen sin problemas.',
      avatar: 'https://images.unsplash.com/photo-1601412436009-d964bd02edbc?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=928&q=80'
    },
    {
      name: 'Carlos Martínez',
      role: 'Jefe de Tecnología',
      bio: 'Carlos lidera el equipo de desarrollo y asegura que la plataforma incorpore las últimas tecnologías para una gestión inmobiliaria eficiente.',
      avatar: 'https://images.unsplash.com/photo-1603415526960-f7e0328c63b1?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=870&q=80'
    },
  ];

  return (
    <Box>
      {/* Hero Section */}
      <Box 
        style={{ 
          backgroundColor: '#f8f9fa',
          padding: '80px 0 60px',
        }}
      >
        <Container size="xl">
          <Title order={1} mb="xl" ta="center">Acerca de Rental Manager</Title>
          <Text size="lg" mb="xl" ta="center" maw={800} mx="auto">
            Somos una empresa dedicada a simplificar la administración de propiedades de alquiler mediante una plataforma tecnológica integral diseñada para propietarios y administradores inmobiliarios.
          </Text>
        </Container>
      </Box>

      {/* Our Story Section */}
      <Container size="xl" py={60}>
        <Grid gutter={50} align="center">
          <Grid.Col span={{ base: 12, md: 6 }}>
            <Title order={2} mb="md">Nuestra Historia</Title>
            <Text mb="md">
              Rental Manager nació en 2019 como respuesta a las necesidades de los propietarios de bienes raíces que buscaban una solución moderna para gestionar sus propiedades de alquiler.
            </Text>
            <Text mb="md">
              Fundada por Nestor Alvarez después de años de enfrentar los retos de la administración inmobiliaria, nuestra plataforma ha evolucionado para convertirse en una herramienta completa que simplifica todos los aspectos de la gestión de alquileres.
            </Text>
            <Text>
              Hoy, servimos a cientos de propietarios y administradores de propiedades en toda Colombia, ayudándoles a optimizar sus procesos, reducir vacantes y aumentar la satisfacción de sus inquilinos.
            </Text>
          </Grid.Col>
          
          <Grid.Col span={{ base: 12, md: 6 }}>
            <Timeline active={3} bulletSize={24} lineWidth={2}>
              <Timeline.Item title="Fundación" bullet={<IconBuildingSkyscraper size={12} />}>
                <Text c="dimmed" size="sm">2019</Text>
                <Text size="sm">Lanzamiento de la primera versión de Rental Manager con funcionalidades básicas de administración de propiedades</Text>
              </Timeline.Item>

              <Timeline.Item title="Expansión" bullet={<IconUserCheck size={12} />}>
                <Text c="dimmed" size="sm">2020</Text>
                <Text size="sm">Incorporación de módulos para gestión de contratos digitales y seguimiento de pagos</Text>
              </Timeline.Item>

              <Timeline.Item title="Reconocimiento" bullet={<IconCertificate size={12} />}>
                <Text c="dimmed" size="sm">2022</Text>
                <Text size="sm">Premio a la innovación en tecnología inmobiliaria por la Asociación Colombiana de Bienes Raíces</Text>
              </Timeline.Item>

              <Timeline.Item title="Presente">
                <Text c="dimmed" size="sm">2025</Text>
                <Text size="sm">Una plataforma integral con módulos avanzados de análisis, reportes y comunicación con inquilinos</Text>
              </Timeline.Item>
            </Timeline>
          </Grid.Col>
        </Grid>
      </Container>

      <Divider />

      {/* Our Team Section */}
      <Container size="xl" py={60}>
        <Title order={2} ta="center" mb={50}>Nuestro Equipo</Title>
        
        <Grid>
          {teamMembers.map((member, index) => (
            <Grid.Col key={index} span={{ base: 12, sm: 6, md: 4 }}>
              <Paper p="xl" radius="md" withBorder style={{ height: '100%' }}>
                <Group mb="md">
                  <Avatar 
                    src={member.avatar}
                    size={80} 
                    radius="xl" 
                  />
                  <div>
                    <Text fw={700} size="lg">{member.name}</Text>
                    <Text c="dimmed">{member.role}</Text>
                  </div>
                </Group>
                <Text>{member.bio}</Text>
              </Paper>
            </Grid.Col>
          ))}
        </Grid>
      </Container>

      {/* Our Values Section */}
      <Box style={{ backgroundColor: '#f8f9fa', padding: '60px 0' }}>
        <Container size="xl">
          <Title order={2} ta="center" mb={50}>Nuestros Valores</Title>
          
          <Grid>
            <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
              <Paper p="xl" radius="md" withBorder h="100%" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Title order={3} mb="sm" ta="center">Innovación</Title>
                <Text ta="center">Constantemente buscamos nuevas tecnologías y métodos para mejorar la gestión inmobiliaria.</Text>
              </Paper>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
              <Paper p="xl" radius="md" withBorder h="100%" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Title order={3} mb="sm" ta="center">Transparencia</Title>
                <Text ta="center">Creemos en la comunicación clara y abierta con nuestros clientes e inquilinos.</Text>
              </Paper>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
              <Paper p="xl" radius="md" withBorder h="100%" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Title order={3} mb="sm" ta="center">Eficiencia</Title>
                <Text ta="center">Optimizamos cada proceso para ahorrar tiempo y recursos a nuestros usuarios.</Text>
              </Paper>
            </Grid.Col>
            
            <Grid.Col span={{ base: 12, sm: 6, md: 3 }}>
              <Paper p="xl" radius="md" withBorder h="100%" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Title order={3} mb="sm" ta="center">Confiabilidad</Title>
                <Text ta="center">Construimos sistemas robustos en los que propietarios e inquilinos pueden confiar.</Text>
              </Paper>
            </Grid.Col>
          </Grid>
        </Container>
      </Box>
    </Box>
  );
} 