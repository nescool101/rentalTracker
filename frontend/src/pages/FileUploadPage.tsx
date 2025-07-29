import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { 
  Container, 
  Paper, 
  Title, 
  Stack, 
  Group, 
  Button, 
  Text,
  Alert,
  Progress,
  Badge,
  Card,
  SimpleGrid,
  Center,
  Loader
} from '@mantine/core';
import { Dropzone, FileWithPath, MIME_TYPES } from '@mantine/dropzone';
import { useQuery, useMutation } from '@tanstack/react-query';
import { notifications } from '@mantine/notifications';
import { 
  IconUpload, 
  IconFile, 
  IconCheck, 
  IconX, 
  IconAlertCircle,
  IconFileTypePdf,
  IconPhoto,
  IconCloudUpload,
  IconLogin
} from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import axios from 'axios';

interface UploadResponse {
  success: boolean;
  file?: {
    id: string;
    name: string;
    web_view_link: string;
    mime_type: string;
    size: number;
  };
  message: string;
  share_link?: string;
}

// API cliente público
const publicApiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// API para subida de archivos
const uploadApiClient = axios.create({
  baseURL: '/api',
});

const uploadApi = {
  validateToken: async (token: string) => {
    const response = await publicApiClient.get(`/upload/validate-token/${token}`);
    return response.data;
  },

  uploadFile: async (file: File, token: string, folderName?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('token', token);
    if (folderName) {
      formData.append('folder_name', folderName);
    }

    const response = await uploadApiClient.post('/upload/file', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  uploadFileAuthenticated: async (file: File, authToken: string) => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await uploadApiClient.post('/upload/file-authenticated', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
        'Authorization': `Bearer ${authToken}`,
      },
    });
    return response.data;
  }
};

function FileUploadPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const { user, isAuthenticated } = useAuth();
  const [selectedFiles, setSelectedFiles] = useState<FileWithPath[]>([]);
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});
  const [uploadedFiles, setUploadedFiles] = useState<UploadResponse[]>([]);

  // Query para validar token
  const { data: tokenData, isLoading: isValidating, error: tokenError } = useQuery({
    queryKey: ['validateToken', token],
    queryFn: () => uploadApi.validateToken(token!),
    enabled: !!token,
    retry: false,
  });

  // Mutation para subir archivos con token
  const uploadMutation = useMutation({
    mutationFn: ({ file, token }: { file: File; token: string }) => 
      uploadApi.uploadFile(file, token),
    onSuccess: (data, variables) => {
      if (data.success) {
        setUploadedFiles(prev => [...prev, data]);
        notifications.show({
          title: '✅ Archivo Subido',
          message: `${variables.file.name} se subió correctamente`,
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        // Remover archivo de la lista
        setSelectedFiles(prev => prev.filter(f => f.name !== variables.file.name));
      } else {
        notifications.show({
          title: '❌ Error',
          message: data.message || 'Error subiendo archivo',
          color: 'red',
          icon: <IconX size={16} />,
        });
      }
      // Limpiar progreso
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[variables.file.name];
        return newProgress;
      });
    },
    onError: (error: any, variables) => {
      notifications.show({
        title: '❌ Error',
        message: error.response?.data?.error || 'Error subiendo archivo',
        color: 'red',
        icon: <IconX size={16} />,
      });
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[variables.file.name];
        return newProgress;
      });
    },
  });

  // Mutation para subir archivos autenticados (sin token)
  const uploadAuthenticatedMutation = useMutation({
    mutationFn: ({ file, authToken }: { file: File; authToken: string }) => 
      uploadApi.uploadFileAuthenticated(file, authToken),
    onSuccess: (data, variables) => {
      if (data.success) {
        setUploadedFiles(prev => [...prev, data]);
        notifications.show({
          title: '✅ Archivo Subido',
          message: `${variables.file.name} se subió correctamente`,
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        // Remover archivo de la lista
        setSelectedFiles(prev => prev.filter(f => f.name !== variables.file.name));
      } else {
        notifications.show({
          title: '❌ Error',
          message: data.message || 'Error subiendo archivo',
          color: 'red',
          icon: <IconX size={16} />,
        });
      }
      // Limpiar progreso
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[variables.file.name];
        return newProgress;
      });
    },
    onError: (error: any, variables) => {
      notifications.show({
        title: '❌ Error',
        message: error.response?.data?.error || 'Error subiendo archivo',
        color: 'red',
        icon: <IconX size={16} />,
      });
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[variables.file.name];
        return newProgress;
      });
    },
  });

  // Tipos de archivo permitidos
  const acceptedTypes = [
    MIME_TYPES.pdf,
    MIME_TYPES.jpeg,
    MIME_TYPES.png,
    MIME_TYPES.gif,
    'image/webp',
    'image/jpg'
  ];

  // Validar archivos
  const validateFiles = (files: FileWithPath[]) => {
    const validFiles: FileWithPath[] = [];
    const invalidFiles: string[] = [];

    files.forEach(file => {
      const isValidType = acceptedTypes.includes(file.type);
      const isValidSize = file.size <= 10 * 1024 * 1024; // 10MB

      if (!isValidType) {
        invalidFiles.push(`${file.name}: Tipo de archivo no permitido`);
      } else if (!isValidSize) {
        invalidFiles.push(`${file.name}: Archivo muy grande (máximo 10MB)`);
      } else {
        validFiles.push(file);
      }
    });

    if (invalidFiles.length > 0) {
      notifications.show({
        title: '⚠️ Archivos Rechazados',
        message: invalidFiles.join('\n'),
        color: 'yellow',
        icon: <IconAlertCircle size={16} />,
      });
    }

    return validFiles;
  };

  const handleFilesDrop = (files: FileWithPath[]) => {
    const validFiles = validateFiles(files);
    setSelectedFiles(prev => [...prev, ...validFiles]);
  };

  const handleUploadFiles = () => {
    if (selectedFiles.length === 0) return;

    selectedFiles.forEach(file => {
      setUploadProgress(prev => ({ ...prev, [file.name]: 0 }));
      
      if (token) {
        // Upload with token
        uploadMutation.mutate({ file, token });
      } else if (user?.token || localStorage.getItem('auth_token')) {
        // Upload authenticated
        const authToken = user?.token || localStorage.getItem('auth_token') || '';
        uploadAuthenticatedMutation.mutate({ file, authToken });
      }
    });
  };

  const removeFile = (fileName: string) => {
    setSelectedFiles(prev => prev.filter(f => f.name !== fileName));
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFileIcon = (type: string) => {
    if (type === 'application/pdf') {
      return <IconFileTypePdf size={20} color="red" />;
    } else if (type.startsWith('image/')) {
      return <IconPhoto size={20} color="blue" />;
    }
    return <IconFile size={20} />;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('es-ES', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Si no hay token y no está autenticado
  if (!token && !isAuthenticated) {
    return (
      <Container size="md" mt={50}>
        <Alert icon={<IconLogin size={16} />} color="orange">
          <Text fw={500}>Autenticación Requerida</Text>
          <Text size="sm">Debe iniciar sesión para subir archivos o usar un enlace de acceso válido.</Text>
          <Group mt="md">
            <Button component="a" href={`/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`}>
              Iniciar Sesión
            </Button>
          </Group>
        </Alert>
      </Container>
    );
  }

  // Si hay token, validar
  if (token) {
    // Si está validando
    if (isValidating) {
      return (
        <Container size="md" mt={50}>
          <Center py={50}>
            <Stack align="center">
              <Loader size="lg" />
              <Text>Validando acceso...</Text>
            </Stack>
          </Center>
        </Container>
      );
    }

    // Si el token es inválido
    if (tokenError) {
      return (
        <Container size="md" mt={50}>
          <Alert icon={<IconX size={16} />} color="red">
            <Text fw={500}>Token inválido</Text>
            <Text size="sm">El token de acceso no es válido o ha expirado.</Text>
          </Alert>
        </Container>
      );
    }

    // Si el token es válido pero el usuario no coincide
    if (tokenData && user && tokenData.email !== user.email) {
      return (
        <Container size="md" mt={50}>
          <Alert icon={<IconAlertCircle size={16} />} color="yellow">
            <Text fw={500}>Usuario no autorizado</Text>
            <Text size="sm">Este enlace de subida no está autorizado para su cuenta.</Text>
            <Text size="sm">Enlace destinado a: <strong>{tokenData.email}</strong></Text>
            <Text size="sm">Usuario actual: <strong>{user.email}</strong></Text>
          </Alert>
        </Container>
      );
    }
  }

  return (
    <Container size="lg" mt={20}>
      <Stack gap="lg">
        {/* Header */}
        <Paper p="xl" withBorder>
          <Stack gap="md">
            <Group justify="space-between">
              <Title order={2}>📁 Subir Documentos</Title>
              <Badge color="green" variant="light">
                {token ? 'Acceso con Token' : 'Acceso Autenticado'}
              </Badge>
            </Group>
            
            {token && tokenData ? (
              <Group>
                <Text>
                  <strong>Destinatario:</strong> {tokenData.name}
                </Text>
                <Text c="dimmed">•</Text>
                <Text>
                  <strong>Email:</strong> {tokenData.email}
                </Text>
                <Text c="dimmed">•</Text>
                <Text>
                  <strong>Válido hasta:</strong> {formatDate(tokenData.expires_at)}
                </Text>
              </Group>
            ) : (
              <Group>
                <Text>
                  <strong>Usuario:</strong> {user?.email}
                </Text>
                <Text c="dimmed">•</Text>
                <Text>
                  <strong>Modo:</strong> Subida autenticada
                </Text>
              </Group>
            )}

            <Alert icon={<IconAlertCircle size={16} />} color="blue" variant="light">
              <Text size="sm">
                <strong>Tipos de archivo permitidos:</strong> PDF, JPG, JPEG, PNG, GIF, WEBP (máximo 20MB por archivo)
              </Text>
            </Alert>
          </Stack>
        </Paper>

        {/* Zona de subida */}
        <Paper p="xl" withBorder>
          <Stack gap="md">
            <Title order={3}>📤 Seleccionar Archivos</Title>
            
            <Dropzone
              onDrop={handleFilesDrop}
              accept={acceptedTypes}
              maxSize={20 * 1024 * 1024}
              multiple
            >
              <Group justify="center" gap="xl" mih={220} style={{ pointerEvents: 'none' }}>
                <Dropzone.Accept>
                  <IconUpload size={52} color="var(--mantine-color-blue-6)" stroke={1.5} />
                </Dropzone.Accept>
                <Dropzone.Reject>
                  <IconX size={52} color="var(--mantine-color-red-6)" stroke={1.5} />
                </Dropzone.Reject>
                <Dropzone.Idle>
                  <IconCloudUpload size={52} color="var(--mantine-color-dimmed)" stroke={1.5} />
                </Dropzone.Idle>

                <div>
                  <Text size="xl" inline>
                    Arrastra archivos aquí o haz clic para seleccionar
                  </Text>
                  <Text size="sm" c="dimmed" inline mt={7}>
                    Solo archivos PDF e imágenes, máximo 20MB por archivo
                  </Text>
                </div>
              </Group>
            </Dropzone>

            {/* Lista de archivos seleccionados */}
            {selectedFiles.length > 0 && (
              <Stack gap="sm" mt="md">
                <Text fw={500}>Archivos Seleccionados ({selectedFiles.length})</Text>
                <SimpleGrid cols={1} spacing="xs">
                  {selectedFiles.map((file) => (
                    <Card key={file.name} padding="sm" withBorder>
                      <Group justify="space-between">
                        <Group>
                          {getFileIcon(file.type)}
                          <Stack gap={0}>
                            <Text size="sm" fw={500}>{file.name}</Text>
                            <Text size="xs" c="dimmed">{formatFileSize(file.size)}</Text>
                          </Stack>
                        </Group>
                        
                        <Group>
                          {uploadProgress[file.name] !== undefined && (
                            <Progress value={uploadProgress[file.name]} size="sm" w={100} />
                          )}
                          <Button
                            size="xs"
                            variant="light"
                            color="red"
                            onClick={() => removeFile(file.name)}
                            disabled={uploadProgress[file.name] !== undefined}
                          >
                            Quitar
                          </Button>
                        </Group>
                      </Group>
                    </Card>
                  ))}
                </SimpleGrid>

                <Group justify="center" mt="md">
                  <Button
                    leftSection={<IconUpload size={16} />}
                    onClick={handleUploadFiles}
                    disabled={selectedFiles.length === 0 || uploadMutation.isPending || uploadAuthenticatedMutation.isPending}
                    loading={uploadMutation.isPending || uploadAuthenticatedMutation.isPending}
                    size="lg"
                  >
                    Subir {selectedFiles.length} Archivo{selectedFiles.length !== 1 ? 's' : ''}
                  </Button>
                </Group>
              </Stack>
            )}
          </Stack>
        </Paper>

        {/* Archivos subidos */}
        {uploadedFiles.length > 0 && (
          <Paper p="xl" withBorder>
            <Stack gap="md">
              <Title order={3}>✅ Archivos Subidos ({uploadedFiles.length})</Title>
              <SimpleGrid cols={1} spacing="sm">
                {uploadedFiles.map((upload, index) => (
                  <Card key={index} padding="sm" withBorder bg="green.0">
                    <Group justify="space-between">
                      <Group>
                        <IconCheck size={20} color="green" />
                        <Stack gap={0}>
                          <Text size="sm" fw={500}>{upload.file?.name}</Text>
                          <Text size="xs" c="dimmed">
                            {upload.file?.size ? formatFileSize(upload.file.size) : ''} • 
                            Subido exitosamente
                          </Text>
                        </Stack>
                      </Group>
                      
                      {upload.share_link && (
                        <Button
                          size="xs"
                          variant="light"
                          component="a"
                          href={upload.share_link}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          Ver en Drive
                        </Button>
                      )}
                    </Group>
                  </Card>
                ))}
              </SimpleGrid>
            </Stack>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}

export default FileUploadPage; 