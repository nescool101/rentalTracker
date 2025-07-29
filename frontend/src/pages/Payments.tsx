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
  NumberInput,
  Select,
  Checkbox,
  LoadingOverlay
} from '@mantine/core';
import { IconCreditCard, IconEdit, IconTrash, IconPlus, IconFilter, IconReceipt } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { useState, useEffect } from 'react';
import { rentPaymentApi, rentalApi, propertyApi, personApi } from '../api/apiService';
import { notifications } from '@mantine/notifications';
import { useDisclosure } from '@mantine/hooks';
import { RentPayment, Rental, Property, User, Person } from '../types';
import { DateInput } from '@mantine/dates';
import { StableModal } from '../components/ui/StableModal';
import { useQuery, useQueryClient } from '@tanstack/react-query';

export default function Payments() {
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const isAdmin = user?.role === 'admin';
  const isManager = user?.role === 'manager';
  const isStandardUser = !isAdmin && !isManager;

  const { 
    data: payments = [], 
    isLoading: isLoadingPayments,
    error: paymentsError,
    refetch: refetchPayments
  } = useQuery<RentPayment[]>({
    queryKey: ['payments', user?.id, user?.role, user?.person_id],
    queryFn: () => rentPaymentApi.getForCurrentUser(user as User | null),
    enabled: !!user,
  });

  const { data: allRentalsGlobal = [], isLoading: isLoadingAllRentals } = useQuery<Rental[]>({
    queryKey: ['allRentalsGlobalForPayments'],
    queryFn: rentalApi.getAll,
    enabled: isAdmin || isManager,
  });

  const { data: allPropertiesGlobal = [], isLoading: isLoadingAllProperties } = useQuery<Property[]>({
    queryKey: ['allPropertiesGlobalForPayments'],
    queryFn: propertyApi.getAll,
    enabled: isAdmin || isManager,
  });
  
  const { data: allPersons = [] } = useQuery<Person[]>({
      queryKey: ['allPersonsForPayments'],
      queryFn: personApi.getAll,
      enabled: isAdmin || isManager,
  });

  const [modalRentals, setModalRentals] = useState<Rental[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentPayment, setCurrentPayment] = useState<Partial<RentPayment> | null>(null);
  const [filterOpened, filterHandlers] = useDisclosure(false);
  const [filterStartDate, setFilterStartDate] = useState<Date | null>(null);
  const [filterEndDate, setFilterEndDate] = useState<Date | null>(null);
  const [filterShowLateOnly, setFilterShowLateOnly] = useState(false);

  useEffect(() => {
    if (!user || !user.person_id) return;
    if (isAdmin) {
      setModalRentals(allRentalsGlobal);
    } else if (isManager && user.person_id) {
      const managerPropertyIds = allPropertiesGlobal
        .filter(p => p.manager_ids?.includes(user.person_id!))
        .map(p => p.id);
      const managerRentals = allRentalsGlobal.filter(r => managerPropertyIds.includes(r.property_id));
      setModalRentals(managerRentals);
    }
  }, [user, isAdmin, isManager, allRentalsGlobal, allPropertiesGlobal]);

  const canManagePayment = (payment?: Partial<RentPayment> | null): boolean => {
    if (!user) return false;
    if (isAdmin) return true;
    if (isManager && user.person_id) {
        if (!payment || !payment.rental_id) return true;
        const rental = allRentalsGlobal.find(r => r.id === payment.rental_id);
        if (!rental) return false;
        const property = allPropertiesGlobal.find(p => p.id === rental.property_id);
        return property?.manager_ids?.includes(user.person_id!) ?? false;
    }
    return false;
  };

  const handleEdit = async (payment: RentPayment) => {
    if (!canManagePayment(payment)) return;
    setCurrentPayment(payment);
    setIsModalOpen(true);
  };

  const handleCreate = async () => {
    if (!isAdmin && !isManager) return;
    if (isAdmin && modalRentals.length === 0 && allRentalsGlobal.length > 0) setModalRentals(allRentalsGlobal);
    
    setCurrentPayment({
      rental_id: modalRentals.length > 0 ? modalRentals[0].id : '',
      payment_date: new Date().toISOString(),
      amount_paid: 0,
      paid_on_time: true
    });
    setIsModalOpen(true);
  };

  const handleDelete = async (payment: RentPayment) => {
    if (!canManagePayment(payment)) {
      notifications.show({ title: 'Error', message: 'No autorizado para eliminar.', color: 'red' });
      return;
    }
    if (confirm('¿Está seguro que desea eliminar este pago?')) {
      try {
        await rentPaymentApi.delete(payment.id!);
        refetchPayments();
        queryClient.invalidateQueries({ queryKey: ['payments', user?.id, user?.role, user?.person_id] });
        notifications.show({ title: 'Éxito', message: 'Pago eliminado.', color: 'green' });
      } catch (error) {
        notifications.show({ title: 'Error', message: 'Error al eliminar el pago.', color: 'red' });
      }
    }
  };

  const handleSubmit = async () => {
    if (!currentPayment || !currentPayment.rental_id) {
      notifications.show({ title: 'Error', message: 'Alquiler es requerido.', color: 'red' });
      return;
    }
    if(!canManagePayment(currentPayment)){
        notifications.show({ title: 'Error', message: 'No autorizado para guardar este pago.', color: 'red' });
        return;
    }

    const paymentToSave = {
      ...currentPayment,
      payment_date: currentPayment.payment_date ? 
        (new Date(currentPayment.payment_date).toISOString()) : 
        new Date().toISOString()
    };

    try {
      if (paymentToSave.id) {
        await rentPaymentApi.update(paymentToSave.id, paymentToSave as RentPayment);
        notifications.show({ title: 'Éxito', message: 'Pago actualizado.', color: 'green' });
      } else {
        const { id, ...newPayment } = paymentToSave;
        await rentPaymentApi.create(newPayment as Omit<RentPayment, 'id'>);
        notifications.show({ title: 'Éxito', message: 'Pago creado.', color: 'green' });
      }
      queryClient.invalidateQueries({ queryKey: ['payments', user?.id, user?.role, user?.person_id] });
      setIsModalOpen(false);
    } catch (error) {
      notifications.show({ title: 'Error', message: 'Error al guardar pago.', color: 'red' });
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString();
  };

  const getPaymentDetails = (payment: RentPayment): { propertyAddress: string, renterName: string } => {
      const rental = allRentalsGlobal.find(r => r.id === payment.rental_id);
      if (!rental) return { propertyAddress: 'Alquiler no encontrado', renterName: 'N/A' };
      const property = allPropertiesGlobal.find(p => p.id === rental.property_id);
      const renter = allPersons.find(p => p.id === rental.renter_id);
      return {
          propertyAddress: property ? `${property.address}${property.apt_number ? ", "+property.apt_number : ""}, ${property.city}` : 'Propiedad desconocida',
          renterName: renter ? renter.full_name : 'Inquilino Desconocido'
      };
  };
  
  const getRentalDisplayForModal = (rental: Rental) => {
    const property = allPropertiesGlobal.find(p => p.id === rental.property_id);
    const renter = allPersons.find(p => p.id === rental.renter_id);
    const propDisplay = property ? `${property.address.substring(0,15)}..` : 'Propiedad desc.';
    const renterDisplay = renter ? renter.full_name : 'Inquilino desc.';
    return `${propDisplay} - ${renterDisplay} (Fin: ${formatDate(rental.end_date)})`;
  };

  const isLoading = isLoadingPayments || ( (isAdmin || isManager) && (isLoadingAllRentals || isLoadingAllProperties));

  return (
    <Container size="xl">
      <LoadingOverlay visible={isLoading} overlayProps={{ blur: 2 }} />
      {paymentsError && <Text color="red" ta="center">Error al cargar pagos: {(paymentsError as Error).message}</Text>}
      
      <Group justify="space-between" mb="xl">
        <Title order={1}>{(isStandardUser && user?.person_id) ? 'Mis Pagos' : 'Gestión de Pagos'}</Title>
        <Group>
          {isAdmin && (
            <Button 
              leftSection={<IconFilter size={16} />} 
              variant="outline"
              onClick={filterHandlers.open}
            >
              Filtros
            </Button>
          )}
          {isAdmin && (
            <Button 
              leftSection={<IconPlus size={16} />} 
              onClick={handleCreate}
              data-testid="add-payment-button"
            >
              Añadir Pago
            </Button>
          )}
        </Group>
      </Group>
      
      <Paper shadow="sm" p="lg" radius="md" withBorder mb="xl">
        <Group mb="lg">
          <ThemeIcon size="xl" color={isStandardUser ? "green" : "blue"} radius="md">
            {isStandardUser ? <IconReceipt size={24} /> : <IconCreditCard size={24} />}
          </ThemeIcon>
          <Title order={2}>{(isStandardUser && user?.person_id) ? 'Historial de Pagos' : 'Pagos de Alquiler'}</Title>
        </Group>
        
        {payments.length === 0 && !isLoadingPayments ? (
          <Text c="dimmed" ta="center" py="xl">
            No se encontraron pagos.
            {(isAdmin || isManager) && ' Use el botón Añadir Pago para crear uno.'}
            {isStandardUser && ' No tiene pagos registrados actualmente.'}
          </Text>
        ) : (
          <Table striped highlightOnHover withTableBorder>
            <Table.Thead>
              <Table.Tr>
                {isAdmin || isManager ? <Table.Th>ID Pago</Table.Th> : null}
                <Table.Th>Propiedad</Table.Th>
                {(isAdmin || isManager) && <Table.Th>Inquilino</Table.Th>}
                <Table.Th>Fecha de Pago</Table.Th>
                <Table.Th>Cantidad</Table.Th>
                <Table.Th>Estado</Table.Th>
                {(isAdmin || isManager) && <Table.Th>Acciones</Table.Th>}
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {payments.map((payment) => {
                const details = getPaymentDetails(payment);
                return (
                  <Table.Tr key={payment.id}>
                    {isAdmin || isManager ? <Table.Td>{payment.id.substring(0,8)}...</Table.Td> : null}
                    <Table.Td>{details.propertyAddress}</Table.Td>
                    {(isAdmin || isManager) && <Table.Td>{details.renterName}</Table.Td>}
                    <Table.Td>{formatDate(payment.payment_date)}</Table.Td>
                    <Table.Td>${payment.amount_paid.toFixed(2)}</Table.Td>
                    <Table.Td>
                      <Badge color={payment.paid_on_time ? 'green' : 'red'}>
                        {payment.paid_on_time ? 'A Tiempo' : 'Atrasado'}
                      </Badge>
                    </Table.Td>
                    {(isAdmin || isManager) && canManagePayment(payment) && (
                      <Table.Td>
                        <Group gap="xs">
                          <ActionIcon variant="subtle" color="blue" onClick={() => handleEdit(payment)} title="Editar">
                            <IconEdit size={16} />
                          </ActionIcon>
                          <ActionIcon variant="subtle" color="red" onClick={() => handleDelete(payment)} title="Eliminar">
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Group>
                      </Table.Td>
                    )}
                  </Table.Tr>
                );
            })}
            </Table.Tbody>
          </Table>
        )}
      </Paper>

      {isModalOpen && (isAdmin || isManager) && currentPayment && (
        <StableModal
          opened={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          title={currentPayment.id ? "Editar Pago" : "Añadir Pago"}
          centered
        >
          <Select
            label="Alquiler Asociado"
            placeholder="Seleccione un alquiler"
            required
            data={modalRentals.map(rental => ({
              value: rental.id,
              label: getRentalDisplayForModal(rental)
            }))}
            value={currentPayment.rental_id || ''}
            onChange={(value) => setCurrentPayment({ ...currentPayment, rental_id: value || '' })}
            mb="md"
            searchable
            nothingFoundMessage="No hay alquileres disponibles"
          />
          <DateInput
            label="Fecha de Pago"
            value={currentPayment.payment_date ? new Date(currentPayment.payment_date) : new Date()}
            onChange={(date) => setCurrentPayment({ ...currentPayment, payment_date: date?.toISOString() })}
            required
            mb="md"
          />
          <NumberInput
            label="Cantidad Pagada"
            placeholder="Ingrese cantidad"
            required
            min={0}
            decimalScale={2}
            fixedDecimalScale
            value={currentPayment.amount_paid || 0}
            onChange={(value) => setCurrentPayment({ ...currentPayment, amount_paid: Number(value) || 0 })}
            mb="md"
          />
          <Checkbox
            label="Pagado a Tiempo"
            checked={currentPayment.paid_on_time || false}
            onChange={(event) => setCurrentPayment({ ...currentPayment, paid_on_time: event.currentTarget.checked })}
            mb="xl"
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={() => setIsModalOpen(false)}>Cancelar</Button>
            <Button onClick={handleSubmit}>Guardar Pago</Button>
          </Group>
        </StableModal>
      )}
      {isAdmin && (
        <StableModal opened={filterOpened} onClose={filterHandlers.close} title="Filtrar Pagos">
            <DateInput label="Desde Fecha" value={filterStartDate} onChange={setFilterStartDate} mb="sm" clearable />
            <DateInput label="Hasta Fecha" value={filterEndDate} onChange={setFilterEndDate} mb="sm" clearable />
            <Checkbox label="Mostrar solo pagos atrasados" checked={filterShowLateOnly} onChange={(e) => setFilterShowLateOnly(e.currentTarget.checked)} mb="xl"/>
            <Group justify="flex-end">
                <Button variant="default" onClick={filterHandlers.close}>Cancelar</Button>
                 <Button disabled>Aplicar (No Implementado)</Button>
            </Group>
        </StableModal>
      )}
    </Container>
  );
} 