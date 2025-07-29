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
  Stack,
  LoadingOverlay
} from '@mantine/core';
import { IconHistory, IconEdit, IconTrash, IconPlus, IconFilter, IconEye } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { useEffect, useState } from 'react';
import { rentalHistoryApi, personApi, rentalApi, propertyApi } from '../api/apiService';
import { DateInput } from '@mantine/dates';
import { RentalHistory, Person, Rental, Property } from '../types';
import { notifications } from '@mantine/notifications';
import { useDisclosure } from '@mantine/hooks';
import { StableModal } from '../components/ui/StableModal';

export default function RentalHistoryPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [histories, setHistories] = useState<RentalHistory[]>([]);
  const [persons, setPersons] = useState<Person[]>([]);
  const [rentals, setRentals] = useState<Rental[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [opened, { open, close }] = useDisclosure(false);
  const [currentHistory, setCurrentHistory] = useState<Partial<RentalHistory> | null>(null);
  const [isViewMode, setIsViewMode] = useState(false);

  const [filterOpened, filterHandlers] = useDisclosure(false);
  const [startDate, setStartDate] = useState<Date | null>(null);
  const [endDate, setEndDate] = useState<Date | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');

  const fetchHistories = async () => {
    if (!user) {
      setHistories([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      let data: RentalHistory[];
      let adminAPIFilters: { status?: string; startDate?: string; endDate?: string } | undefined = undefined;

      if (isAdmin) {
        if (statusFilter) {
          adminAPIFilters = { status: statusFilter };
        } else if (startDate && endDate) {
          adminAPIFilters = { startDate: startDate.toISOString(), endDate: endDate.toISOString() };
        }
      }
      // getForCurrentUser will pass adminAPIFilters to getAll if user is admin, otherwise filters are ignored by getAll.
      // The backend /api/rental-history (RentalHistoryController.GetAll) handles role scoping and applies admin filters from query params.
      data = await rentalHistoryApi.getForCurrentUser(user, adminAPIFilters);
      setHistories(data);
    } catch (error) {
      console.error('Failed to fetch rental histories:', error);
      notifications.show({
        title: 'Error',
        message: 'Error al cargar datos de historial de alquileres',
        color: 'red'
      });
      setHistories([]); // Clear histories on error
    } finally {
      setLoading(false);
    }
  };

  const fetchRelatedData = async () => {
    setLoading(true);
    try {
      const [personsData, rentalsData, propertiesData] = await Promise.all([
        personApi.getAll(),
        rentalApi.getAll(),
        propertyApi.getAll()
      ]);
      setPersons(personsData);
      setRentals(rentalsData);
      setProperties(propertiesData);
    } catch (error) {
      console.error('Failed to fetch related data:', error);
    } finally {
    }
  };

  useEffect(() => {
    if (user) {
      fetchHistories();
      if (isAdmin || user.role === 'manager') {
        fetchRelatedData();
      }
    }
  }, [user, statusFilter, startDate, endDate]);

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
      status: '',
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
        notifications.show({
          title: 'Éxito',
          message: 'Registro de historial eliminado correctamente',
          color: 'green'
        });
        fetchHistories();
      } catch (error) {
        console.error('Failed to delete history record:', error);
        notifications.show({
          title: 'Error',
          message: 'Error al eliminar el registro de historial',
          color: 'red'
        });
      } finally {
        setLoading(false);
      }
    }
  };

  const handleSubmitModal = async () => {
    if (isViewMode) {
      close();
      return;
    }
    if (!isAdmin) return;

    if (!currentHistory || !currentHistory.person_id || !currentHistory.rental_id || !currentHistory.status) {
      notifications.show({
        title: 'Error de Validación',
        message: 'Por favor complete todos los campos requeridos (Persona, Alquiler, Estado).',
        color: 'red'
      });
      return;
    }

    setLoading(true);
    try {
      const historyToSubmit: Partial<RentalHistory> = {
        ...currentHistory,
        end_date: currentHistory.end_date ? 
          (typeof currentHistory.end_date === 'string' && currentHistory.end_date.includes('T') ? 
            currentHistory.end_date : 
            new Date(currentHistory.end_date).toISOString().split('T')[0] + 'T00:00:00Z') : 
            new Date().toISOString(),
      };

      if (historyToSubmit.id) {
        await rentalHistoryApi.update(historyToSubmit.id, historyToSubmit);
        notifications.show({ title: 'Éxito', message: 'Registro de historial actualizado.', color: 'green' });
      } else {
        await rentalHistoryApi.create(historyToSubmit as Omit<RentalHistory, 'id'>);
        notifications.show({ title: 'Éxito', message: 'Registro de historial creado.', color: 'green' });
      }
      close();
      fetchHistories();
    } catch (error) {
      console.error('Failed to save history record:', error);
      const errorMsg = (error as any).response?.data?.error || 'Error al guardar el registro.';
      notifications.show({ title: 'Error', message: errorMsg, color: 'red'});
    } finally {
      setLoading(false);
    }
  };

  const applyFilters = () => {
    if (!isAdmin) return;
    fetchHistories();
    filterHandlers.close();
  };

  const clearAdminFilters = () => {
    if (!isAdmin) return;
    setStartDate(null);
    setEndDate(null);
    setStatusFilter('');
    filterHandlers.close();
  };

  const formatDate = (dateString: string | Date | undefined) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch (e) {
      return 'Fecha Inválida';
    }
  };

  const getPersonName = (personId: string) => {
    const person = persons.find(p => p.id === personId);
    return person ? person.full_name : personId ? 'ID: '+personId.substring(0,8) : 'Desconocido';
  };
  
  const getRentalInfo = (rentalId: string) => {
    const rental = rentals.find(r => r.id === rentalId);
    if (!rental) return rentalId ? 'Alquiler ID: '+rentalId.substring(0,8) : 'Desconocido';
    const property = properties.find(p => p.id === rental.property_id);
    const renter = persons.find(p => p.id === rental.renter_id);
    let info = `Alquiler de ${renter ? renter.full_name : 'inquilino desconocido'}`;
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
      case 'pending': return 'blue';
      default: return 'gray';
    }
  };

  const translateStatus = (status: string = '') => {
    switch (status?.toLowerCase()) {
      case 'active': return 'Activo';
      case 'terminated': return 'Terminado';
      case 'expired': return 'Expirado';
      case 'pending': return 'Pendiente';
      default: return status;
    }
  };

  const personOptions = isAdmin ? persons.map(p => ({ value: p.id, label: p.full_name })) : [];
  const rentalOptions = isAdmin ? rentals.map(r => ({ value: r.id, label: getRentalInfo(r.id) })) : [];
  const statusOptions = [
    { value: 'active', label: 'Activo' },
    { value: 'terminated', label: 'Terminado' },
    { value: 'expired', label: 'Expirado' },
    { value: 'pending', label: 'Pendiente' },
  ];

  return (
    <Container fluid>
      <Group justify="space-between" mb="lg">
        <Group>
          <ThemeIcon size="xl" radius="md" variant="gradient" gradient={{ from: 'cyan', to: 'blue' }}>
            <IconHistory size={28} />
          </ThemeIcon>
          <Title order={2}>Historial de Alquileres</Title>
        </Group>
        <Group>
          {isAdmin && (
            <Button leftSection={<IconFilter size={14} />} onClick={filterHandlers.open} variant="outline">
              Filtrar
            </Button>
          )}
          {isAdmin && (
            <Button leftSection={<IconPlus size={14} />} onClick={handleCreate}>
              Nuevo Historial
            </Button>
          )}
        </Group>
      </Group>

      <Paper shadow="sm" p="md" withBorder>
        <LoadingOverlay visible={loading} />
        {histories.length === 0 && !loading && (
          <Text ta="center" p="lg">No hay registros de historial para mostrar.</Text>
        )}
        {histories.length > 0 && (
          <Table striped highlightOnHover verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Persona</Table.Th>
                <Table.Th>Alquiler (Propiedad)</Table.Th>
                <Table.Th>Estado</Table.Th>
                <Table.Th>Fecha Fin</Table.Th>
                <Table.Th>Razón Fin</Table.Th>
                <Table.Th>Acciones</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {histories.map((history) => (
                <Table.Tr key={history.id}>
                  <Table.Td>{getPersonName(history.person_id)}</Table.Td>
                  <Table.Td>{getRentalInfo(history.rental_id)}</Table.Td>
                  <Table.Td>
                    <Badge color={getStatusBadgeColor(history.status)} variant="light">
                      {translateStatus(history.status)}
                    </Badge>
                  </Table.Td>
                  <Table.Td>{formatDate(history.end_date)}</Table.Td>
                  <Table.Td>{history.end_reason || '-'}</Table.Td>
                  <Table.Td>
                    <Group gap="xs" wrap="nowrap">
                      <ActionIcon variant="subtle" color="blue" onClick={() => handleView(history)} title="Ver Detalles">
                        <IconEye size={18} />
                      </ActionIcon>
                      {isAdmin && (
                        <ActionIcon variant="subtle" color="yellow" onClick={() => handleEdit(history)} title="Editar">
                          <IconEdit size={18} />
                        </ActionIcon>
                      )}
                      {isAdmin && (
                        <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(history.id)} title="Eliminar">
                          <IconTrash size={18} />
                        </ActionIcon>
                      )}
                    </Group>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Paper>

      {isAdmin && (
        <StableModal opened={filterOpened} onClose={filterHandlers.close} title="Filtrar Historial">
          <Stack>
            <Select
              label="Estado"
              placeholder="Seleccione un estado"
              data={statusOptions}
              value={statusFilter}
              onChange={(value) => setStatusFilter(value || '')}
              clearable
            />
            <DateInput
              label="Fecha de Inicio (Rango)"
              value={startDate}
              onChange={setStartDate}
              placeholder="YYYY-MM-DD"
              clearable
            />
            <DateInput
              label="Fecha de Fin (Rango)"
              value={endDate}
              onChange={setEndDate}
              placeholder="YYYY-MM-DD"
              clearable
            />
            <Group justify="flex-end" mt="md">
              <Button variant="default" onClick={clearAdminFilters}>Limpiar</Button>
              <Button onClick={applyFilters}>Aplicar Filtros</Button>
            </Group>
          </Stack>
        </StableModal>
      )}

      <StableModal 
        opened={opened} 
        onClose={close} 
        title={isViewMode ? 'Detalles del Historial' : (currentHistory?.id ? 'Editar Historial' : 'Nuevo Historial')}
        size="lg"
      >
        <Stack gap="md">
          <Select
            label="Persona"
            placeholder="Seleccione una persona"
            data={personOptions}
            value={currentHistory?.person_id || ''}
            onChange={(value) => setCurrentHistory(prev => ({ ...prev, person_id: value || '' }))}
            disabled={isViewMode || !isAdmin}
            required
            searchable
          />
          <Select
            label="Alquiler (Inquilino en Propiedad)"
            placeholder="Seleccione un alquiler"
            data={rentalOptions}
            value={currentHistory?.rental_id || ''}
            onChange={(value) => setCurrentHistory(prev => ({ ...prev, rental_id: value || '' }))}
            disabled={isViewMode || !isAdmin}
            required
            searchable
          />
          <Select
            label="Estado"
            placeholder="Seleccione un estado"
            data={statusOptions}
            value={currentHistory?.status || ''}
            onChange={(value) => setCurrentHistory(prev => ({ ...prev, status: value || '' }))}
            disabled={isViewMode || !isAdmin}
            required
          />
          <DateInput
            label="Fecha de Fin"
            value={currentHistory?.end_date ? new Date(currentHistory.end_date) : null}
            onChange={(date) => setCurrentHistory(prev => ({ ...prev, end_date: date ? date.toISOString().split('T')[0] : '' }))}
            placeholder="YYYY-MM-DD"
            disabled={isViewMode || !isAdmin}
            required
          />
          <TextInput
            label="Razón de Finalización"
            placeholder="Ej: Contrato finalizado, Mudanza, etc."
            value={currentHistory?.end_reason || ''}
            onChange={(event) => setCurrentHistory(prev => ({ ...prev, end_reason: event.currentTarget.value }))}
            disabled={isViewMode || !isAdmin}
          />
          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={close}>{(isViewMode || !isAdmin) ? 'Cerrar' : 'Cancelar'}</Button>
            {!isViewMode && isAdmin && (
              <Button onClick={handleSubmitModal}>Guardar Historial</Button>
            )}
          </Group>
        </Stack>
      </StableModal>
    </Container>
  );
} 