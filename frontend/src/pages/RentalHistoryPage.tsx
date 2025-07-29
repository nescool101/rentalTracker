import { 
  Title, 
  Container, 
  Text, 
  Group, 
  Button,
  Badge,
  Table,
  ActionIcon,
  TextInput,
  Select,
  LoadingOverlay,
  Textarea
} from '@mantine/core';
import { IconEdit, IconTrash, IconPlus, IconFilter, IconEye } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { useEffect, useState } from 'react';
import { rentalHistoryApi, personApi, rentalApi, propertyApi } from '../api/apiService';
import { DateInput } from '@mantine/dates';
import { RentalHistory, Person, Rental, Property, User } from '../types';
import { notifications } from '@mantine/notifications';
import { useDisclosure } from '@mantine/hooks';
import { StableModal } from '../components/ui/StableModal';
import { useQuery, useQueryClient } from '@tanstack/react-query';

export default function RentalHistoryPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const isAdmin = user?.role === 'admin';
  const isStandardUser = user?.role === 'user' || user?.role === 'resident';

  const [histories, setHistories] = useState<RentalHistory[]>([]);
  const [persons, setPersons] = useState<Person[]>([]);
  const [rentals, setRentals] = useState<Rental[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [opened, { open, close }] = useDisclosure(false);
  const [currentHistory, setCurrentHistory] = useState<Partial<RentalHistory> | null>(null);
  const [isViewMode, setIsViewMode] = useState(false);

  const [filterOpened, filterHandlers] = useDisclosure(false);
  const [adminStartDateFilter, setAdminStartDateFilter] = useState<Date | null>(null);
  const [adminEndDateFilter, setAdminEndDateFilter] = useState<Date | null>(null);
  const [adminStatusFilter, setAdminStatusFilter] = useState<string>('');

  const { 
    data: fetchedHistories = [], 
    isLoading: isLoadingHistories, 
    error: fetchError,
    refetch: refetchHistories
  } = useQuery<RentalHistory[]>({
    queryKey: ['rentalHistories', user?.id, user?.role, adminStatusFilter, adminStartDateFilter?.toISOString(), adminEndDateFilter?.toISOString()],
    queryFn: async () => {
      if (!user) return [];
      const adminAPIFilters: { status?: string; startDate?: string; endDate?: string } = {};
      if (isAdmin) {
        if (adminStatusFilter) adminAPIFilters.status = adminStatusFilter;
        if (adminStartDateFilter) adminAPIFilters.startDate = adminStartDateFilter.toISOString();
        if (adminEndDateFilter) adminAPIFilters.endDate = adminEndDateFilter.toISOString();
      }
      return await rentalHistoryApi.getForCurrentUser(user as User, adminAPIFilters);
    },
    enabled: !!user,
  });

  useEffect(() => {
    if (fetchedHistories) {
      setHistories(fetchedHistories);
    }
  }, [fetchedHistories]);

  useEffect(() => {
    setLoading(isLoadingHistories);
  }, [isLoadingHistories]);
  
  useEffect(() => {
    if (fetchError) {
        console.error('Failed to fetch rental histories via useQuery:', fetchError);
        notifications.show({
            title: 'Error de Carga',
            message: (fetchError as Error).message || 'Error al cargar datos de historial de alquileres',
            color: 'red'
        });
        setHistories([]);
    }
  }, [fetchError]);

  const { data: personsData } = useQuery<Person[]>({ 
      queryKey: ['personsAllForRentalHistory'], 
      queryFn: personApi.getAll, 
      enabled: isAdmin || user?.role === 'manager' 
  });
  useEffect(() => { if (personsData) setPersons(personsData); }, [personsData]);

  const { data: rentalsData } = useQuery<Rental[]>({ 
      queryKey: ['rentalsAllForRentalHistory'], 
      queryFn: rentalApi.getAll, 
      enabled: isAdmin || user?.role === 'manager' 
  });
  useEffect(() => { if (rentalsData) setRentals(rentalsData); }, [rentalsData]);

  const { data: propertiesData } = useQuery<Property[]>({ 
      queryKey: ['propertiesAllForRentalHistory'], 
      queryFn: propertyApi.getAll, 
      enabled: isAdmin || user?.role === 'manager' 
  });
  useEffect(() => { if (propertiesData) setProperties(propertiesData); }, [propertiesData]);

  const handleView = (history: RentalHistory) => {
    setCurrentHistory(history);
    setIsViewMode(true);
    open();
  };

  const handleEdit = (history: RentalHistory) => {
    if (!isAdmin) return;
    setCurrentHistory(history);
    setIsViewMode(false);
    open();
  };

  const handleCreate = () => {
    if (!isAdmin) return;
    setCurrentHistory({
      person_id: '',
      rental_id: '',
      status: 'active', 
      end_reason: '',
      end_date: new Date().toISOString()
    });
    setIsViewMode(false);
    open();
  };

  const handleDelete = async (id: string) => {
    if (!isAdmin) return;
    if (confirm('¿Está seguro que desea eliminar este registro de historial?')) {
      setLoading(true);
      try {
        await rentalHistoryApi.delete(id);
        notifications.show({ title: 'Éxito', message: 'Registro de historial eliminado.', color: 'green' });
        queryClient.invalidateQueries({ queryKey: ['rentalHistories'] });
      } catch (error) {
        notifications.show({ title: 'Error', message: 'Error al eliminar el registro.', color: 'red' });
      } finally {
        setLoading(false);
      }
    }
  };

  const handleSubmitModal = async () => {
    if (isViewMode || !isAdmin || !currentHistory) {
      close();
      return;
    }
    if (!currentHistory.person_id || !currentHistory.rental_id || !currentHistory.status) {
      notifications.show({ title: 'Error de Validación', message: 'Persona, Alquiler y Estado son req.', color: 'red'});
      return;
    }
    setLoading(true);
    try {
      const historyToSubmit: Partial<RentalHistory> = {
        ...currentHistory,
        end_date: currentHistory.end_date ? new Date(currentHistory.end_date).toISOString() : undefined,
      };

      if (historyToSubmit.id) {
        await rentalHistoryApi.update(historyToSubmit.id, historyToSubmit as RentalHistory);
      } else {
        await rentalHistoryApi.create(historyToSubmit as Omit<RentalHistory, 'id'>);
      }
      notifications.show({ title: 'Éxito', message: 'Registro de historial guardado.', color: 'green' });
      close();
      queryClient.invalidateQueries({ queryKey: ['rentalHistories'] });
    } catch (error) {
      const errorMsg = (error as any).response?.data?.error || 'Error al guardar.';
      notifications.show({ title: 'Error', message: errorMsg, color: 'red'});
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string | Date | undefined) => {
    if (!dateString) return 'N/A';
    try { return new Date(dateString).toLocaleDateString(); } catch (e) { return 'Fecha Inválida'; }
  };

  const getPersonName = (personId?: string) => {
    if (!personId && !isStandardUser) return 'N/A';
    if (isStandardUser && user?.person_id === personId && user?.email) return user.email;
    const person = persons.find(p => p.id === personId);
    return person ? person.full_name : (personId ? 'ID: '+personId.substring(0,8)+'...' : 'Desconocido');
  };
  
  const getRentalInfo = (rentalId?: string) => {
    if (!rentalId) return 'N/A';
    const rental = rentals.find(r => r.id === rentalId);
    if (!rental) return rentalId ? 'Alquiler ID: '+rentalId.substring(0,8)+'...' : 'Desconocido';
    const property = properties.find(p => p.id === rental.property_id);
    const renterName = getPersonName(rental.renter_id);
    let info = `Alquiler de ${renterName}`;
    if (property) {
      info += ` en ${property.address}`;
    }
    return info;
  };

  const getStatusBadgeColor = (status: string = '') => {
    switch (status?.toLowerCase()) {
      case 'active': return 'green';
      case 'terminated': return 'red';
      case 'expired': return 'orange';
      default: return 'gray';
    }
  };

  const pageTitle = isStandardUser ? "Mi Historial de Alquiler" : "Gestión de Historial de Alquileres";

  return (
    <Container size="xl">
      <LoadingOverlay visible={loading || isLoadingHistories} />
      {fetchError && <Text color="red" ta="center">Error al cargar historial: {(fetchError as Error).message}</Text>}

      <Group justify="space-between" mb="xl">
        <Title order={1}>{pageTitle}</Title>
        {isAdmin && (
          <Group>
            <Button leftSection={<IconFilter size={16}/>} variant="outline" onClick={filterHandlers.open}>Filtros Admin</Button>
            <Button leftSection={<IconPlus size={16} />} onClick={handleCreate}>Crear Registro</Button>
          </Group>
        )}
      </Group>
      
      {histories.length === 0 && !isLoadingHistories ? (
        <Text c="dimmed" ta="center" py="xl">
          No se encontraron registros en el historial de alquileres.
          {isAdmin && ' Use el botón Crear Registro para añadir uno.'}
          {isStandardUser && ' No tiene historial de alquiler disponible.'}
        </Text>
      ) : (
        <Table striped highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              {!isStandardUser && <Table.Th>Persona</Table.Th>}
              <Table.Th>Alquiler (Propiedad)</Table.Th>
              <Table.Th>Estado</Table.Th>
              <Table.Th>Razón de Finalización</Table.Th>
              <Table.Th>Fecha de Finalización</Table.Th>
              <Table.Th>Acciones</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {histories.map((history) => (
              <Table.Tr key={history.id}>
                {!isStandardUser && <Table.Td>{getPersonName(history.person_id)}</Table.Td>}
                <Table.Td>{getRentalInfo(history.rental_id)}</Table.Td>
                <Table.Td><Badge color={getStatusBadgeColor(history.status)}>{history.status}</Badge></Table.Td>
                <Table.Td>{history.end_reason || 'N/A'}</Table.Td>
                <Table.Td>{formatDate(history.end_date)}</Table.Td>
                <Table.Td>
                  <ActionIcon variant="subtle" color="blue" onClick={() => handleView(history)} title="Ver Detalles">
                    <IconEye size={16} />
                  </ActionIcon>
                  {isAdmin && (
                    <>
                      <ActionIcon variant="subtle" color="orange" onClick={() => handleEdit(history)} title="Editar" ml="xs">
                        <IconEdit size={16} />
                      </ActionIcon>
                      <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(history.id)} title="Eliminar" ml="xs">
                        <IconTrash size={16} />
                      </ActionIcon>
                    </>
                  )}
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      )}

      <StableModal 
        opened={opened && (isAdmin || isViewMode)}
        onClose={close} 
        title={isViewMode ? "Ver Historial" : (currentHistory?.id ? "Editar Historial" : "Crear Historial")}
        centered size="lg"
      >
        <TextInput label="ID Persona" disabled={isViewMode || !isAdmin} required={isAdmin && !isViewMode} value={currentHistory?.person_id || ''} onChange={(e) => setCurrentHistory({...currentHistory, person_id: e.target.value})} mb="sm" error={!currentHistory?.person_id && !isViewMode && isAdmin ? "Requerido" : null}/>
        <TextInput label="ID Alquiler" disabled={isViewMode || !isAdmin} required={isAdmin && !isViewMode} value={currentHistory?.rental_id || ''} onChange={(e) => setCurrentHistory({...currentHistory, rental_id: e.target.value})} mb="sm" error={!currentHistory?.rental_id && !isViewMode && isAdmin ? "Requerido" : null}/>
        <Select label="Estado" disabled={isViewMode || !isAdmin} required={isAdmin && !isViewMode} data={['active', 'terminated', 'expired']} value={currentHistory?.status || ''} onChange={(val) => setCurrentHistory({...currentHistory, status: val || ''})} mb="sm" error={!currentHistory?.status && !isViewMode && isAdmin ? "Requerido" : null}/>
        <Textarea label="Razón de Finalización" disabled={isViewMode || !isAdmin} value={currentHistory?.end_reason || ''} onChange={(e) => setCurrentHistory({...currentHistory, end_reason: e.target.value})} mb="sm" />
        <DateInput label="Fecha de Finalización" disabled={isViewMode || !isAdmin} required={isAdmin && !isViewMode} value={currentHistory?.end_date ? new Date(currentHistory.end_date) : null} onChange={(date) => setCurrentHistory({...currentHistory, end_date: date?.toISOString()})} mb="xl" error={!currentHistory?.end_date && !isViewMode && isAdmin ? "Requerido" : null}/>
        
        <Group justify="flex-end">
          <Button variant="default" onClick={close}>{ (isViewMode && !isAdmin) ? "Cerrar" : (isAdmin && isViewMode ? "Cerrar" : "Cancelar")}</Button>
          {!isViewMode && isAdmin && <Button onClick={handleSubmitModal}>Guardar Historial</Button>}
        </Group>
      </StableModal>

      {isAdmin && (
          <StableModal opened={filterOpened} onClose={filterHandlers.close} title="Filtrar Historial de Alquileres (Admin)">
              <Select label="Estado" placeholder="Filtrar por estado" clearable data={['active', 'terminated', 'expired']} value={adminStatusFilter} onChange={(val) => setAdminStatusFilter(val || '')} mb="sm"/>
              <DateInput label="Desde Fecha de Finalización" value={adminStartDateFilter} onChange={setAdminStartDateFilter} mb="sm" clearable/>
              <DateInput label="Hasta Fecha de Finalización" value={adminEndDateFilter} onChange={setAdminEndDateFilter} mb="xl" clearable/>
              <Group justify="flex-end">
                  <Button variant="default" onClick={() => { setAdminStatusFilter(''); setAdminStartDateFilter(null); setAdminEndDateFilter(null); filterHandlers.close(); refetchHistories(); }}>Limpiar y Cerrar</Button>
                   <Button onClick={() => { refetchHistories(); filterHandlers.close();}}>Aplicar y Cerrar</Button>
              </Group>
          </StableModal>
      )}
    </Container>
  );
} 