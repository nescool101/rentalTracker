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
  Textarea,
  Text
} from '@mantine/core';
import { DateInput } from '@mantine/dates';
import { notifications } from '@mantine/notifications';
import { IconEdit, IconTrash, IconPlus, IconSearch, IconRefresh } from '@tabler/icons-react';
import { useQuery } from '@tanstack/react-query';
import { propertyApi, personApi, maintenanceRequestApi, rentalApi } from '../api/apiService';
import { useAuth } from '../contexts/AuthContext';
import type { MaintenanceRequest, Property, Person, User } from '../types';
import { StableModal } from '../components/ui/StableModal';

export default function MaintenanceRequests() {
  const { user } = useAuth();

  const isAdmin = user?.role === 'admin';
  const isManager = user?.role === 'manager';
  const isResident = user?.role === 'resident' || user?.role === 'user';

  const [searchTerm, setSearchTerm] = useState('');
  const [selectedRequest, setSelectedRequest] = useState<Partial<MaintenanceRequest> | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // State for properties and persons dropdowns, filtered by role
  const [modalProperties, setModalProperties] = useState<Property[]>([]);
  const [modalPersons, setModalPersons] = useState<Person[]>([]);

  useEffect(() => {
    if (!isModalOpen) {
      setSelectedRequest(null);
    }
  }, [isModalOpen]);

  const { 
    data: requests = [], 
    isLoading: isLoadingRequests,
    error: requestsError,
    refetch: refetchRequests
  } = useQuery<MaintenanceRequest[]>({
    queryKey: ['maintenance-requests', user?.id, user?.role, user?.person_id],
    queryFn: () => maintenanceRequestApi.getForCurrentUser(user as User), // Cast user as User
    enabled: !!user,
  });

  // Fetch all properties and persons for Admin/Manager for dropdowns initially
  // Residents will have their specific property/person set when opening the modal
  const { data: allPropertiesGlobal = [] } = useQuery<Property[]>({
    queryKey: ['allPropertiesGlobal'],
    queryFn: propertyApi.getAll,
    enabled: isAdmin || isManager, 
  });

  const { data: allPersonsGlobal = [] } = useQuery<Person[]>({
    queryKey: ['allPersonsGlobal'],
    queryFn: personApi.getAll,
    enabled: isAdmin || isManager,
  });

  // Prepare data for modal dropdowns based on role
  useEffect(() => {
    if (!user) return;
    if (isAdmin) {
      setModalProperties(allPropertiesGlobal);
      setModalPersons(allPersonsGlobal);
    }
    // For manager/resident, this will be further refined when opening the modal (handleAdd/handleEdit)
  }, [user, isAdmin, allPropertiesGlobal, allPersonsGlobal]);


  const filteredRequests = requests?.filter(request => 
    request && request.description && 
    request.description.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  const canEditRequest = (request: MaintenanceRequest): boolean => {
    if (!user) return false;
    if (isAdmin) return true;
    if (isManager) {
      // Check if request.property_id is among manager's properties
      // This requires fetching manager's properties, could be optimized
      return modalProperties.some(p => p.id === request.property_id);
    }
    if (isResident) {
      return request.renter_id === user.person_id; // Can edit own requests (description typically)
    }
    return false;
  };

  const canChangeStatus = (request: MaintenanceRequest): boolean => {
    if (!user) return false;
    return isAdmin || (isManager && modalProperties.some(p => p.id === request.property_id));
  };

  const canDeleteRequest = (request: MaintenanceRequest): boolean => {
    if (!user) return false;
    if (isAdmin) return true;
    if (isManager) {
      return modalProperties.some(p => p.id === request.property_id);
    }
    if (isResident) {
      return request.renter_id === user.person_id;
    }
    return false;
  };

  const handleEdit = async (request: MaintenanceRequest) => {
    if (!user) return;
    setSelectedRequest({...request});

    if (isAdmin) {
      setModalProperties(allPropertiesGlobal);
      setModalPersons(allPersonsGlobal);
    } else if (isManager && user.person_id) {
      const managerProps = await propertyApi.getByManagerId(user.person_id);
      setModalProperties(managerProps);
      // For managers, renters could be from any of their properties
      // This might need a more specific fetch if too many persons
      setModalPersons(allPersonsGlobal); 
    } else if (isResident && user.person_id) {
      const currentProp = allPropertiesGlobal.find(p => p.id === request.property_id);
      setModalProperties(currentProp ? [currentProp] : []);
      const self = allPersonsGlobal.find(p => p.id === user.person_id);
      setModalPersons(self ? [self] : []);
    }
    setIsModalOpen(true);
  };

  const handleAdd = async () => {
    if (!user || !user.person_id) {
        notifications.show({title: 'Información Requerida', message: 'Su ID de persona no está configurado. No puede crear solicitudes.', color: 'orange'});
        return;
    }
    let defaultPropertyId = '';
    let defaultRenterId = user.person_id;
    let propertiesForModal: Property[] = [];
    let personsForModal: Person[] = [];

    if (isAdmin) {
      propertiesForModal = allPropertiesGlobal;
      personsForModal = allPersonsGlobal;
    } else if (isManager && user.person_id) {
      const managerProps = await propertyApi.getByManagerId(user.person_id);
      propertiesForModal = managerProps;
      if (managerProps.length > 0) defaultPropertyId = managerProps[0].id;
      personsForModal = allPersonsGlobal;
    } else if (isResident && user.person_id) {
      try {
        const rentals = await rentalApi.getByRenterId(user.person_id) || [];
        let associatedProperty: Property | undefined;
        if (rentals.length > 0) {
          const activeRental = rentals.find(r => new Date(r.end_date) > new Date());
          const propId = activeRental ? activeRental.property_id : (rentals[0]?.property_id);
          if (propId) {
            associatedProperty = allPropertiesGlobal.find(p => p.id === propId);
          }
        }
        if(!associatedProperty){
            const residentProps = await propertyApi.getByResidentId(user.person_id);
            if(residentProps && residentProps.length > 0) associatedProperty = residentProps[0];
        }

        if (associatedProperty) {
          propertiesForModal = [associatedProperty];
          defaultPropertyId = associatedProperty.id;
        } else {
            notifications.show({title: 'Propiedad no encontrada', message: 'No se encontró una propiedad asociada para crear la solicitud.', color: 'orange'});
        }
        const self = allPersonsGlobal.find(p => p.id === user.person_id);
        personsForModal = self ? [self] : [];
      } catch (e) { 
          console.error("Error fetching resident data for add request:", e);
          notifications.show({title: 'Error de Datos', message: 'No se pudo cargar su información de propiedad/inquilino.', color: 'red'});
      }
    }
    setModalProperties(propertiesForModal);
    setModalPersons(personsForModal);

    setSelectedRequest({
      id: '',
      property_id: defaultPropertyId, 
      renter_id: defaultRenterId,
      description: '',
      request_date: new Date().toISOString(),
      status: 'pending'
    });
    setIsModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    // Authorization check should happen at API level, but good for UI too
    const requestToDelete = requests.find(r => r.id === id);
    if (!requestToDelete || !canDeleteRequest(requestToDelete)) {
        notifications.show({title: 'Error', message: 'No autorizado para eliminar.', color: 'red'});
        return;
    }
    try {
      await maintenanceRequestApi.delete(id);
      refetchRequests();
      notifications.show({ title: 'Éxito', message: 'Solicitud eliminada.', color: 'green' });
    } catch (error) {
      notifications.show({ title: 'Error', message: 'Error al eliminar.', color: 'red' });
    }
  };

  const getPropertyAddress = (id: string): string => {
    const property = allPropertiesGlobal.find(p => p.id === id);
    return property ? property.address : 'Desconocido';
  };

  const getRenterName = (id: string): string => {
    const person = allPersonsGlobal.find(p => p.id === id);
    return person ? person.full_name : 'Desconocido';
  };

  const getStatusColor = (status?: string): string => {
    switch(status?.toLowerCase()) {
      case 'pending': return 'yellow';
      case 'in progress':
      case 'in_progress': return 'blue';
      case 'completed': return 'green';
      case 'cancelled': return 'red';
      default: return 'gray';
    }
  };
  const translateStatus = (status?: string) => {
    switch (status?.toLowerCase()) {
      case 'pending': return 'Pendiente';
      case 'in progress':
      case 'in_progress': return 'En Progreso';
      case 'completed': return 'Completado';
      case 'cancelled': return 'Cancelado';
      default: return status || 'Desconocido';
    }
  };

  const handleSaveRequest = async () => {
    if (!selectedRequest || !selectedRequest.description) {
      notifications.show({ title: 'Error', message: 'La descripción es requerida.', color: 'red' });
      return;
    }
    if (!selectedRequest.property_id && (isAdmin || isManager)){
        notifications.show({ title: 'Error', message: 'La propiedad es requerida para administradores/encargados.', color: 'red' });
        return;
    }

    if (isResident && user?.person_id) {
        if (!selectedRequest.id || selectedRequest.renter_id === user.person_id) {
            selectedRequest.renter_id = user.person_id;
        } else {
            notifications.show({ title: 'Error', message: 'No puede cambiar el solicitante de esta petición.', color: 'red' });
            return;
        }
    }

    const requestToSave = {
      ...selectedRequest,
      request_date: selectedRequest.request_date ? 
        (new Date(selectedRequest.request_date).toISOString()) : 
        new Date().toISOString()
    };

    try {
      if (requestToSave.id) {
        // API should enforce detailed field update permissions
        await maintenanceRequestApi.update(requestToSave.id, requestToSave as MaintenanceRequest);
        notifications.show({ title: 'Éxito', message: 'Solicitud actualizada.', color: 'green' });
      } else {
        const { id, ...newRequest } = requestToSave;
        await maintenanceRequestApi.create(newRequest as Omit<MaintenanceRequest, 'id'>);
        notifications.show({ title: 'Éxito', message: 'Solicitud creada.', color: 'green' });
      }
      refetchRequests();
      setIsModalOpen(false);
    } catch (error) {
      console.error('Error saving maintenance request:', error);
      notifications.show({ title: 'Error', message: 'Error al guardar.', color: 'red' });
    }
  };
  
  // Determine if the current user can create any maintenance request
  const canCreateAnyRequest = isAdmin || isManager || isResident;

  return (
    <Container size="xl">
      <LoadingOverlay visible={isLoadingRequests || !user} />
      {requestsError && <Text color="red">Error al cargar solicitudes: {requestsError.message}</Text>}
      
      <Group justify="space-between" mb="md">
        <Title order={1}>Solicitudes de Mantenimiento</Title>
        <Group>
          <Button 
            leftSection={<IconRefresh size={16} />}
            variant="outline"
            onClick={() => refetchRequests()}
          >
            Actualizar
          </Button>
          {canCreateAnyRequest && (
            <Button 
              leftSection={<IconPlus size={16} />} 
              onClick={handleAdd}
              data-testid="add-maintenance-request-button"
            >
              Nueva Solicitud
            </Button>
          )}
        </Group>
      </Group>
      
      <TextInput
        placeholder="Buscar por descripción..."
        mb="md"
        leftSection={<IconSearch size={16} />}
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.currentTarget.value)}
      />

      <Table striped highlightOnHover withTableBorder>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Propiedad</Table.Th>
            <Table.Th>Solicitante</Table.Th>
            <Table.Th>Descripción</Table.Th>
            <Table.Th>Fecha</Table.Th>
            <Table.Th>Estado</Table.Th>
            <Table.Th>Acciones</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {filteredRequests.length === 0 ? (
            <Table.Tr>
              <Table.Td colSpan={6} align="center">No se encontraron solicitudes</Table.Td>
            </Table.Tr>
          ) : (
            filteredRequests.map((request) => (
              <Table.Tr key={request.id}>
                <Table.Td>{getPropertyAddress(request.property_id)}</Table.Td>
                <Table.Td>{getRenterName(request.renter_id)}</Table.Td>
                <Table.Td>{request.description}</Table.Td>
                <Table.Td>{new Date(request.request_date).toLocaleDateString()}</Table.Td>
                <Table.Td>
                  <Badge color={getStatusColor(request.status)}>
                    {translateStatus(request.status)}
                  </Badge>
                </Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    {canEditRequest(request) && (
                      <ActionIcon variant="subtle" color="blue" onClick={() => handleEdit(request)}>
                        <IconEdit size={16} />
                      </ActionIcon>
                    )}
                    {canDeleteRequest(request) && (
                      <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(request.id!)}>
                        <IconTrash size={16} />
                      </ActionIcon>
                    )}
                  </Group>
                </Table.Td>
              </Table.Tr>
            ))
          )}
        </Table.Tbody>
      </Table>

      {isModalOpen && selectedRequest && (
        <StableModal 
          opened={isModalOpen} 
          onClose={() => setIsModalOpen(false)} 
          title={selectedRequest.id ? "Editar Solicitud" : "Nueva Solicitud"}
          centered size="lg"
        >
          <Select
            label="Propiedad"
            placeholder="Seleccione una propiedad"
            data={modalProperties.map(p => ({ value: p.id, label: p.address }))}
            value={selectedRequest.property_id}
            onChange={(value) => setSelectedRequest(prev => ({...prev, property_id: value || ''}))}
            required
            disabled={isResident && !!selectedRequest.id} // Resident cannot change property on existing request
            mb="md"
          />
          <Select
            label="Solicitante (Inquilino)"
            placeholder="Seleccione un inquilino"
            data={modalPersons.map(p => ({ value: p.id, label: p.full_name }))}
            value={selectedRequest.renter_id}
            onChange={(value) => setSelectedRequest(prev => ({...prev, renter_id: value || ''}))}
            required
            disabled={isResident} // Resident is always the requester
            mb="md"
          />
          <Textarea
            label="Descripción"
            placeholder="Describa el problema"
            required
            value={selectedRequest.description}
            onChange={(e) => {
              // Use optional chaining for robustness
              const value = e.currentTarget?.value;
              if (value !== undefined) { // Only update state if value is valid
                setSelectedRequest(prev => ({...prev, description: value}));
              }
            }}
            disabled={isManager && !!selectedRequest.id && selectedRequest.renter_id !== user?.person_id} // Manager can only edit description if not their own? Or never?
            mb="md"
            minRows={3}
          />
          <DateInput
            label="Fecha de Solicitud"
            value={selectedRequest.request_date ? new Date(selectedRequest.request_date) : new Date()}
            onChange={(date) => setSelectedRequest(prev => ({...prev, request_date: date?.toISOString() || new Date().toISOString()}))}
            required
            disabled={!isAdmin && !!selectedRequest.id} // Only admin can change date of existing request
            mb="md"
          />
          <Select
            label="Estado"
            placeholder="Seleccione un estado"
            data={[
              { value: 'pending', label: 'Pendiente' },
              { value: 'in_progress', label: 'En Progreso' },
              { value: 'completed', label: 'Completado' },
              { value: 'cancelled', label: 'Cancelado' },
            ]}
            value={selectedRequest.status}
            onChange={(value) => setSelectedRequest(prev => ({...prev, status: value || 'pending'}))}
            required
            disabled={!canChangeStatus(selectedRequest as MaintenanceRequest)} // Use selectedRequest directly
            mb="xl"
          />
          <Group justify="flex-end">
            <Button variant="outline" onClick={() => setIsModalOpen(false)}>Cancelar</Button>
            <Button onClick={handleSaveRequest}>Guardar</Button>
          </Group>
        </StableModal>
      )}
    </Container>
  );
} 