import { useState, useEffect } from 'react';
import { 
  Title, 
  Container, 
  Table, 
  ActionIcon, 
  Group, 
  Button, 
  TextInput,
  LoadingOverlay,
  Text,
  Paper
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconEdit, IconTrash, IconPlus, IconSearch, IconUserCircle } from '@tabler/icons-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { personApi } from '../api/apiService';
import type { Person } from '../types';
import { StableModal } from '../components/ui/StableModal';
import { useAuth } from '../contexts/AuthContext';

export default function Persons() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const isManager = user?.role === 'manager';
  const isStandardUser = !isAdmin && !isManager; 

  const [searchTerm, setSearchTerm] = useState('');
  const [selectedPerson, setSelectedPerson] = useState<Person | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isModalReadOnly, setIsModalReadOnly] = useState(false);
  const [formSubmitted, setFormSubmitted] = useState(false);
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!isModalOpen) {
      setFormSubmitted(false);
      setSelectedPerson(null);
      setIsModalReadOnly(false);
    }
  }, [isModalOpen]);

  const { 
    data: persons = [], 
    isLoading, 
    error 
  } = useQuery<Person[]>({
    queryKey: ['persons', user?.role, user?.person_id],
    queryFn: async () => {
      if (!user) return [];
      if (isAdmin || isManager) {
        return personApi.getAll();
      }
      if (user.person_id) { 
        try {
          const person = await personApi.getById(user.person_id);
          return person ? [person] : [];
        } catch (err) {
          console.error("Failed to fetch user's own person data:", err);
          notifications.show({
            title: 'Error',
            message: 'No se pudo cargar su información personal.',
            color: 'red'
          });
          return [];
        }
      }
      return [];
    },
    enabled: !!user,
  });

  const filteredPersons = persons.filter(person => 
    person && (
      person.full_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      person.phone?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      person.nit?.toLowerCase().includes(searchTerm.toLowerCase()) 
    )
  );

  const handleView = (person: Person) => {
    setSelectedPerson(person);
    setIsModalReadOnly(true);
    setFormSubmitted(false);
    setIsModalOpen(true);
  };
  
  const handleEdit = (person: Person) => {
    setSelectedPerson(person);
    setIsModalReadOnly(false);
    setFormSubmitted(false);
    setIsModalOpen(true);
  };

  const handleAdd = () => {
    setSelectedPerson({
      id: '',
      full_name: '',
      phone: '',
      nit: ''
    });
    setIsModalReadOnly(false);
    setFormSubmitted(false);
    setIsModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!isAdmin) {
        notifications.show({ title: 'Acceso Denegado', message: 'No tiene permisos para eliminar personas.', color: 'red'});
        return;
    }
    if (window.confirm('¿Está seguro que desea eliminar esta persona? Esta acción es irreversible.')) {
      try {
        await personApi.delete(id);
        queryClient.invalidateQueries({ queryKey: ['persons', user?.role, user?.person_id] });
        notifications.show({
          title: 'Éxito',
          message: 'Persona eliminada correctamente',
          color: 'green',
        });
      } catch (error) {
        notifications.show({
          title: 'Error',
          message: 'Error al eliminar la persona',
          color: 'red',
        });
      }
    }
  };

  const handleSavePerson = async () => {
    setFormSubmitted(true);
    
    if (!selectedPerson) {
      notifications.show({ title: 'Error', message: 'No hay datos de persona para guardar', color: 'red' });
      return;
    }

    if (!selectedPerson.full_name || !selectedPerson.phone || !selectedPerson.nit) {
      notifications.show({
        title: 'Error de Validación',
        message: 'Por favor complete todos los campos requeridos (Nombre, Teléfono, NIT).',
        color: 'red',
      });
      return;
    }

    if (!isAdmin && !isManager) {
        notifications.show({ title: 'Acceso Denegado', message: 'No tiene permisos para guardar personas.', color: 'red'});
        return;
    }

    try {
      if (selectedPerson.id) {
        await personApi.update(selectedPerson.id, selectedPerson);
        notifications.show({ title: 'Éxito', message: 'Persona actualizada correctamente', color: 'green'});
      } else {
        const { id, ...newPersonData } = selectedPerson;
        await personApi.create(newPersonData);
        notifications.show({ title: 'Éxito', message: 'Persona creada correctamente', color: 'green'});
      }
      queryClient.invalidateQueries({ queryKey: ['persons', user?.role, user?.person_id] });
      setIsModalOpen(false);
    } catch (error) {
      console.error('Error saving person:', error);
      notifications.show({ title: 'Error', message: 'Error al guardar la persona', color: 'red'});
    }
  };

  const canPerformActions = isAdmin || isManager;
  const canDelete = isAdmin;
  const canCreate = isAdmin || isManager;
  const showSearchBar = isAdmin || isManager;

  let modalTitle = "Ver Persona";
  if (canCreate && !selectedPerson?.id) {
    modalTitle = "Añadir Persona";
  } else if (canPerformActions && selectedPerson?.id && !isModalReadOnly) {
    modalTitle = "Editar Persona";
  }


  return (
    <Container size="xl">
      <LoadingOverlay visible={isLoading || !user} />
      {error && <Text color="red">Error al cargar datos de personas: {error instanceof Error ? error.message : 'Error desconocido'}</Text>}
      
      <Group justify="space-between" mb="md">
        <Title order={1}>Gestión de Personas</Title>
        {isAdmin && (
          <Button 
            leftSection={<IconPlus size={16} />} 
            onClick={handleAdd}
            data-testid="add-person-button"
          >
            Añadir Persona
          </Button>
        )}
      </Group>
      
      {showSearchBar && (
        <TextInput
          placeholder="Buscar personas por nombre, teléfono o NIT..."
          mb="md"
          leftSection={<IconSearch size={16} />}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.currentTarget.value)}
        />
      )}

      {isStandardUser && filteredPersons.length === 1 && (
        <Paper shadow="sm" p="lg" radius="md" withBorder mb="xl">
          <Group mb="sm">
            <IconUserCircle size={24} />
            <Title order={3}>Mi Información Personal</Title>
          </Group>
          <Text><strong>Nombre:</strong> {filteredPersons[0].full_name}</Text>
          <Text><strong>Teléfono:</strong> {filteredPersons[0].phone}</Text>
          <Text><strong>NIT:</strong> {filteredPersons[0].nit}</Text>
          <Button mt="md" onClick={() => handleView(filteredPersons[0])}>Ver Detalles</Button>
        </Paper>
      )}

      {(isAdmin || isManager) && (
        <Table striped highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Nombre</Table.Th>
              <Table.Th>Teléfono</Table.Th>
              <Table.Th>NIT</Table.Th>
              {canPerformActions && <Table.Th>Acciones</Table.Th>} 
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {filteredPersons.length === 0 && (isAdmin || isManager) ? (
              <Table.Tr>
                <Table.Td colSpan={canPerformActions ? 4 : 3} align="center">No se encontraron personas.</Table.Td>
              </Table.Tr>
            ) : (
              filteredPersons.map((person) => (
                <Table.Tr key={person.id}>
                  <Table.Td>{person.full_name}</Table.Td>
                  <Table.Td>{person.phone}</Table.Td>
                  <Table.Td>{person.nit}</Table.Td>
                  {canPerformActions && (
                    <Table.Td>
                      <Group gap="xs">
                        <ActionIcon variant="subtle" color="blue" onClick={() => handleEdit(person)} title="Editar Persona">
                          <IconEdit size={16} />
                        </ActionIcon>
                        {canDelete && (
                          <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(person.id)} title="Eliminar Persona">
                            <IconTrash size={16} />
                          </ActionIcon>
                        )}
                      </Group>
                    </Table.Td>
                  )}
                </Table.Tr>
              ))
            )}
          </Table.Tbody>
        </Table>
      )}
      
      {selectedPerson && (
        <StableModal 
          opened={isModalOpen} 
          onClose={() => setIsModalOpen(false)} 
          title={modalTitle}
          centered
          size="md"
          padding="xl"
          radius="md"
        >
          <TextInput
            label="Nombre Completo"
            placeholder="Ingrese nombre completo"
            required
            readOnly={isModalReadOnly}
            mb="md"
            value={selectedPerson.full_name || ''}
            onChange={(e) => setSelectedPerson({ ...selectedPerson, full_name: e.currentTarget.value })}
            error={formSubmitted && !selectedPerson.full_name ? "El nombre completo es requerido" : null}
          />
          <TextInput
            label="Teléfono"
            placeholder="Ingrese número de teléfono"
            required
            readOnly={isModalReadOnly}
            mb="md"
            value={selectedPerson.phone || ''}
            onChange={(e) => setSelectedPerson({ ...selectedPerson, phone: e.currentTarget.value })}
            error={formSubmitted && !selectedPerson.phone ? "El teléfono es requerido" : null}
          />
          <TextInput
            label="NIT"
            placeholder="Ingrese NIT"
            required
            readOnly={isModalReadOnly}
            mb="lg"
            value={selectedPerson.nit || ''}
            onChange={(e) => setSelectedPerson({ ...selectedPerson, nit: e.currentTarget.value })}
            error={formSubmitted && !selectedPerson.nit ? "El NIT es requerido" : null}
          />
          {!isModalReadOnly && (isAdmin || isManager) && (
            <Group justify="flex-end">
              <Button variant="default" onClick={() => setIsModalOpen(false)}>Cancelar</Button>
              <Button onClick={handleSavePerson} data-testid="save-person-button">Guardar</Button>
            </Group>
          )}
          {isModalReadOnly && (
             <Group justify="flex-end">
              <Button variant="default" onClick={() => setIsModalOpen(false)}>Cerrar</Button>
            </Group>
          )}
        </StableModal>
      )}
    </Container>
  );
} 