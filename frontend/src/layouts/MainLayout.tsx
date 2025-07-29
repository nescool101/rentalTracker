import { ReactNode, useState } from 'react';
import { Outlet, Link, useLocation } from 'react-router-dom';
import {
  AppShell,
  Burger,
  Group,
  NavLink,
  Text,
  Title,
  UnstyledButton,
  useMantineColorScheme,
  useComputedColorScheme,
  ActionIcon,
  Button
} from '@mantine/core';
import { IconHome, IconBuilding, IconUser, IconTools, IconReportMoney, IconNote, IconMoon, IconSun, IconHistory, IconLogout, IconUsers, IconCurrencyDollar, IconMail, IconFileTypePdf, IconUserPlus, IconCreditCard, IconCloudUpload, IconFileText } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { PendingActivationNotice } from '../components/PendingActivationNotice';

interface MainLayoutProps {
  children?: ReactNode;
}

export default function MainLayout({ children }: MainLayoutProps) {
  const [opened, setOpened] = useState(false);
  const location = useLocation();
  const { setColorScheme } = useMantineColorScheme();
  const computedColorScheme = useComputedColorScheme('light');
  const { user, logout } = useAuth();
  
  const isDarkMode = computedColorScheme === 'dark';
  const isPendingActivation = user?.status === 'activenopaid';

  const toggleColorScheme = () => {
    setColorScheme(isDarkMode ? 'light' : 'dark');
  };

  const navItemsBase = [
    { label: 'Panel Principal', icon: <IconHome size={20} stroke={1.5} />, to: '/dashboard' },
    { label: 'Propiedades', icon: <IconBuilding size={20} stroke={1.5} />, to: '/properties' },
    { label: 'Personas', icon: <IconUser size={20} stroke={1.5} />, to: '/persons' },
    { label: 'Mantenimiento', icon: <IconTools size={20} stroke={1.5} />, to: '/maintenance' },
    { label: 'Pagos', icon: <IconReportMoney size={20} stroke={1.5} />, to: '/payments' },
    { label: 'Historial de Alquiler', icon: <IconHistory size={20} stroke={1.5} />, to: '/rental-history' },
    { label: 'Contratos', icon: <IconNote size={20} stroke={1.5} />, to: '/contracts' },
  ];

  let navItems = [...navItemsBase];

  // Add file upload for all users
  navItems.push({ label: 'Subir Archivos', icon: <IconCloudUpload size={20} stroke={1.5} />, to: '/file-upload' });

  if (user?.role === 'admin') {
    navItems.push({ label: 'Gestión de Usuarios', icon: <IconUsers size={20} stroke={1.5} />, to: '/users' });
    navItems.push({ label: 'Precios de Alquileres', icon: <IconCurrencyDollar size={20} stroke={1.5} />, to: '/admin/rental-pricing' });
    navItems.push({ label: 'Envío Manual de Emails', icon: <IconMail size={20} stroke={1.5} />, to: '/admin/manual-email-sender' });
    navItems.push({ label: 'Generación de Contratos', icon: <IconFileTypePdf size={20} stroke={1.5} />, to: '/admin/contract-generation' });
    navItems.push({ label: 'Invitaciones a Managers', icon: <IconUserPlus size={20} stroke={1.5} />, to: '/admin/manager-invitations' });
    navItems.push({ label: 'Cuentas Bancarias', icon: <IconCreditCard size={20} stroke={1.5} />, to: '/admin/bank-accounts' });
    navItems.push({ label: 'Gestión de Archivos', icon: <IconCloudUpload size={20} stroke={1.5} />, to: '/admin/file-upload' });
    navItems.push({ label: 'Archivos Subidos', icon: <IconFileText size={20} stroke={1.5} />, to: '/admin/file-management' });

  }
  
  // Add profile link for all users
  navItems.push({ label: 'Mi Perfil', icon: <IconUser size={20} stroke={1.5} />, to: '/profile' });

  return (
    <AppShell
      header={{ height: { base: 56, sm: 60 } }}
      navbar={{ 
        width: { base: 280, sm: 300 }, 
        breakpoint: 'sm', 
        collapsed: { mobile: !opened, desktop: false } 
      }}
      padding={{ base: 'xs', sm: 'md' }}
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger
              opened={opened}
              onClick={() => setOpened((o) => !o)}
              hiddenFrom="sm"
              size="sm"
            />
            <UnstyledButton component={Link} to="/dashboard">
              <Title order={3}>Rental Manager</Title>
            </UnstyledButton>
          </Group>
          <Group>
            {user && (
              <Text size="sm" c="dimmed">
                {user.email}
              </Text>
            )}
            <ActionIcon 
              variant="default" 
              onClick={toggleColorScheme} 
              size="lg"
            >
              {isDarkMode ? <IconSun size={18} /> : <IconMoon size={18} />}
            </ActionIcon>
            <Button
              variant="subtle"
              color="red"
              leftSection={<IconLogout size={18} />}
              onClick={logout}
            >
              Salir
            </Button>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md" style={{ overflowY: 'auto', height: '100%' }}>
        {!isPendingActivation ? (
          <>
            <Text size="sm" fw={500} c="dimmed" mb="xs">Menú Principal</Text>
            <div style={{ 
              maxHeight: 'calc(100vh - 120px)', 
              overflowY: 'auto',
              paddingRight: '8px',
              marginRight: '-8px'
            }}>
              {navItems.map((item) => (
                <NavLink
                  key={item.label}
                  component={Link}
                  to={item.to}
                  label={item.label}
                  leftSection={item.icon}
                  active={location.pathname === item.to}
                  variant="filled"
                  mb={5}
                />
              ))}
            </div>
          </>
        ) : (
          <>
            <Text size="sm" fw={500} c="dimmed" mb="xs">Acceso Limitado</Text>
            <NavLink
              component={Link}
              to="/dashboard"
              label="Panel Principal"
              leftSection={<IconHome size={20} stroke={1.5} />}
              active={location.pathname === '/dashboard'}
              variant="filled"
              mb={5}
            />
            <NavLink
              component={Link}
              to="/profile"
              label="Mi Perfil"
              leftSection={<IconUser size={20} stroke={1.5} />}
              active={location.pathname === '/profile'}
              variant="filled"
              mb={5}
            />
          </>
        )}
      </AppShell.Navbar>

      <AppShell.Main>
        {isPendingActivation && location.pathname !== '/dashboard' && location.pathname !== '/profile' ? (
          <PendingActivationNotice />
        ) : (
          children || <Outlet />
        )}
      </AppShell.Main>
    </AppShell>
  );
} 