import { ReactNode } from 'react';
import { Outlet, Link, useLocation } from 'react-router-dom';
import {
  AppShell,
  Group,
  Text,
  Title,
  UnstyledButton,
  Button,
  ActionIcon,
  useComputedColorScheme,
  useMantineColorScheme,
  Burger,
  Divider,
  Box
} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { IconMoon, IconSun } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';

interface PublicLayoutProps {
  children?: ReactNode;
}

export default function PublicLayout({ children }: PublicLayoutProps) {
  const [opened, { toggle }] = useDisclosure(false);
  const location = useLocation();
  const { setColorScheme } = useMantineColorScheme();
  const computedColorScheme = useComputedColorScheme('light');
  const { isAuthenticated } = useAuth();
  
  const isDarkMode = computedColorScheme === 'dark';

  const toggleColorScheme = () => {
    setColorScheme(isDarkMode ? 'light' : 'dark');
  };

  const navItems = [
    { label: 'Inicio', to: '/' },
    { label: 'Acerca de', to: '/about' },
    { label: 'Contacto', to: '/contact' },
  ];

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 300, breakpoint: 'sm', collapsed: { desktop: true, mobile: !opened } }}
      padding={0}
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger
              opened={opened}
              onClick={toggle}
              hiddenFrom="sm"
              size="sm"
            />
            <UnstyledButton component={Link} to="/">
              <Title order={3}>Rental Manager</Title>
            </UnstyledButton>
          </Group>
          
          <Group>
            <Group gap="xs" visibleFrom="sm">
              {navItems.map((item) => (
                <UnstyledButton
                  key={item.label}
                  component={Link}
                  to={item.to}
                  style={{
                    fontWeight: location.pathname === item.to ? 'bold' : 'normal',
                    padding: '8px 12px',
                  }}
                >
                  <Text>{item.label}</Text>
                </UnstyledButton>
              ))}
            </Group>
            
            <ActionIcon 
              variant="default" 
              onClick={toggleColorScheme} 
              size="lg"
            >
              {isDarkMode ? <IconSun size={18} /> : <IconMoon size={18} />}
            </ActionIcon>
            
            {isAuthenticated ? (
              <Button component={Link} to="/dashboard">
                Dashboard
              </Button>
            ) : (
              <Button component={Link} to="/login">
                Iniciar Sesión
              </Button>
            )}
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Title order={4} mb="md">Menú</Title>
        {navItems.map((item) => (
          <UnstyledButton
            key={item.label}
            component={Link}
            to={item.to}
            style={{
              padding: '12px 0',
              width: '100%',
              display: 'block',
              fontWeight: location.pathname === item.to ? 'bold' : 'normal',
            }}
          >
            {item.label}
          </UnstyledButton>
        ))}
        
        <Divider my="md" />
        
        <Button component={Link} to="/login" fullWidth>
          Iniciar Sesión
        </Button>
      </AppShell.Navbar>

      <AppShell.Main>
        {children || <Outlet />}
        
        {/* Footer for Public Pages */}
        <Box style={{ backgroundColor: '#f1f1f1', padding: '20px 0' }}>
          <Group justify="center">
            <Text size="sm" c="dimmed">
              © {new Date().getFullYear()} Rental Manager. Todos los derechos reservados.
            </Text>
          </Group>
        </Box>
      </AppShell.Main>
    </AppShell>
  );
} 