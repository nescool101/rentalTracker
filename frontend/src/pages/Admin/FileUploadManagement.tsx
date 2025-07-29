import { useState } from 'react';
import { useForm } from '@mantine/form';
import { 
  Container, 
  Paper, 
  Title, 
  Stack, 
  Group, 
  Button, 
  TextInput, 
  NumberInput,
  Table,
  Badge,
  Text,
  Tabs,
  Alert,
  ActionIcon,
  Tooltip,
  Select
} from '@mantine/core';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notifications } from '@mantine/notifications';
import { 
  IconUpload, 
  IconMail, 
  IconCheck, 
  IconX, 
  IconClock,
  IconExternalLink,
  IconRefresh,
  IconInfoCircle,
  IconUser
} from '@tabler/icons-react';
import { authService } from '../../services/authService';
import axios from 'axios';
import { userApi } from '../../api/apiService';
import type { User } from '../../types';

// Tipos
interface UploadToken {
  token: string;
  email: string;
  name: string;
  created_at: string;
  expires_at: string;
  used: boolean;
}

interface GenerateUploadLinkRequest {
  recipient_email: string;
  recipient_name: string;
  user_id: string;
  expiration_days: number;
}

// API cliente
const apiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor para token de autenticación
apiClient.interceptors.request.use(
  (config) => {
    const token = authService.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// API functions
const fileUploadApi = {
  generateUploadLink: async (data: GenerateUploadLinkRequest) => {
    const response = await apiClient.post('/admin/file-upload/generate-link', data);
    return response.data;
  },

  getUploadTokens: async () => {
    const response = await apiClient.get('/admin/file-upload/tokens');
    return response.data;
  }
};

function FileUploadManagement() {
  const [activeTab, setActiveTab] = useState<string>('generate');
  const [isGenerating, setIsGenerating] = useState(false);
  const queryClient = useQueryClient();

  // Form para generar enlace
  const form = useForm<GenerateUploadLinkRequest>({
    initialValues: {
      recipient_email: '',
      recipient_name: '',
      user_id: '',
      expiration_days: 7,
    },
    validate: {
      recipient_email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Email inválido'),
      recipient_name: (value) => (value.length < 2 ? 'Nombre debe tener al menos 2 caracteres' : null),
      user_id: (value) => (value.length === 0 ? 'Debe seleccionar un usuario' : null),
      expiration_days: (value) => (value < 1 || value > 365 ? 'Debe ser entre 1 y 365 días' : null),
    },
  });

  // Query para obtener tokens
  const { data: tokensData, isLoading: isLoadingTokens, refetch: refetchTokens } = useQuery({
    queryKey: ['uploadTokens'],
    queryFn: fileUploadApi.getUploadTokens,
    enabled: activeTab === 'tokens',
  });

  // Query para obtener usuarios
  const { data: users, isLoading: isLoadingUsers } = useQuery({
    queryKey: ['users'],
    queryFn: userApi.getAll,
    enabled: activeTab === 'generate',
  });

  // Mutation para generar enlace
  const generateLinkMutation = useMutation({
    mutationFn: fileUploadApi.generateUploadLink,
    onSuccess: (data) => {
      notifications.show({
        title: '✅ Enlace Generado',
        message: `Enlace de subida enviado a ${data.recipient}`,
        color: 'green',
        icon: <IconCheck size={16} />,
      });
      form.reset();
      queryClient.invalidateQueries({ queryKey: ['uploadTokens'] });
    },
    onError: (error: any) => {
      notifications.show({
        title: '❌ Error',
        message: error.response?.data?.error || 'Error generando enlace',
        color: 'red',
        icon: <IconX size={16} />,
      });
    },
    onSettled: () => {
      setIsGenerating(false);
    },
  });

  const handleGenerateLink = (values: GenerateUploadLinkRequest) => {
    setIsGenerating(true);
    generateLinkMutation.mutate(values);
  };

  // Handler para cambio de usuario
  const handleUserChange = (userId: string | null) => {
    if (userId && users) {
      const selectedUser = users.find((user: User) => user.id === userId);
      if (selectedUser) {
        form.setFieldValue('user_id', userId);
        form.setFieldValue('recipient_email', selectedUser.email);
        // Buscar el nombre de la persona asociada si existe
        if (selectedUser.person_id) {
          // Aquí podrías hacer otra query para obtener la información de la persona
          // Por ahora usamos el email como nombre
          form.setFieldValue('recipient_name', selectedUser.email);
        }
      }
    }
  };

  // Preparar opciones de usuarios
  const userOptions = users?.map((user: User) => ({
    value: user.id,
    label: `${user.email} (${user.role})`,
  })) || [];

  // Formatear fecha
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('es-ES', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Estado del token
  const getTokenStatus = (token: UploadToken) => {
    const now = new Date();
    const expiresAt = new Date(token.expires_at);
    
    if (token.used) {
      return { label: 'Usado', color: 'blue' };
    } else if (now > expiresAt) {
      return { label: 'Expirado', color: 'red' };
    } else {
      return { label: 'Activo', color: 'green' };
    }
  };

  return (
    <Container size="xl" mt={20}>
      <Stack gap="lg">
        <Group justify="space-between">
          <Title order={2}>📁 Gestión de Subida de Archivos</Title>
          <Group>
            <Badge color="blue" variant="light">
              Solo PDFs e Imágenes
            </Badge>
            <Badge color="green" variant="light">
              Admin/Manager
            </Badge>
          </Group>
        </Group>

        <Alert icon={<IconInfoCircle size={16} />} color="blue" variant="light">
          <Text size="sm">
            <strong>Información:</strong> Los archivos se suben directamente a Google Drive. 
            Los enlaces generados tienen validez temporal y son únicos por destinatario.
          </Text>
        </Alert>

        <Tabs value={activeTab} onChange={(value) => setActiveTab(value || 'generate')}>
          <Tabs.List>
            <Tabs.Tab value="generate" leftSection={<IconMail size={16} />}>
              Generar Enlace
            </Tabs.Tab>
            <Tabs.Tab value="tokens" leftSection={<IconClock size={16} />}>
              Tokens Activos
            </Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="generate" pt="lg">
            <Paper p="xl" withBorder>
              <Stack gap="md">
                <Title order={3}>📤 Generar Enlace de Subida</Title>
                <Text c="dimmed">
                  Envía un enlace seguro a una persona para que pueda subir documentos al sistema.
                </Text>

                <form onSubmit={form.onSubmit(handleGenerateLink)}>
                  <Stack gap="md">
                    <Select
                      label="Seleccionar Usuario"
                      placeholder={isLoadingUsers ? "Cargando usuarios..." : "Seleccione un usuario"}
                      data={userOptions}
                      searchable
                      clearable
                      leftSection={<IconUser size={16} />}
                      disabled={isLoadingUsers}
                      required
                      {...form.getInputProps('user_id')}
                      onChange={handleUserChange}
                    />

                    <Group grow>
                      <TextInput
                        label="Nombre del Destinatario"
                        placeholder="Se completa automáticamente"
                        required
                        disabled
                        {...form.getInputProps('recipient_name')}
                      />
                      <TextInput
                        label="Email del Destinatario"
                        placeholder="Se completa automáticamente"
                        type="email"
                        required
                        disabled
                        {...form.getInputProps('recipient_email')}
                      />
                    </Group>

                    <NumberInput
                      label="Días de Validez"
                      description="Número de días que el enlace estará activo"
                      placeholder="7"
                      min={1}
                      max={365}
                      {...form.getInputProps('expiration_days')}
                    />

                    <Group justify="flex-end" mt="md">
                      <Button
                        type="submit"
                        leftSection={<IconUpload size={16} />}
                        loading={isGenerating}
                        disabled={!form.isValid()}
                      >
                        Generar y Enviar Enlace
                      </Button>
                    </Group>
                  </Stack>
                </form>
              </Stack>
            </Paper>
          </Tabs.Panel>

          <Tabs.Panel value="tokens" pt="lg">
            <Paper p="xl" withBorder>
              <Stack gap="md">
                <Group justify="space-between">
                  <Title order={3}>🎫 Tokens de Subida</Title>
                  <Tooltip label="Actualizar lista">
                    <ActionIcon
                      variant="light"
                      onClick={() => refetchTokens()}
                      loading={isLoadingTokens}
                    >
                      <IconRefresh size={16} />
                    </ActionIcon>
                  </Tooltip>
                </Group>

                {isLoadingTokens ? (
                  <Text ta="center" py="xl">Cargando tokens...</Text>
                ) : tokensData?.tokens?.length > 0 ? (
                  <Table striped highlightOnHover>
                    <Table.Thead>
                      <Table.Tr>
                        <Table.Th>Destinatario</Table.Th>
                        <Table.Th>Email</Table.Th>
                        <Table.Th>Creado</Table.Th>
                        <Table.Th>Expira</Table.Th>
                        <Table.Th>Estado</Table.Th>
                        <Table.Th>Token</Table.Th>
                      </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                      {tokensData.tokens.map((token: UploadToken) => {
                        const status = getTokenStatus(token);
                        return (
                          <Table.Tr key={token.token}>
                            <Table.Td>
                              <Text fw={500}>{token.name}</Text>
                            </Table.Td>
                            <Table.Td>
                              <Text size="sm" c="dimmed">{token.email}</Text>
                            </Table.Td>
                            <Table.Td>
                              <Text size="sm">{formatDate(token.created_at)}</Text>
                            </Table.Td>
                            <Table.Td>
                              <Text size="sm">{formatDate(token.expires_at)}</Text>
                            </Table.Td>
                            <Table.Td>
                              <Badge color={status.color} variant="light" size="sm">
                                {status.label}
                              </Badge>
                            </Table.Td>
                            <Table.Td>
                              <Group gap="xs">
                                <Text 
                                  size="xs" 
                                  c="dimmed" 
                                  style={{ fontFamily: 'monospace' }}
                                >
                                  {token.token.substring(0, 16)}...
                                </Text>
                                <ActionIcon
                                  size="sm"
                                  variant="light"
                                  onClick={() => {
                                    navigator.clipboard.writeText(token.token);
                                    notifications.show({
                                      message: 'Token copiado al portapapeles',
                                      color: 'blue',
                                    });
                                  }}
                                >
                                  <IconExternalLink size={12} />
                                </ActionIcon>
                              </Group>
                            </Table.Td>
                          </Table.Tr>
                        );
                      })}
                    </Table.Tbody>
                  </Table>
                ) : (
                  <Text ta="center" py="xl" c="dimmed">
                    No hay tokens generados aún
                  </Text>
                )}
              </Stack>
            </Paper>
          </Tabs.Panel>
        </Tabs>
      </Stack>
    </Container>
  );
}

export default FileUploadManagement; 