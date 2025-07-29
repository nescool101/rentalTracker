import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Paper,
  Table,
  Badge,
  Button,
  Group,
  Text,
  ActionIcon,
  Modal,
  Stack,
  Alert,
  LoadingOverlay,
  Anchor,
  Tooltip,
  Select,
  TextInput,
  Card,
  Grid,
  useMantineTheme,
} from '@mantine/core';
import {
  IconCheck,
  IconX,
  IconDownload,
  IconEye,
  IconAlertCircle,
  IconFileText,
  IconPhoto,
  IconFile,
  IconTrash,
  IconCloud,
  IconDatabase,
  IconUser,
  IconFilter,
  IconSearch,
} from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';
import { useDisclosure } from '@mantine/hooks';

interface SupabaseFileInfo {
  name: string;
  size: number;
  path: string;
  mime_type: string;
  uploaded_at: string;
  download_url: string;
}

interface FileManagementResponse {
  success: boolean;
  files: SupabaseFileInfo[];
}

const FileManagement: React.FC = () => {
  const theme = useMantineTheme();
  const [files, setFiles] = useState<SupabaseFileInfo[]>([]);
  const [filteredFiles, setFilteredFiles] = useState<SupabaseFileInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<SupabaseFileInfo | null>(null);
  const [modalOpened, { open: openModal, close: closeModal }] = useDisclosure(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUser, setSelectedUser] = useState<string>('all');

  // Estados para estadísticas
  const [totalFiles, setTotalFiles] = useState(0);
  const [totalSize, setTotalSize] = useState(0);
  const [uniqueUsers, setUniqueUsers] = useState<string[]>([]);

  // Cargar archivos
  const fetchFiles = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('auth_token');
      const response = await fetch('/api/admin/file-upload/files', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Error cargando archivos');
      }

      const data: FileManagementResponse = await response.json();
      const filesList = data.files || [];
      setFiles(filesList);
      setFilteredFiles(filesList);
      
      // Calcular estadísticas
      setTotalFiles(filesList.length);
      setTotalSize(filesList.reduce((sum, file) => sum + file.size, 0));
      
      // Extraer usuarios únicos de las rutas de archivos
      const users = Array.from(new Set(
        filesList.map(file => {
          const pathParts = file.path.split('/');
          return pathParts[0]?.replace('user_', '') || 'unknown';
        })
      ));
      setUniqueUsers(users);
      
    } catch (error) {
      console.error('Error:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudieron cargar los archivos',
        color: 'red',
        icon: <IconX size={16} />,
      });
    } finally {
      setLoading(false);
    }
  };

  // Filtrar archivos
  useEffect(() => {
    let filtered = files;

    // Filtrar por tipo de archivo
    if (filterType !== 'all') {
      filtered = filtered.filter(file => {
        const mimeType = file.mime_type || '';
        switch (filterType) {
          case 'images':
            return mimeType.startsWith('image/');
          case 'documents':
            return mimeType.includes('pdf') || mimeType.includes('doc') || mimeType.includes('text');
          case 'archives':
            return mimeType.includes('zip') || mimeType.includes('rar');
          default:
            return true;
        }
      });
    }

    // Filtrar por usuario
    if (selectedUser !== 'all') {
      filtered = filtered.filter(file => {
        const userFromPath = file.path.split('/')[0]?.replace('user_', '') || 'unknown';
        return userFromPath === selectedUser;
      });
    }

    // Filtrar por término de búsqueda
    if (searchTerm) {
      filtered = filtered.filter(file =>
        file.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    setFilteredFiles(filtered);
  }, [files, filterType, selectedUser, searchTerm]);

  // Eliminar archivo
  const deleteFile = async (filePath: string) => {
    try {
      setActionLoading(filePath);
      const token = localStorage.getItem('auth_token');
      const encodedPath = encodeURIComponent(filePath);
      const response = await fetch(`/api/admin/file-upload/files/${encodedPath}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Error eliminando archivo');
      }

      notifications.show({
        title: '🗑️ Archivo Eliminado',
        message: 'El archivo ha sido eliminado exitosamente',
        color: 'green',
        icon: <IconCheck size={16} />,
      });

      // Actualizar lista
      await fetchFiles();
    } catch (error) {
      console.error('Error:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo eliminar el archivo',
        color: 'red',
        icon: <IconX size={16} />,
      });
    } finally {
      setActionLoading(null);
    }
  };

  // Descargar archivo SIN eliminar
  const downloadFileOnly = async (filePath: string, fileName: string) => {
    try {
      setActionLoading(filePath);
      const token = localStorage.getItem('auth_token');
      const encodedPath = encodeURIComponent(filePath);
      const response = await fetch(`/api/admin/file-upload/files/download-only/${encodedPath}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Error descargando archivo');
      }

      // Crear URL para descargar el archivo
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      notifications.show({
        title: '📥 Archivo Descargado',
        message: 'El archivo se descargó (sin eliminar de Supabase)',
        color: 'blue',
        icon: <IconDownload size={16} />,
      });

    } catch (error) {
      console.error('Error:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo descargar el archivo',
        color: 'red',
        icon: <IconX size={16} />,
      });
    } finally {
      setActionLoading(null);
    }
  };

  // Descargar archivo Y eliminar
  const downloadAndDeleteFile = async (filePath: string, fileName: string) => {
    try {
      setActionLoading(filePath);
      const token = localStorage.getItem('auth_token');
      const encodedPath = encodeURIComponent(filePath);
      const response = await fetch(`/api/admin/file-upload/files/download/${encodedPath}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Error descargando archivo');
      }

      // Crear URL para descargar el archivo
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      notifications.show({
        title: '📥 Archivo Descargado y Eliminado',
        message: 'El archivo se descargó y se eliminó automáticamente de Supabase',
        color: 'green',
        icon: <IconDownload size={16} />,
      });

      // Actualizar lista
      await fetchFiles();
    } catch (error) {
      console.error('Error:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo descargar el archivo',
        color: 'red',
        icon: <IconX size={16} />,
      });
    } finally {
      setActionLoading(null);
    }
  };

  // Formatear tamaño de archivo
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Formatear fecha
  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString('es-ES');
  };

  // Obtener usuario de la ruta del archivo
  const getUserFromPath = (path: string): string => {
    const pathParts = path.split('/');
    return pathParts[0]?.replace('user_', '') || 'unknown';
  };

  // Obtener icono según tipo de archivo
  const getFileIcon = (mimeType: string) => {
    if (mimeType.startsWith('image/')) return <IconPhoto size={16} />;
    if (mimeType.includes('pdf')) return <IconFileText size={16} />;
    return <IconFile size={16} />;
  };

  // Obtener color según tipo de archivo
  const getFileTypeColor = (mimeType: string): string => {
    if (mimeType.startsWith('image/')) return 'blue';
    if (mimeType.includes('pdf')) return 'red';
    if (mimeType.includes('doc')) return 'cyan';
    if (mimeType.includes('zip') || mimeType.includes('rar')) return 'orange';
    return 'gray';
  };

  // Ver detalles del archivo
  const viewFileDetails = (file: SupabaseFileInfo) => {
    setSelectedFile(file);
    openModal();
  };

  useEffect(() => {
    fetchFiles();
  }, []);

  return (
    <Container size="xl" py="md">
      <Title order={2} mb="md">
        <Group>
          <IconCloud size={28} color={theme.colors.blue[6]} />
          Gestión de Archivos - Supabase Storage
        </Group>
      </Title>

      {/* Estadísticas */}
      <Grid mb="md">
        <Grid.Col span={{ base: 12, md: 4 }}>
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Group justify="space-between">
              <div>
                <Text size="sm" c="dimmed">Total de Archivos</Text>
                <Text size="xl" fw={700}>{totalFiles}</Text>
              </div>
              <IconDatabase size={32} color={theme.colors.blue[6]} />
            </Group>
          </Card>
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 4 }}>
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Group justify="space-between">
              <div>
                <Text size="sm" c="dimmed">Espacio Utilizado</Text>
                <Text size="xl" fw={700}>{formatFileSize(totalSize)}</Text>
              </div>
              <IconCloud size={32} color={theme.colors.green[6]} />
            </Group>
          </Card>
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 4 }}>
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Group justify="space-between">
              <div>
                <Text size="sm" c="dimmed">Usuarios Activos</Text>
                <Text size="xl" fw={700}>{uniqueUsers.length}</Text>
              </div>
              <IconUser size={32} color={theme.colors.orange[6]} />
            </Group>
          </Card>
        </Grid.Col>
      </Grid>

      {/* Filtros */}
      <Paper shadow="sm" p="md" mb="md">
        <Group justify="space-between" mb="md">
          <Text fw={500}>Filtros</Text>
          <Button
            variant="subtle"
            size="xs"
            onClick={() => {
              setFilterType('all');
              setSelectedUser('all');
              setSearchTerm('');
            }}
          >
            Limpiar Filtros
          </Button>
        </Group>
        <Group grow>
          <TextInput
            placeholder="Buscar archivos..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            leftSection={<IconSearch size={16} />}
          />
          <Select
            placeholder="Tipo de archivo"
            value={filterType}
            onChange={(value) => setFilterType(value || 'all')}
            data={[
              { value: 'all', label: 'Todos los tipos' },
              { value: 'images', label: 'Imágenes' },
              { value: 'documents', label: 'Documentos' },
              { value: 'archives', label: 'Archivos comprimidos' },
            ]}
            leftSection={<IconFilter size={16} />}
          />
          <Select
            placeholder="Usuario"
            value={selectedUser}
            onChange={(value) => setSelectedUser(value || 'all')}
            data={[
              { value: 'all', label: 'Todos los usuarios' },
              ...uniqueUsers.map(user => ({ value: user, label: user })),
            ]}
            leftSection={<IconUser size={16} />}
          />
        </Group>
      </Paper>

      {/* Alerta informativa */}
      <Alert
        icon={<IconAlertCircle size={16} />}
        title="Información Importante"
        color="blue"
        mb="md"
      >
        <Text size="sm">
          Los archivos están almacenados en <strong>Supabase Storage</strong> con límite de 1GB.
          Cuando un administrador descarga un archivo, se <strong>elimina automáticamente</strong> de Supabase.
        </Text>
      </Alert>

      {/* Tabla de archivos */}
      <Paper shadow="sm" p="md" pos="relative">
        <LoadingOverlay visible={loading} />
        
        <Group justify="space-between" mb="md">
          <Text fw={500}>
            Archivos ({filteredFiles.length}/{totalFiles})
          </Text>
          <Button
            variant="light"
            leftSection={<IconDownload size={16} />}
            onClick={fetchFiles}
            loading={loading}
          >
            Actualizar
          </Button>
        </Group>

        {filteredFiles.length === 0 ? (
          <Text ta="center" c="dimmed" py="xl">
            No hay archivos para mostrar
          </Text>
        ) : (
          <Table.ScrollContainer minWidth={800}>
            <Table verticalSpacing="md" highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Archivo</Table.Th>
                  <Table.Th>Usuario</Table.Th>
                  <Table.Th>Tamaño</Table.Th>
                  <Table.Th>Tipo</Table.Th>
                  <Table.Th>Fecha</Table.Th>
                  <Table.Th>Acciones</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {filteredFiles.map((file) => (
                  <Table.Tr key={file.path}>
                    <Table.Td>
                      <Group gap="sm">
                        {getFileIcon(file.mime_type)}
                        <div>
                          <Text size="sm" fw={500}>
                            {file.name}
                          </Text>
                          <Text size="xs" c="dimmed">
                            {file.path}
                          </Text>
                        </div>
                      </Group>
                    </Table.Td>
                    <Table.Td>
                      <Badge 
                        variant="light" 
                        color="blue"
                        leftSection={<IconUser size={12} />}
                      >
                        {getUserFromPath(file.path)}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">
                        {formatFileSize(file.size)}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge 
                        variant="light" 
                        color={getFileTypeColor(file.mime_type)}
                      >
                        {file.mime_type.split('/')[1]?.toUpperCase() || 'UNKNOWN'}
                      </Badge>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">
                        {formatDate(file.uploaded_at)}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Group gap="xs">
                        <Tooltip label="Ver detalles">
                          <ActionIcon
                            variant="light"
                            color="blue"
                            onClick={() => viewFileDetails(file)}
                          >
                            <IconEye size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="Descargar">
                          <ActionIcon
                            variant="light"
                            color="blue"
                            loading={actionLoading === file.path}
                            onClick={() => downloadFileOnly(file.path, file.name)}
                          >
                            <IconDownload size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="Descargar y Eliminar">
                          <ActionIcon
                            variant="light"
                            color="green"
                            loading={actionLoading === file.path}
                            onClick={() => downloadAndDeleteFile(file.path, file.name)}
                          >
                            <IconDownload size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="Eliminar">
                          <ActionIcon
                            variant="light"
                            color="red"
                            loading={actionLoading === file.path}
                            onClick={() => deleteFile(file.path)}
                          >
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Tooltip>
                      </Group>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>
        )}
      </Paper>

      {/* Modal de detalles */}
      <Modal
        opened={modalOpened}
        onClose={closeModal}
        title={
          <Group>
            <IconFile size={20} />
            <Text fw={500}>Detalles del Archivo</Text>
          </Group>
        }
        size="md"
      >
        {selectedFile && (
          <Stack gap="md">
            <div>
              <Text size="sm" c="dimmed">Nombre del archivo</Text>
              <Text fw={500}>{selectedFile.name}</Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">Ruta completa</Text>
              <Text size="sm" ff="monospace">{selectedFile.path}</Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">Usuario</Text>
              <Badge variant="light" color="blue">
                {getUserFromPath(selectedFile.path)}
              </Badge>
            </div>
            <div>
              <Text size="sm" c="dimmed">Tamaño</Text>
              <Text>{formatFileSize(selectedFile.size)}</Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">Tipo MIME</Text>
              <Badge variant="light" color={getFileTypeColor(selectedFile.mime_type)}>
                {selectedFile.mime_type}
              </Badge>
            </div>
            <div>
              <Text size="sm" c="dimmed">Fecha de subida</Text>
              <Text>{formatDate(selectedFile.uploaded_at)}</Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">URL de descarga</Text>
              <Anchor 
                href={selectedFile.download_url} 
                target="_blank" 
                size="sm"
                style={{ wordBreak: 'break-all' }}
              >
                {selectedFile.download_url}
              </Anchor>
            </div>
                         <Group mt="md">
               <Button
                 variant="light"
                 color="blue"
                 leftSection={<IconDownload size={16} />}
                 onClick={() => {
                   downloadFileOnly(selectedFile.path, selectedFile.name);
                   closeModal();
                 }}
               >
                 Descargar
               </Button>
               <Button
                 variant="light"
                 color="green"
                 leftSection={<IconDownload size={16} />}
                 onClick={() => {
                   downloadAndDeleteFile(selectedFile.path, selectedFile.name);
                   closeModal();
                 }}
               >
                 Descargar y Eliminar
               </Button>
               <Button
                 variant="light"
                 color="red"
                 leftSection={<IconTrash size={16} />}
                 onClick={() => {
                   deleteFile(selectedFile.path);
                   closeModal();
                 }}
               >
                 Solo Eliminar
               </Button>
             </Group>
          </Stack>
        )}
      </Modal>
    </Container>
  );
};

export default FileManagement; 