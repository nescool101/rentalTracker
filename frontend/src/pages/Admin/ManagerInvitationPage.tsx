import { useState } from 'react';
import {
  Container,
  Paper,
  Title,
  TextInput,
  Button,
  Stack,
  Alert,
  Text,
  Textarea,
  List,
  Group,
  CopyButton,
  ActionIcon,
  Tooltip,
  Code
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconAlertCircle, IconMail, IconCopy, IconCopyCheck } from '@tabler/icons-react';
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || '';

export default function ManagerInvitationPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [sentInvitations, setSentInvitations] = useState<Array<{ email: string, tempPassword: string }>>([]);

  const form = useForm({
    initialValues: {
      email: '',
      name: '',
      additionalMessage: '',
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Email inválido'),
      name: (value) => (value.length < 2 ? 'El nombre es obligatorio' : null),
    },
  });

  // Generate a random 8-digit password
  const generateTempPassword = (): string => {
    return Math.floor(10000000 + Math.random() * 90000000).toString();
  };

  const handleSubmit = async (values: typeof form.values) => {
    setLoading(true);
    setError('');
    
    const tempPassword = generateTempPassword();
    const apiUrl = `${API_URL}/api/invitations/manager`;
    console.log(`Sending invitation request to: ${apiUrl}`, values);

    try {
      const response = await axios.post(
        apiUrl,
        {
          email: values.email,
          name: values.name,
          message: values.additionalMessage,
          tempPassword: tempPassword,
          status: 'newuser'
        },
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
          },
        }
      );

      if (response.status === 201 || response.status === 200) {
        notifications.show({
          title: 'Invitación enviada',
          message: `Se ha enviado una invitación a ${values.email}`,
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        
        setSentInvitations([...sentInvitations, { 
          email: values.email, 
          tempPassword: tempPassword 
        }]);
        
        form.reset();
        setSuccess(true);
      } else {
        setError('Hubo un problema al enviar la invitación');
      }
    } catch (err) {
      console.error('Error sending invitation:', err);
      if (axios.isAxiosError(err) && err.response) {
        setError(err.response.data.error || 'Error en el servidor. Intente de nuevo más tarde.');
      } else {
        setError('Error de conexión. Verifique su conexión a internet e intente de nuevo.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container size="md" my={40}>
      <Paper radius="md" p="xl" withBorder>
        <Title order={2} ta="center" mb="lg">
          Invitar Administradores de Propiedades
        </Title>

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

        {success && sentInvitations.length > 0 && (
          <Alert 
            icon={<IconCheck size={16} />} 
            title="Invitaciones enviadas" 
            color="green" 
            mb="md"
          >
            <Text mb="sm">Se han enviado invitaciones a los siguientes usuarios con sus contraseñas temporales:</Text>
            <List>
              {sentInvitations.map((invitation, index) => (
                <List.Item key={index}>
                  <Group>
                    <Text>{invitation.email}</Text>
                    <Text>Contraseña temporal: </Text>
                    <Code>{invitation.tempPassword}</Code>
                    <CopyButton value={invitation.tempPassword} timeout={2000}>
                      {({ copied, copy }) => (
                        <Tooltip label={copied ? 'Copiado' : 'Copiar'} position="right">
                          <ActionIcon color={copied ? 'teal' : 'gray'} onClick={copy}>
                            {copied ? <IconCopyCheck size="1rem" /> : <IconCopy size="1rem" />}
                          </ActionIcon>
                        </Tooltip>
                      )}
                    </CopyButton>
                  </Group>
                </List.Item>
              ))}
            </List>
            <Text mt="md" size="sm" color="dimmed">
              Asegúrese de guardar o compartir estas contraseñas de forma segura. Los usuarios deberán
              cambiarlas al iniciar sesión por primera vez.
            </Text>
          </Alert>
        )}

        <Text color="dimmed" size="sm" mb="lg">
          Envíe invitaciones por correo electrónico a nuevos administradores de propiedades para que se registren en el sistema.
          Se generará una contraseña temporal de 8 dígitos que el administrador deberá utilizar para su primer inicio de sesión.
        </Text>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Nombre del administrador"
              placeholder="Juan Pérez"
              required
              {...form.getInputProps('name')}
            />
            <TextInput
              label="Correo electrónico"
              placeholder="correo@ejemplo.com"
              required
              {...form.getInputProps('email')}
            />
            <Textarea
              label="Mensaje adicional (opcional)"
              placeholder="Bienvenido a nuestra aplicación, use las credenciales que se le enviarán por separado para acceder al sistema."
              minRows={3}
              {...form.getInputProps('additionalMessage')}
            />

            <Button 
              type="submit" 
              mt="md" 
              loading={loading}
              leftSection={<IconMail size={16} />}
            >
              Enviar Invitación
            </Button>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
} 