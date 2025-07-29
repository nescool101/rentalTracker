import { 
  Title, 
  Container, 
  Text, 
  Paper, 
  Group, 
  ThemeIcon,
  Button,
  Badge,
  Table,
  ActionIcon,
  TextInput,
  Select,
  PasswordInput,
  LoadingOverlay
} from '@mantine/core';
import { IconUsers, IconEdit, IconTrash, IconPlus } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { useEffect, useState } from 'react';
import { userApi, personApi } from '../api/apiService';
import { notifications } from '@mantine/notifications';
import { User, Person } from '../types';
import { StableModal } from '../components/ui/StableModal';

export default function Users() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [users, setUsers] = useState<User[]>([]);
  const [persons, setPersons] = useState<Person[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<Partial<User> | null>(null);
  const [plainTextPassword, setPlainTextPassword] = useState<string>('');

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const data = await userApi.getAll();
      setUsers(data);
    } catch (error) {
      console.error('Failed to fetch users:', error);
      notifications.show({
        title: 'Error',
        message: 'Error al cargar datos de usuarios',
        color: 'red'
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchPersons = async () => {
    try {
      const data = await personApi.getAll();
      setPersons(data);
    } catch (error) {
      console.error('Failed to fetch persons:', error);
    }
  };

  useEffect(() => {
    fetchUsers();
    fetchPersons();
  }, []);

  const handleEdit = (user: User) => {
    setCurrentUser({
      ...user,
      // Don't show existing password
    });
    setPlainTextPassword('');
    setIsModalOpen(true);
  };

  const handleCreate = () => {
    setCurrentUser({
      role: 'user',
      email: '',
      person_id: ''
    });
    setPlainTextPassword('');
    setIsModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (confirm('¿Está seguro que desea eliminar este usuario?')) {
      setLoading(true);
      try {
        await userApi.delete(id);
        notifications.show({
          title: 'Éxito',
          message: 'Usuario eliminado exitosamente',
          color: 'green'
        });
        fetchUsers();
      } catch (error) {
        console.error('Failed to delete user:', error);
        notifications.show({
          title: 'Error',
          message: 'Error al eliminar usuario',
          color: 'red'
        });
      } finally {
        setLoading(false);
      }
    }
  };

  const handleSubmit = async () => {
    if (!currentUser || !currentUser.email || (!currentUser.id && !plainTextPassword)) {
      notifications.show({
        title: 'Error',
        message: 'Email y contraseña (para nuevos usuarios) son requeridos.',
        color: 'red'
      });
      return;
    }

    setLoading(true);
    try {
      if (currentUser.id) {
        // Update existing user
        const dataToUpdate: {
          email: string;
          role: string;
          person_id?: string;
          password_base64?: string; // Optional: only if password is being changed
        } = {
          email: currentUser.email!, // Non-null assertion as email is required for existing user
          role: currentUser.role!,   // Non-null assertion as role is required for existing user
        };

        if (currentUser.person_id) {
          dataToUpdate.person_id = currentUser.person_id;
        }

        if (plainTextPassword) { // If a new password was entered in the form
          if (plainTextPassword.length < 8) {
            notifications.show({ title: 'Error', message: 'La nueva contraseña debe tener al menos 8 caracteres.', color: 'red' });
            setLoading(false);
            return;
          }
          dataToUpdate.password_base64 = btoa(plainTextPassword); // Encode and add to payload
        }
        // If plainTextPassword is empty, password_base64 is NOT added to dataToUpdate.
        // This implies to the backend that the password should not be changed.

        await userApi.update(currentUser.id, dataToUpdate);
        notifications.show({
          title: 'Éxito',
          message: 'Usuario actualizado exitosamente',
          color: 'green'
        });
      } else {
        // Create new user
        if (!plainTextPassword || plainTextPassword.length < 8) {
            notifications.show({ title: 'Error', message: 'La contraseña es requerida y debe tener al menos 8 caracteres para nuevos usuarios.', color: 'red' });
            setLoading(false);
            return;
        }
        const userData: Omit<User, 'id'> & { password_base64: string } = {
            email: currentUser.email!,
            role: currentUser.role!,
            password_base64: btoa(plainTextPassword), // Encode for creation
        };
        if (currentUser.person_id) {
            userData.person_id = currentUser.person_id;
        }
        await userApi.create(userData);
        notifications.show({
          title: 'Éxito',
          message: 'Usuario creado exitosamente',
          color: 'green'
        });
      }
      setIsModalOpen(false);
      fetchUsers();
    } catch (error) {
      console.error('Failed to save user:', error);
      notifications.show({
        title: 'Error',
        message: 'Error al guardar usuario',
        color: 'red'
      });
    } finally {
      setLoading(false);
    }
  };

  // Get person name by ID
  const getPersonName = (personId: string) => {
    const person = persons.find(p => p.id === personId);
    return person ? person.full_name : 'No asignado';
  };

  // Get role badge color
  const getRoleBadgeColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return 'red';
      case 'manager':
        return 'blue';
      case 'user':
        return 'green';
      default:
        return 'gray';
    }
  };

  // Translate role names
  const translateRole = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return 'Administrador';
      case 'manager':
        return 'Rentista';
      case 'user':
        return 'Usuario';
      default:
        return role;
    }
  };

  return (
    <Container size="xl">
      <LoadingOverlay visible={loading} overlayProps={{ blur: 2 }} />
      
      <Group justify="space-between" mb="xl">
        <Title order={1}>Gestión de Usuarios</Title>
        {isAdmin && (
          <Button 
            leftSection={<IconPlus size={16} />} 
            onClick={handleCreate}
            data-testid="add-user-button"
          >
            Agregar Usuario
          </Button>
        )}
      </Group>
      
      <Paper shadow="sm" p="lg" radius="md" withBorder mb="xl">
        <Group mb="lg">
          <ThemeIcon size="xl" color="indigo" radius="md">
            <IconUsers size={24} />
          </ThemeIcon>
          <Title order={2}>Usuarios</Title>
        </Group>
        
        {users.length === 0 ? (
          <Text c="dimmed" ta="center" py="xl">
            No se encontraron usuarios. {isAdmin && 'Use el botón Agregar Usuario para crear uno.'}
          </Text>
        ) : (
          <Table striped highlightOnHover withTableBorder>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>ID</Table.Th>
                <Table.Th>Email</Table.Th>
                <Table.Th>Rol</Table.Th>
                <Table.Th>Persona</Table.Th>
                {isAdmin && <Table.Th>Acciones</Table.Th>}
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {users.map(user => (
                <Table.Tr key={user.id}>
                  <Table.Td>{user.id.substring(0, 8)}...</Table.Td>
                  <Table.Td>{user.email}</Table.Td>
                  <Table.Td>
                    <Badge color={getRoleBadgeColor(user.role)}>
                      {translateRole(user.role)}
                    </Badge>
                  </Table.Td>
                  <Table.Td>{getPersonName(user.person_id || '')}</Table.Td>
                  {isAdmin && (
                    <Table.Td>
                      <Group gap="xs">
                        <ActionIcon 
                          variant="subtle" 
                          color="blue"
                          onClick={() => handleEdit(user)}
                        >
                          <IconEdit size={16} />
                        </ActionIcon>
                        <ActionIcon 
                          variant="subtle" 
                          color="red"
                          onClick={() => handleDelete(user.id)}
                        >
                          <IconTrash size={16} />
                        </ActionIcon>
                      </Group>
                    </Table.Td>
                  )}
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Paper>

      {/* User Form Modal */}
      <StableModal 
        opened={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        title={currentUser?.id ? "Editar Usuario" : "Agregar Usuario"}
        centered
        size="md"
        padding="xl"
        radius="md"
        transitionProps={{ transition: 'fade', duration: 300 }}
        overlayProps={{ blur: 8 }}
        styles={{
          header: { 
            backgroundColor: "#f8f9fa", 
            borderBottom: "1px solid #e9ecef", 
            padding: "16px",
            borderTopLeftRadius: "8px",
            borderTopRightRadius: "8px"
          },
          body: { 
            padding: "20px" 
          }
        }}
      >
        <TextInput
          label="Email"
          placeholder="Ingrese email"
          required
          mb="md"
          value={currentUser?.email || ''}
          onChange={(event) => {
            if (currentUser) {
              setCurrentUser({
                ...currentUser,
                email: event.currentTarget.value
              });
            }
          }}
          error={!currentUser?.email ? "Email es requerido" : null}
        />
        
        <PasswordInput
          label={`${currentUser?.id ? 'Nueva ' : ''}Contraseña`}
          placeholder={currentUser?.id ? "Dejar en blanco para mantener la contraseña actual" : "Ingrese contraseña"}
          required={!currentUser?.id}
          mb="md"
          value={plainTextPassword}
          onChange={(event) => {
            setPlainTextPassword(event.currentTarget.value);
          }}
        />
        
        <Select
          label="Rol"
          placeholder="Seleccione rol"
          required
          mb="md"
          data={[
            { value: 'admin', label: 'Administrador' },
            { value: 'manager', label: 'Rentista' },
            { value: 'user', label: 'Usuario' }
          ]}
          value={currentUser?.role || ''}
          onChange={(value) => {
            if (currentUser) {
              setCurrentUser({
                ...currentUser,
                role: value || 'user'
              });
            }
          }}
        />
        
        <Select
          label="Persona"
          placeholder="Vincular a una persona (opcional)"
          mb="xl"
          clearable
          data={persons.map(person => ({
            value: person.id,
            label: person.full_name
          }))}
          value={currentUser?.person_id || ''}
          onChange={(value) => {
            if (currentUser) {
              setCurrentUser({
                ...currentUser,
                person_id: value || ''
              });
            }
          }}
        />
        
        <Group justify="flex-end">
          <Button variant="outline" onClick={() => setIsModalOpen(false)}>Cancelar</Button>
          <Button 
            onClick={(e) => {
              e.preventDefault();
              handleSubmit();
            }}
            data-testid="save-user-button"
          >
            Guardar
          </Button>
        </Group>
      </StableModal>
    </Container>
  );
} 