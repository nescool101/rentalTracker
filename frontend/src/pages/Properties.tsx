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
  Badge,
  Select,
  Text,
  Paper,
  Card,
  ThemeIcon,
  MultiSelect
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconEdit, IconTrash, IconPlus, IconSearch, IconHomeCheck, IconBuilding } from '@tabler/icons-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { propertyApi, personApi, userApi } from '../api/apiService';
import { useAuth } from '../contexts/AuthContext';
import type { Property, Person, User as AppUser } from '../types';
import { StableModal } from '../components/ui/StableModal';
import axios from 'axios';

export default function Properties() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedProperty, setSelectedProperty] = useState<Property | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [formSubmitted, setFormSubmitted] = useState(false);
  const [departments, setDepartments] = useState<{value: string, label: string}[]>([]);
  const [loadingDepartments, setLoadingDepartments] = useState(false);
  const queryClient = useQueryClient();
  const { user } = useAuth();

  const isAdmin = user?.role === 'admin';
  const isManager = user?.role === 'manager';
  const isStandardUser = !isAdmin && !isManager;

  useEffect(() => {
    if (!isModalOpen) {
      setFormSubmitted(false);
      setSelectedProperty(null);
    }
  }, [isModalOpen]);

  // Fetch departments from Colombia API
  useEffect(() => {
    const fetchDepartments = async () => {
      setLoadingDepartments(true);
      try {
        const response = await axios.get('https://api-colombia.com/api/v1/Department');
        const departmentOptions = response.data.map((dept: any) => ({
          value: dept.id.toString(),
          label: dept.name
        }));
        setDepartments(departmentOptions);
      } catch (error) {
        console.error('Error fetching departments:', error);
        notifications.show({
          title: 'Error',
          message: 'No se pudieron cargar los departamentos. Por favor, intenta nuevamente.',
          color: 'red',
        });
      } finally {
        setLoadingDepartments(false);
      }
    };

    fetchDepartments();
  }, []);

  const { 
    data: properties = [], 
    isLoading, 
    error,
  } = useQuery<Property[]>({
    queryKey: ['properties', user?.role, user?.person_id],
    queryFn: async () => {
      if (!user) return [];
      try {
        if (isAdmin || isManager) {
          return await propertyApi.getAll(); 
        } else if (user.person_id) {
          const residentProperties = await propertyApi.getByResidentId(user.person_id);
          return residentProperties || [];
        }
        return [];
      } catch (err) {
        console.error('Error fetching properties:', err);
        notifications.show({
          title: 'Error de Carga',
          message: 'No se pudieron cargar las propiedades.',
          color: 'red',
        });
        return [];
      }
    },
    enabled: !!user,
  });

  const { data: persons = [] } = useQuery<Person[]>({
    queryKey: ['persons', 'all', 'forPropertiesPageDropdowns'],
    queryFn: personApi.getAll,
    enabled: isAdmin || isManager,
  });

  const { data: allUsers = [] } = useQuery<AppUser[]>({
    queryKey: ['users', 'all', 'forPropertiesPageManagerDropdown'],
    queryFn: userApi.getAll,
    enabled: isAdmin,
  });

  const managerPersons = persons.filter(person => 
    allUsers.some(u => u.person_id === person.id && (u.role === 'manager' || u.role === 'admin'))
  );

  const filteredProperties = properties.filter(property => 
    property && (
      property.address?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      property.city?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      property.department?.toLowerCase().includes(searchTerm.toLowerCase())
    )
  );
  
  const handleEdit = (property: Property) => {
    setSelectedProperty({ ...property });
    setFormSubmitted(false);
    setIsModalOpen(true);
  };

  const handleAdd = () => {
    setSelectedProperty({
      id: '',
      address: '',
      apt_number: '',
      city: '',
      state: '',
      department: '',
      department_id: '',
      zip_code: '',
      type: '',
      resident_id: isManager && !isAdmin && user?.person_id ? user.person_id : '',
      manager_ids: isManager && !isAdmin && user?.person_id ? [user.person_id] : []
    });
    setFormSubmitted(false);
    setIsModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!isAdmin) {
      notifications.show({title: 'Acceso Denegado', message: 'No tiene permisos para eliminar propiedades.', color: 'red'});
      return;
    }
    if (window.confirm('¿Está seguro que desea eliminar esta propiedad?')) {
      try {
        await propertyApi.delete(id);
        queryClient.invalidateQueries({ queryKey: ['properties', user?.role, user?.person_id] });
        notifications.show({ title: 'Éxito', message: 'Propiedad eliminada correctamente', color: 'green' });
      } catch (err) {
        notifications.show({ title: 'Error', message: 'Error al eliminar la propiedad', color: 'red' });
      }
    }
  };

  const getPersonName = (id: string | undefined): string => {
    if (!id) return 'No asignado';
    const localPersons = persons || [];
    const localManagerPersons = managerPersons || [];
    const person = localPersons.find(p => p.id === id);
    return person ? person.full_name : (localManagerPersons.find(mp => mp.id === id)?.full_name || 'ID no encontrado');
  };

  const getManagerNames = (managerIds: string[] | undefined): string => {
    if (!managerIds || managerIds.length === 0) return 'No asignado';
    return managerIds.map(id => getPersonName(id)).join(', ');
  };

  const handleSaveProperty = async () => {
    setFormSubmitted(true);
    if (!selectedProperty) return;

    if (!selectedProperty.address || !selectedProperty.city || 
        !selectedProperty.department_id || !selectedProperty.zip_code || 
        !selectedProperty.type || 
        (isAdmin && (!selectedProperty.resident_id || !selectedProperty.manager_ids || selectedProperty.manager_ids.length === 0)) ||
        (isManager && !isAdmin && !selectedProperty.resident_id) ) {
      notifications.show({ title: 'Error de Validación', message: 'Por favor complete todos los campos requeridos.', color: 'red' });
      return;
    }
    
    const payload: Partial<Property> = { ...selectedProperty };

    if (isManager && !isAdmin && user?.person_id) {
        const currentManagerIds = payload.manager_ids ? [...payload.manager_ids] : [];
        if (!currentManagerIds.includes(user.person_id)) {
            currentManagerIds.push(user.person_id);
        }
        payload.manager_ids = currentManagerIds;
        
        if (!payload.resident_id || payload.resident_id === '') {
            payload.resident_id = user.person_id;
        }
    }

    try {
      if (payload.id) {
        await propertyApi.update(payload.id, payload as Property);
        notifications.show({ title: 'Éxito', message: 'Propiedad actualizada correctamente', color: 'green' });
      } else {
        const { id, ...newPropertyData } = payload;
        await propertyApi.create(newPropertyData as Omit<Property, 'id'>);
        notifications.show({ title: 'Éxito', message: 'Propiedad creada correctamente', color: 'green' });
      }
      queryClient.invalidateQueries({ queryKey: ['properties', user?.role, user?.person_id] });
      setIsModalOpen(false);
    } catch (err) {
      console.error('Error saving property:', err);
      notifications.show({ title: 'Error', message: 'Error al guardar la propiedad', color: 'red' });
    }
  };
  
  const canManageProperties = isAdmin || isManager;
  const modalTitle = selectedProperty?.id ? (canManageProperties ? "Editar Propiedad" : "Detalles de la Propiedad") : "Añadir Propiedad";

  return (
    <Container size="xl">
      <LoadingOverlay visible={isLoading && !error} />
      {error && <Text color="red" ta="center" py="xl">Error al cargar propiedades: {(error as Error).message}</Text>}

      <Group justify="space-between" mb="md">
        <Title order={1}>Propiedades</Title>
        {isAdmin && (
          <Button leftSection={<IconPlus size={16} />} onClick={handleAdd} data-testid="add-property-button">
            Añadir Propiedad
          </Button>
        )}
      </Group>

      {isManager && !isAdmin && (
        <Paper withBorder p="md" mb="lg" radius="md">
          <Group align="center" mb="xs">
            <ThemeIcon color="blue" size="md">
              <IconBuilding size={16} />
            </ThemeIcon>
            <Text fw={500}>Información para Encargados</Text>
          </Group>
          <Text size="sm" mb="md">Como encargado, usted puede registrar una propiedad durante el proceso de registro. El administrador puede asignarle múltiples propiedades para gestionar o cambiar su estado de pago para activar su propiedad.</Text>
          
          {properties.filter(p => p.manager_ids?.includes(user?.person_id || '')).length === 0 && (
            <Text size="sm" fs="italic" c="dimmed">
              Para registrar su propiedad, complete el proceso de registro inicial del sistema. Si ya lo ha completado, contacte al administrador.
            </Text>
          )}
        </Paper>
      )}

      {(isAdmin || isManager) && properties.length > 1 && (
        <TextInput
          placeholder="Buscar propiedades por dirección, ciudad o departamento..."
          mb="md"
          leftSection={<IconSearch size={16} />}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.currentTarget.value)}
        />
      )}

      {(isAdmin || isManager) && !isLoading && !error && (
        <Table striped highlightOnHover withTableBorder mt="lg">
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Dirección</Table.Th>
              <Table.Th>Ciudad</Table.Th>
              <Table.Th>Departamento</Table.Th>
              <Table.Th>Tipo</Table.Th>
              <Table.Th>Residente</Table.Th>
              <Table.Th>Encargado (Manager)</Table.Th>
              {canManageProperties && <Table.Th>Acciones</Table.Th>}
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {filteredProperties.length === 0 ? (
              <Table.Tr><Table.Td colSpan={canManageProperties ? 7 : 6} ta="center">No se encontraron propiedades.</Table.Td></Table.Tr>
            ) : (
              filteredProperties.map((property) => (
                <Table.Tr key={property.id}>
                  <Table.Td>{property.address}{property.apt_number ? `, ${property.apt_number}` : ''}</Table.Td>
                  <Table.Td>{property.city}</Table.Td>
                  <Table.Td>{property.department || 'No especificado'}</Table.Td>
                  <Table.Td><Badge>{property.type}</Badge></Table.Td>
                  <Table.Td>{getPersonName(property.resident_id)}</Table.Td>
                  <Table.Td>{getManagerNames(property.manager_ids)}</Table.Td>
                  {canManageProperties && (
                    <Table.Td>
                      <Group gap="xs">
                        <ActionIcon variant="subtle" color="blue" onClick={() => handleEdit(property)} title="Editar">
                          <IconEdit size={16} />
                        </ActionIcon>
                        {isAdmin && (
                          <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(property.id)} title="Eliminar">
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

      {isStandardUser && !isLoading && !error && (
        <>
          {filteredProperties.length === 1 ? (
            <Card withBorder radius="md" p="xl" shadow="sm" mt="lg">
                <Group mb="xs">
                    <ThemeIcon variant="light" size={36} radius="md">
                        <IconHomeCheck size={20} />
                    </ThemeIcon>
                    <Title order={3}>Mi Propiedad</Title>
                </Group>
              <Text><strong>Dirección:</strong> {filteredProperties[0].address}{filteredProperties[0].apt_number ? `, ${filteredProperties[0].apt_number}` : ''}</Text>
              <Text><strong>Ciudad:</strong> {filteredProperties[0].city}</Text>
              <Text><strong>Departamento:</strong> {filteredProperties[0].department || 'No especificado'}</Text>
              <Text><strong>Código Postal:</strong> {filteredProperties[0].zip_code}</Text>
              <Text><strong>Tipo:</strong> {filteredProperties[0].type}</Text>
              <Text><strong>Encargado:</strong> {getManagerNames(filteredProperties[0].manager_ids)}</Text>
            </Card>
          ) : (
            <Paper withBorder p="xl" radius="md" shadow="sm" mt="lg">
                <Group justify="center" mb="sm">
                    <ThemeIcon variant="light" color="gray" size={48} radius="xl">
                        <IconBuilding size={28}/>
                    </ThemeIcon>
                </Group>
              <Text ta="center" c="dimmed">No tiene una propiedad asignada actualmente o no se pudo cargar la información.</Text>
              <Text ta="center" c="dimmed" fz="xs">Si cree que esto es un error, por favor contacte con la administración.</Text>
            </Paper>
          )}
        </>
      )}

      {selectedProperty && (isModalOpen && canManageProperties) && (
        <StableModal 
          opened={isModalOpen} 
          onClose={() => setIsModalOpen(false)} 
          title={modalTitle}
          centered size="lg"
        >
          <TextInput label="Dirección" placeholder="Ingrese dirección" required mb="sm" 
            value={selectedProperty.address} 
            onChange={(e) => setSelectedProperty({...selectedProperty, address: e.target.value})}
            error={formSubmitted && !selectedProperty.address ? "Campo requerido" : null}
           />
          <TextInput label="Nº Apartamento (Opcional)" placeholder="Ej: Apto 101" mb="sm" 
            value={selectedProperty.apt_number || ''} 
            onChange={(e) => setSelectedProperty({...selectedProperty, apt_number: e.target.value})}
          />
          <Group grow mb="sm">
            <Select 
              label="Departamento" 
              placeholder={loadingDepartments ? "Cargando departamentos..." : "Seleccione el departamento"}
              data={departments}
              disabled={loadingDepartments}
              searchable
              required
              value={selectedProperty.department_id || ''} 
              onChange={(value) => {
                const selectedDept = departments.find(dept => dept.value === value);
                setSelectedProperty({
                  ...selectedProperty, 
                  department_id: value || '',
                  department: selectedDept ? selectedDept.label : ''
                });
              }}
              error={formSubmitted && !selectedProperty.department_id ? "Campo requerido" : null}
            />
            <TextInput label="Ciudad" placeholder="Ingrese ciudad" required 
              value={selectedProperty.city} 
              onChange={(e) => setSelectedProperty({...selectedProperty, city: e.target.value})}
              error={formSubmitted && !selectedProperty.city ? "Campo requerido" : null}
            />
          </Group>
          <Group grow mb="sm">
            <TextInput label="Código Postal" placeholder="Ingrese código postal" required 
              value={selectedProperty.zip_code} 
              onChange={(e) => setSelectedProperty({...selectedProperty, zip_code: e.target.value})}
              error={formSubmitted && !selectedProperty.zip_code ? "Campo requerido" : null}
            />
            <Select label="Tipo de Propiedad" placeholder="Seleccione tipo" required 
              data={['Apartamento', 'Casa', 'Condominio', 'Comercial', 'Otro']} 
              value={selectedProperty.type} 
              onChange={(value) => setSelectedProperty({...selectedProperty, type: value || ''})}
              error={formSubmitted && !selectedProperty.type ? "Campo requerido" : null}
            />
          </Group>
          
          {isAdmin && (
              <>
                <Select label="Residente Asignado" placeholder="Seleccione residente" clearable mb="sm" searchable 
                data={persons.map(p => ({ value: p.id, label: p.full_name }))} 
                value={selectedProperty.resident_id || ''} 
                onChange={(value) => setSelectedProperty({...selectedProperty, resident_id: value || ''})} 
                error={formSubmitted && !selectedProperty.resident_id ? "Debe asignar un residente" : null}
                />
                <MultiSelect
                  label="Encargado(s) (Manager)"
                  placeholder="Seleccione encargado(s)"
                  clearable
                  searchable
                  mb="lg"
                  data={managerPersons.map(p => ({ value: p.id, label: p.full_name }))}
                  value={selectedProperty.manager_ids || []}
                  onChange={(value) => setSelectedProperty({...selectedProperty, manager_ids: value || []})}
                  error={formSubmitted && isAdmin && (!selectedProperty.manager_ids || selectedProperty.manager_ids.length === 0) ? "Debe asignar al menos un encargado" : null}
                />
            </>
          )}
          {isManager && !isAdmin && (
              <>
                <Select label="Residente Asignado" placeholder="Seleccione residente" clearable mb="sm" searchable 
                data={persons.map(p => ({ value: p.id, label: p.full_name }))}
                value={selectedProperty.resident_id || ''} 
                onChange={(value) => setSelectedProperty({...selectedProperty, resident_id: value || ''})} 
                error={formSubmitted && !selectedProperty.resident_id ? "Debe asignar un residente" : null}
                description="De forma temporal, usted puede ser asignado como residente hasta que se cree el inquilino. Si no selecciona a nadie, se usará su información automáticamente."
                />
                <TextInput 
                    label="Encargado(s) (Manager)" 
                    disabled 
                    value={getManagerNames(selectedProperty.manager_ids)} 
                    mb="lg" 
                    description={selectedProperty.id ? "Para cambiar encargados, contacte a un administrador." : "Usted será asignado como encargado."}
                />
            </>
          )}

          <Group justify="flex-end">
            <Button variant="default" onClick={() => setIsModalOpen(false)}>Cancelar</Button>
            <Button onClick={handleSaveProperty}>Guardar Propiedad</Button>
          </Group>
        </StableModal>
      )}
    </Container>
  );
} 