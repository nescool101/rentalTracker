import { useState, useEffect } from 'react';
import {
  Container,
  Paper,
  Title,
  TextInput,
  Button,
  Stack,
  Alert,
  Text,
  Group,
  Modal,
  Table,
  ActionIcon,
  Badge,
  Loader,
  Center,
  Select,
  Pagination,

  Flex,
  Box,
  Tooltip,
  LoadingOverlay
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconAlertCircle, IconEdit, IconTrash, IconPlus, IconSearch, IconEye, IconUser } from '@tabler/icons-react';
import { bankAccountApi, personApi } from '../../api/apiService';
import { useAuth } from '../../contexts/AuthContext';
import type { BankAccount, Person } from '../../types';

interface BankAccountWithPerson extends BankAccount {
  person_name?: string;
}

export default function BankAccountManagement() {
  const { user } = useAuth();
  const [bankAccounts, setBankAccounts] = useState<BankAccountWithPerson[]>([]);
  const [persons, setPersons] = useState<Person[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [modalOpened, setModalOpened] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'view'>('create');
  const [selectedAccount, setSelectedAccount] = useState<BankAccount | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterPersonId, setFilterPersonId] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [operationLoading, setOperationLoading] = useState(false);

  const itemsPerPage = 10;

  // Check if user has admin privileges
  if (!user || user.role !== 'admin') {
    return (
      <Container size="md" my={40}>
        <Alert 
          icon={<IconAlertCircle size={16} />} 
          title="Acceso denegado" 
          color="red"
        >
          Solo los administradores pueden acceder a esta página.
        </Alert>
      </Container>
    );
  }

  const form = useForm({
    initialValues: {
      person_id: '',
      bank_name: '',
      account_type: 'Savings',
      account_number: '',
      account_holder: '',
    },
    validate: {
      person_id: (value) => (value ? null : 'Debe seleccionar una persona'),
      bank_name: (value) => (value.length < 2 ? 'El nombre del banco es obligatorio' : null),
      account_number: (value) => (value.length < 5 ? 'El número de cuenta debe tener al menos 5 caracteres' : null),
      account_holder: (value) => (value.length < 2 ? 'El nombre del titular es obligatorio' : null),
    },
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    
    try {
      const [accountsData, personsData] = await Promise.all([
        bankAccountApi.getAll(),
        personApi.getAll()
      ]);
      
      // Enrich bank accounts with person names
      const enrichedAccounts = accountsData.map(account => ({
        ...account,
        person_name: personsData.find(p => p.id === account.person_id)?.full_name || 'Desconocido'
      }));
      
      setBankAccounts(enrichedAccounts);
      setPersons(personsData);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError('Error al cargar los datos. Intente de nuevo.');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenModal = (mode: 'create' | 'edit' | 'view', account?: BankAccount) => {
    setModalMode(mode);
    setSelectedAccount(account || null);
    
    if (mode === 'create') {
      form.reset();
    } else if (account) {
      form.setValues({
        person_id: account.person_id,
        bank_name: account.bank_name,
        account_type: account.account_type,
        account_number: account.account_number,
        account_holder: account.account_holder,
      });
    }
    
    setModalOpened(true);
  };

  const handleCloseModal = () => {
    setModalOpened(false);
    setSelectedAccount(null);
    form.reset();
  };

  const handleSubmit = async (values: typeof form.values) => {
    setOperationLoading(true);
    
    try {
      if (modalMode === 'create') {
        await bankAccountApi.create(values);
        notifications.show({
          title: 'Cuenta bancaria creada',
          message: 'La cuenta bancaria se ha creado exitosamente.',
          color: 'green',
          icon: <IconCheck size={16} />,
        });
      } else if (modalMode === 'edit' && selectedAccount) {
        await bankAccountApi.update(selectedAccount.id, values);
        notifications.show({
          title: 'Cuenta bancaria actualizada',
          message: 'La cuenta bancaria se ha actualizado exitosamente.',
          color: 'green',
          icon: <IconCheck size={16} />,
        });
      }
      
      handleCloseModal();
      await fetchData();
    } catch (err) {
      console.error('Error saving bank account:', err);
      notifications.show({
        title: 'Error',
        message: 'Hubo un error al guardar la cuenta bancaria.',
        color: 'red',
        icon: <IconAlertCircle size={16} />,
      });
    } finally {
      setOperationLoading(false);
    }
  };

  const handleDelete = async (account: BankAccount) => {
    if (!confirm(`¿Está seguro de que desea eliminar la cuenta bancaria de ${account.bank_name}?`)) {
      return;
    }

    try {
      await bankAccountApi.delete(account.id);
      notifications.show({
        title: 'Cuenta bancaria eliminada',
        message: 'La cuenta bancaria se ha eliminado exitosamente.',
        color: 'green',
        icon: <IconCheck size={16} />,
      });
      await fetchData();
    } catch (err) {
      console.error('Error deleting bank account:', err);
      notifications.show({
        title: 'Error',
        message: 'Hubo un error al eliminar la cuenta bancaria.',
        color: 'red',
        icon: <IconAlertCircle size={16} />,
      });
    }
  };

  // Filter bank accounts based on search and filter criteria
  const filteredAccounts = bankAccounts.filter(account => {
    const matchesSearch = !searchTerm || 
      account.bank_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      account.account_number.includes(searchTerm) ||
      account.account_holder.toLowerCase().includes(searchTerm.toLowerCase()) ||
      account.person_name?.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesFilter = !filterPersonId || account.person_id === filterPersonId;
    
    return matchesSearch && matchesFilter;
  });

  // Paginate filtered results
  const totalPages = Math.ceil(filteredAccounts.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedAccounts = filteredAccounts.slice(startIndex, startIndex + itemsPerPage);

  const getAccountTypeBadge = (type: string) => {
    const colors: Record<string, string> = {
      'Savings': 'blue',
      'Checking': 'green',
      'Corriente': 'green',
      'Ahorros': 'blue',
    };
    return <Badge color={colors[type] || 'gray'}>{type}</Badge>;
  };

  if (loading) {
    return (
      <Container size="xl" my={40}>
        <Center>
          <Loader size="lg" />
        </Center>
      </Container>
    );
  }

  return (
    <Container size="xl" my={40}>
      <Paper radius="md" p="xl" withBorder>
        <Group justify="space-between" mb="lg">
          <Title order={2}>Gestión de Cuentas Bancarias</Title>
          <Button
            leftSection={<IconPlus size={16} />}
            onClick={() => handleOpenModal('create')}
          >
            Agregar Cuenta
          </Button>
        </Group>

        {error && (
          <Alert 
            icon={<IconAlertCircle size={16} />} 
            title="Error" 
            color="red" 
            mb="md"
          >
            {error}
          </Alert>
        )}

        {/* Search and Filter Controls */}
        <Flex gap="md" mb="lg" wrap="wrap">
          <TextInput
            placeholder="Buscar por banco, número de cuenta, titular o persona..."
            leftSection={<IconSearch size={16} />}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ flex: 1, minWidth: '300px' }}
          />
          <Select
            placeholder="Filtrar por persona"
            leftSection={<IconUser size={16} />}
            data={[
              { value: '', label: 'Todas las personas' },
              ...persons.map(person => ({ value: person.id, label: person.full_name }))
            ]}
            value={filterPersonId}
            onChange={(value) => setFilterPersonId(value || '')}
            clearable
            style={{ minWidth: '200px' }}
          />
        </Flex>

        {/* Statistics */}
        <Group mb="lg">
          <Text size="sm" color="dimmed">
            Total: {filteredAccounts.length} cuenta{filteredAccounts.length !== 1 ? 's' : ''}
          </Text>
          {(searchTerm || filterPersonId) && (
            <Text size="sm" color="dimmed">
              (Filtrado de {bankAccounts.length} total{bankAccounts.length !== 1 ? 'es' : ''})
            </Text>
          )}
        </Group>

        {/* Bank Accounts Table */}
        <Box style={{ overflowX: 'auto' }}>
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Persona</Table.Th>
                <Table.Th>Banco</Table.Th>
                <Table.Th>Tipo de Cuenta</Table.Th>
                <Table.Th>Número de Cuenta</Table.Th>
                <Table.Th>Titular</Table.Th>
                <Table.Th>Acciones</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {paginatedAccounts.length === 0 ? (
                <Table.Tr>
                  <Table.Td colSpan={6}>
                    <Text ta="center" color="dimmed" py="xl">
                      {bankAccounts.length === 0 
                        ? 'No hay cuentas bancarias registradas' 
                        : 'No se encontraron cuentas que coincidan con los filtros'
                      }
                    </Text>
                  </Table.Td>
                </Table.Tr>
              ) : (
                paginatedAccounts.map((account) => (
                  <Table.Tr key={account.id}>
                    <Table.Td>
                      <Text fw={500}>{account.person_name}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text>{account.bank_name}</Text>
                    </Table.Td>
                    <Table.Td>
                      {getAccountTypeBadge(account.account_type)}
                    </Table.Td>
                    <Table.Td>
                      <Text ff="monospace">{account.account_number}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text>{account.account_holder}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Group gap="xs">
                        <Tooltip label="Ver detalles">
                          <ActionIcon 
                            variant="light" 
                            color="blue"
                            onClick={() => handleOpenModal('view', account)}
                          >
                            <IconEye size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="Editar">
                          <ActionIcon 
                            variant="light" 
                            color="yellow"
                            onClick={() => handleOpenModal('edit', account)}
                          >
                            <IconEdit size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="Eliminar">
                          <ActionIcon 
                            variant="light" 
                            color="red"
                            onClick={() => handleDelete(account)}
                          >
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Tooltip>
                      </Group>
                    </Table.Td>
                  </Table.Tr>
                ))
              )}
            </Table.Tbody>
          </Table>
        </Box>

        {/* Pagination */}
        {totalPages > 1 && (
          <Group justify="center" mt="lg">
            <Pagination
              value={currentPage}
              onChange={setCurrentPage}
              total={totalPages}
              size="sm"
            />
          </Group>
        )}

        {/* Modal for Create/Edit/View */}
        <Modal
          opened={modalOpened}
          onClose={handleCloseModal}
          title={
            modalMode === 'create' ? 'Crear Cuenta Bancaria' :
            modalMode === 'edit' ? 'Editar Cuenta Bancaria' :
            'Detalles de la Cuenta Bancaria'
          }
          size="md"
        >
          <Box pos="relative">
            <LoadingOverlay visible={operationLoading} />
            
            {modalMode === 'view' && selectedAccount ? (
              <Stack gap="md">
                <Group>
                  <Text fw={500}>Persona:</Text>
                  <Text>{bankAccounts.find(a => a.id === selectedAccount.id)?.person_name}</Text>
                </Group>
                <Group>
                  <Text fw={500}>Banco:</Text>
                  <Text>{selectedAccount.bank_name}</Text>
                </Group>
                <Group>
                  <Text fw={500}>Tipo de Cuenta:</Text>
                  {getAccountTypeBadge(selectedAccount.account_type)}
                </Group>
                <Group>
                  <Text fw={500}>Número de Cuenta:</Text>
                  <Text ff="monospace">{selectedAccount.account_number}</Text>
                </Group>
                <Group>
                  <Text fw={500}>Titular:</Text>
                  <Text>{selectedAccount.account_holder}</Text>
                </Group>
                <Group justify="flex-end" mt="lg">
                  <Button variant="default" onClick={handleCloseModal}>
                    Cerrar
                  </Button>
                  <Button onClick={() => {
                    setModalMode('edit');
                    form.setValues({
                      person_id: selectedAccount.person_id,
                      bank_name: selectedAccount.bank_name,
                      account_type: selectedAccount.account_type,
                      account_number: selectedAccount.account_number,
                      account_holder: selectedAccount.account_holder,
                    });
                  }}>
                    Editar
                  </Button>
                </Group>
              </Stack>
            ) : (
              <form onSubmit={form.onSubmit(handleSubmit)}>
                <Stack gap="md">
                  <Select
                    label="Persona"
                    placeholder="Seleccione una persona"
                    required
                    data={persons.map(person => ({ value: person.id, label: person.full_name }))}
                    searchable
                    disabled={modalMode === 'view'}
                    {...form.getInputProps('person_id')}
                  />
                  
                  <TextInput
                    label="Nombre del Banco"
                    placeholder="Ej: Banco de Colombia"
                    required
                    disabled={modalMode === 'view'}
                    {...form.getInputProps('bank_name')}
                  />
                  
                  <Select
                    label="Tipo de Cuenta"
                    placeholder="Seleccione el tipo"
                    required
                    data={[
                      { value: 'Savings', label: 'Ahorros' },
                      { value: 'Checking', label: 'Corriente' },
                      { value: 'Ahorros', label: 'Ahorros' },
                      { value: 'Corriente', label: 'Corriente' },
                    ]}
                    disabled={modalMode === 'view'}
                    {...form.getInputProps('account_type')}
                  />
                  
                  <TextInput
                    label="Número de Cuenta"
                    placeholder="1234567890"
                    required
                    disabled={modalMode === 'view'}
                    {...form.getInputProps('account_number')}
                  />
                  
                  <TextInput
                    label="Titular de la Cuenta"
                    placeholder="Nombre completo del titular"
                    required
                    disabled={modalMode === 'view'}
                    {...form.getInputProps('account_holder')}
                  />

                  <Group justify="flex-end" mt="lg">
                    <Button variant="default" onClick={handleCloseModal}>
                      Cancelar
                    </Button>
                    {modalMode !== 'view' && (
                      <Button type="submit" loading={operationLoading}>
                        {modalMode === 'create' ? 'Crear' : 'Actualizar'}
                      </Button>
                    )}
                  </Group>
                </Stack>
              </form>
            )}
          </Box>
        </Modal>
      </Paper>
    </Container>
  );
} 