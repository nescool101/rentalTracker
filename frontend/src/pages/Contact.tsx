import { useState } from 'react';
import { 
  Container, 
  Title, 
  Text, 
  TextInput, 
  Textarea, 
  Button, 
  Group, 
  SimpleGrid, 
  Card, 
  Box,
  ThemeIcon
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { 
  IconPhone, 
  IconAt, 
  IconMapPin, 
  IconCheck 
} from '@tabler/icons-react';

export default function Contact() {
  const [loading, setLoading] = useState(false);
  
  const form = useForm({
    initialValues: {
      name: '',
      email: '',
      phone: '',
      subject: '',
      message: '',
    },
    validate: {
      name: (value) => (value.trim().length < 2 ? 'El nombre es requerido' : null),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Email inválido'),
      message: (value) => (value.trim().length < 10 ? 'El mensaje debe tener al menos 10 caracteres' : null),
    },
  });

  const handleSubmit = (_values: typeof form.values) => {
    setLoading(true);
    
    // Simulate API call
    setTimeout(() => {
      setLoading(false);
      notifications.show({
        title: 'Mensaje enviado',
        message: 'Nos pondremos en contacto contigo pronto',
        color: 'green',
        icon: <IconCheck size={16} />,
      });
      form.reset();
    }, 1000);
  };

  const contactInfo = [
    {
      title: 'Email',
      description: 'info@rentalmanager.com',
      icon: IconAt,
      color: 'blue',
    },
    {
      title: 'Teléfono',
      description: '+57 300 123 4567',
      icon: IconPhone,
      color: 'teal',
    },

    {
      title: 'Dirección',
      description: 'Calle 123 #45-67, Bogotá, Colombia',
      icon: IconMapPin,
      color: 'red',
    },
  ];

  return (
    <Box>
      {/* Hero Section */}
      <Box 
        style={{ 
          backgroundColor: '#f8f9fa',
          padding: '80px 0 60px',
        }}
      >
        <Container size="xl">
          <Title order={1} mb="md" ta="center">Contacto</Title>
          <Text size="lg" mb="xl" ta="center" maw={800} mx="auto">
            Estamos aquí para ayudarte. Contáctanos y un miembro de nuestro equipo te responderá lo antes posible.
          </Text>
        </Container>
      </Box>

      <Container size="xl" py={60}>
        <SimpleGrid cols={{ base: 1, md: 2 }} spacing={50}>
          {/* Contact Form */}
          <div>
            <Title order={2} mb="xl">Envíanos un mensaje</Title>
            
            <form onSubmit={form.onSubmit(handleSubmit)}>
              <TextInput
                label="Nombre"
                placeholder="Tu nombre"
                required
                mb="md"
                {...form.getInputProps('name')}
              />
              
              <TextInput
                label="Email"
                placeholder="tu@email.com"
                required
                mb="md"
                {...form.getInputProps('email')}
              />
              
              <TextInput
                label="Teléfono"
                placeholder="Tu número de teléfono"
                mb="md"
                {...form.getInputProps('phone')}
              />
              
              <TextInput
                label="Asunto"
                placeholder="El asunto de tu mensaje"
                mb="md"
                {...form.getInputProps('subject')}
              />
              
              <Textarea
                label="Mensaje"
                placeholder="Tu mensaje"
                required
                minRows={4}
                mb="xl"
                {...form.getInputProps('message')}
              />
              
              <Button 
                type="submit" 
                loading={loading}
                fullWidth
                size="md"
              >
                Enviar Mensaje
              </Button>
            </form>
          </div>
          
          {/* Contact Information */}
          <div>
            <Title order={2} mb="xl">Información de Contacto</Title>
            
            <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md">
              {contactInfo.map((item, index) => (
                <Card key={index} withBorder padding="lg" radius="md">
                  <Group align="flex-start">
                    <ThemeIcon size={40} radius="md" color={item.color}>
                      <item.icon size={20} />
                    </ThemeIcon>
                    <div>
                      <Text fw={500} size="lg" mb={5}>{item.title}</Text>
                      <Text size="sm">{item.description}</Text>
                    </div>
                  </Group>
                </Card>
              ))}
            </SimpleGrid>
            
            <Card withBorder padding="lg" radius="md" mt="xl">
              <Title order={3} mb="md">Horario de Atención</Title>
              <Text mb="xs">Lunes a Viernes: 9:00 AM - 6:00 PM</Text>
              <Text mb="xs">Sábados: 9:00 AM - 1:00 PM</Text>
              <Text>Domingos y Festivos: Cerrado</Text>
            </Card>
          </div>
        </SimpleGrid>
      </Container>
      
      {/* Map Section */}
      <Box mb={60}>
        <iframe 
          src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d254508.39280650213!2d-74.27348877470106!3d4.648620599999993!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x8e3f9bfd2da6cb29%3A0x239d635520a33914!2zQm9nb3TDoQ!5e0!3m2!1ses!2sco!4v1654789542739!5m2!1ses!2sco" 
          width="100%" 
          height="450" 
          style={{ border: 0 }} 
          allowFullScreen 
          loading="lazy" 
        />
      </Box>
    </Box>
  );
} 