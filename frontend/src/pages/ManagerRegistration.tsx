import { useState } from 'react';
import { 
  Container, 
  Paper, 
  Title, 
  TextInput, 
  Button, 
  Group, 
  Stack,
  NumberInput,
  Select,
  MultiSelect,
  Divider,
  Text,
  Stepper,
  Box,
  Loader,
  Alert,
  PasswordInput,
  Textarea
} from '@mantine/core';
import { DateInput } from '@mantine/dates';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconAlertCircle } from '@tabler/icons-react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const API_URL = import.meta.env.VITE_API_URL || '';

const utilityOptions = [
  { value: 'water', label: 'Agua' },
  { value: 'electricity', label: 'Electricidad' },
  { value: 'gas', label: 'Gas' },
  { value: 'internet', label: 'Internet' },
  { value: 'cable', label: 'Cable/TV' },
  { value: 'trash', label: 'Basura' },
  { value: 'sewer', label: 'Alcantarillado' },
  { value: 'landscaping', label: 'Jardinería' },
  { value: 'hoa', label: 'Cuota HOA' },
  { value: 'pool', label: 'Mantenimiento de piscina' },
];

const propertyTypeOptions = [
  { value: 'apartment', label: 'Apartamento' },
  { value: 'house', label: 'Casa' },
  { value: 'condo', label: 'Condominio' },
  { value: 'townhouse', label: 'Casa adosada' },
  { value: 'studio', label: 'Estudio' },
  { value: 'other', label: 'Otro' },
];

const accountTypeOptions = [
  { value: 'checking', label: 'Cuenta corriente' },
  { value: 'savings', label: 'Cuenta de ahorros' },
];

export default function ManagerRegistration() {
  const [active, setActive] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const { user } = useAuth();
  const navigate = useNavigate();

  const handleStepChange = (nextStep: number) => {
    const currentValidation = validateStep(active);
    if (currentValidation || nextStep < active) {
      setActive(nextStep);
    }
  };

  const form = useForm({
    initialValues: {
      // Person information
      full_name: '',
      phone: '',
      nit: '',
      email: '',
      password: '',
      confirm_password: '',

      // Property information
      property_address: '',
      property_apt_number: '',
      property_city: '',
      property_state: '',
      property_zip_code: '',
      property_type: '',

      // Bank account information
      bank_name: '',
      account_type: '',
      account_number: '',
      account_holder: '',

      // Pricing information
      monthly_rent: 0,
      security_deposit: 0,
      utilities_included: [] as string[],
      tenant_responsible_for: [] as string[],
      late_fee: 0,
      due_day: 1,

      // Rental information
      start_date: new Date(),
      end_date: new Date(new Date().setFullYear(new Date().getFullYear() + 1)),
      payment_terms: '',
    },
    validate: {
      full_name: (value) => (!value ? 'El nombre es obligatorio' : null),
      phone: (value) => (!value ? 'El teléfono es obligatorio' : null),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Email inválido'),
      password: (value) => (value.length < 8 ? 'La contraseña debe tener al menos 8 caracteres' : null),
      confirm_password: (value, values) => (value !== values.password ? 'Las contraseñas no coinciden' : null),
      
      property_address: (value) => (!value ? 'La dirección es obligatoria' : null),
      property_city: (value) => (!value ? 'La ciudad es obligatoria' : null),
      property_state: (value) => (!value ? 'El estado/provincia es obligatorio' : null),
      property_type: (value) => (!value ? 'El tipo de propiedad es obligatorio' : null),
      
      bank_name: (value) => (!value ? 'El nombre del banco es obligatorio' : null),
      account_type: (value) => (!value ? 'El tipo de cuenta es obligatorio' : null),
      account_number: (value) => (!value ? 'El número de cuenta es obligatorio' : null),
      account_holder: (value) => (!value ? 'El titular de la cuenta es obligatorio' : null),
      
      monthly_rent: (value) => (value <= 0 ? 'La renta mensual debe ser mayor que cero' : null),
      security_deposit: (value) => (value < 0 ? 'El depósito de seguridad no puede ser negativo' : null),
      due_day: (value) => (value < 1 || value > 31 ? 'El día de vencimiento debe estar entre 1 y 31' : null),
    },
  });

  const validateStep = (step: number): boolean => {
    switch (step) {
      case 0: // Personal information
        return !form.validateField('full_name').hasError &&
               !form.validateField('phone').hasError &&
               !form.validateField('email').hasError &&
               !form.validateField('password').hasError &&
               !form.validateField('confirm_password').hasError;
      case 1: // Property information
        return !form.validateField('property_address').hasError &&
               !form.validateField('property_city').hasError &&
               !form.validateField('property_state').hasError &&
               !form.validateField('property_type').hasError;
      case 2: // Bank account information
        return !form.validateField('bank_name').hasError &&
               !form.validateField('account_type').hasError &&
               !form.validateField('account_number').hasError &&
               !form.validateField('account_holder').hasError;
      case 3: // Pricing information
        return !form.validateField('monthly_rent').hasError &&
               !form.validateField('security_deposit').hasError &&
               !form.validateField('due_day').hasError;
      default:
        return true;
    }
  };

  const nextStep = () => {
    if (validateStep(active) && active < 4) {
      setActive((current) => current + 1);
    }
  };

  const prevStep = () => {
    setActive((current) => (current > 0 ? current - 1 : current));
  };

  const handleSubmit = async (values: typeof form.values) => {
    // Validate all steps before submitting
    for (let i = 0; i < 4; i++) {
      if (!validateStep(i)) {
        handleStepChange(i);
        return;
      }
    }

    if (!user || !user.token) {
      setError('No hay una sesión activa. Por favor inicie sesión nuevamente.');
      setTimeout(() => navigate('/login'), 2000);
      return;
    }

    setLoading(true);
    setError('');

    // Format data for API
    const payload = {
      // Person information
      full_name: values.full_name,
      phone: values.phone,
      nit: values.nit,
      email: values.email,
      password: values.password,

      // Property information
      property_address: values.property_address,
      property_apt_number: values.property_apt_number,
      property_city: values.property_city,
      property_state: values.property_state,
      property_zip_code: values.property_zip_code,
      property_type: values.property_type,

      // Bank account information
      bank_name: values.bank_name,
      account_type: values.account_type,
      account_number: values.account_number,
      account_holder: values.account_holder,

      // Pricing information
      monthly_rent: values.monthly_rent,
      security_deposit: values.security_deposit,
      utilities_included: values.utilities_included,
      tenant_responsible_for: values.tenant_responsible_for,
      late_fee: values.late_fee,
      due_day: values.due_day,

      // Rental information
      start_date: values.start_date,
      end_date: values.end_date,
      payment_terms: values.payment_terms,
    };

    try {
      const response = await axios.post(
        `${API_URL}/api/register/manager`, 
        payload,
        {
          headers: {
            Authorization: `Bearer ${user.token}`
          }
        }
      );
      
      if (response.status === 201) {
        notifications.show({
          title: 'Registro exitoso',
          message: 'Su solicitud ha sido recibida y está pendiente de aprobación.',
          color: 'green',
          icon: <IconCheck size={16} />,
        });
        setSuccess(true);
      } else {
        setError('Hubo un problema al procesar su solicitud.');
      }
    } catch (err) {
      console.error('Registration error:', err);
      if (axios.isAxiosError(err) && err.response) {
        setError(err.response.data.error || 'Error en el servidor. Intente de nuevo más tarde.');
      } else {
        setError('Error de conexión. Verifique su conexión a internet e intente de nuevo.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <Container size="md" my={40}>
        <Paper radius="md" p="xl" withBorder>
          <Stack align="center" gap="lg">
            <IconCheck size={50} color="green" />
            <Title order={2} ta="center">¡Registro exitoso!</Title>
            <Text ta="center">
              Su solicitud de registro como administrador ha sido recibida correctamente. El estado de su cuenta es actualmente "pendiente".
              Un administrador revisará su información y aprobará su cuenta pronto.
            </Text>
            <Text ta="center">
              Una vez aprobada, recibirá un correo electrónico de confirmación y podrá iniciar sesión en el sistema.
            </Text>
            <Button component={Link} to="/login" mt="lg">
              Ir a la página de inicio de sesión
            </Button>
          </Stack>
        </Paper>
      </Container>
    );
  }

  return (
    <Container size="lg" my={40}>
      <Paper radius="md" p="xl" withBorder>
        <Title order={2} ta="center" mb="lg">
          Registro de Administrador de Propiedades
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

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stepper active={active} onStepClick={handleStepChange} mb="xl">
            <Stepper.Step label="Información personal" description="Datos personales">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Nombre completo"
                  placeholder="Juan Pérez"
                  required
                  {...form.getInputProps('full_name')}
                />
                <TextInput
                  label="Teléfono"
                  placeholder="+57 300 123 4567"
                  required
                  {...form.getInputProps('phone')}
                />
                <TextInput
                  label="NIT/Identificación fiscal"
                  placeholder="123456789"
                  {...form.getInputProps('nit')}
                />
                <TextInput
                  label="Email"
                  placeholder="correo@ejemplo.com"
                  required
                  {...form.getInputProps('email')}
                />
                <PasswordInput
                  label="Contraseña"
                  placeholder="Su contraseña"
                  required
                  {...form.getInputProps('password')}
                />
                <PasswordInput
                  label="Confirmar contraseña"
                  placeholder="Confirme su contraseña"
                  required
                  {...form.getInputProps('confirm_password')}
                />
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Propiedad" description="Datos de la propiedad">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Dirección"
                  placeholder="Calle 123"
                  required
                  {...form.getInputProps('property_address')}
                />
                <TextInput
                  label="Apartamento/Unidad"
                  placeholder="Apto 123"
                  {...form.getInputProps('property_apt_number')}
                />
                <Group grow>
                  <TextInput
                    label="Ciudad"
                    placeholder="Bogotá"
                    required
                    {...form.getInputProps('property_city')}
                  />
                  <TextInput
                    label="Estado/Provincia"
                    placeholder="Cundinamarca"
                    required
                    {...form.getInputProps('property_state')}
                  />
                </Group>
                <TextInput
                  label="Código postal"
                  placeholder="110111"
                  {...form.getInputProps('property_zip_code')}
                />
                <Select
                  label="Tipo de propiedad"
                  placeholder="Seleccione el tipo de propiedad"
                  data={propertyTypeOptions}
                  required
                  {...form.getInputProps('property_type')}
                />
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Cuenta bancaria" description="Información bancaria">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Nombre del banco"
                  placeholder="Banco Nacional"
                  required
                  {...form.getInputProps('bank_name')}
                />
                <Select
                  label="Tipo de cuenta"
                  placeholder="Seleccione el tipo de cuenta"
                  data={accountTypeOptions}
                  required
                  {...form.getInputProps('account_type')}
                />
                <TextInput
                  label="Número de cuenta"
                  placeholder="12345678901234"
                  required
                  {...form.getInputProps('account_number')}
                />
                <TextInput
                  label="Titular de la cuenta"
                  placeholder="Juan Pérez"
                  required
                  {...form.getInputProps('account_holder')}
                />
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Precios" description="Información de precios">
              <Stack gap="md" mt="md">
                <Group grow>
                  <NumberInput
                    label="Renta mensual"
                    placeholder="0.00"
                    required
                    min={0}
                    withAsterisk
                    hideControls
                    {...form.getInputProps('monthly_rent')}
                  />
                  <NumberInput
                    label="Depósito de seguridad"
                    placeholder="0.00"
                    min={0}
                    hideControls
                    {...form.getInputProps('security_deposit')}
                  />
                </Group>
                <NumberInput
                  label="Cargo por pago tardío"
                  placeholder="0.00"
                  min={0}
                  hideControls
                  {...form.getInputProps('late_fee')}
                />
                <NumberInput
                  label="Día de vencimiento del pago"
                  placeholder="1"
                  min={1}
                  max={31}
                  required
                  {...form.getInputProps('due_day')}
                />
                <MultiSelect
                  label="Servicios incluidos"
                  placeholder="Seleccione los servicios incluidos en la renta"
                  data={utilityOptions}
                  {...form.getInputProps('utilities_included')}
                />
                <MultiSelect
                  label="Servicios a cargo del inquilino"
                  placeholder="Seleccione los servicios a cargo del inquilino"
                  data={utilityOptions}
                  {...form.getInputProps('tenant_responsible_for')}
                />
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Condiciones" description="Términos del contrato">
              <Stack gap="md" mt="md">
                <Group grow>
                  <DateInput
                    label="Fecha de inicio"
                    placeholder="Seleccione una fecha"
                    value={form.values.start_date}
                    onChange={(value) => form.setFieldValue('start_date', value || new Date())}
                  />
                  <DateInput
                    label="Fecha de fin"
                    placeholder="Seleccione una fecha"
                    value={form.values.end_date}
                    onChange={(value) => form.setFieldValue('end_date', value || new Date())}
                  />
                </Group>
                <Textarea
                  label="Términos de pago"
                  placeholder="Especifique los términos de pago"
                  {...form.getInputProps('payment_terms')}
                />
                <Box mt="xl">
                  <Divider my="md" />
                  <Text size="sm" color="dimmed" ta="center" mb="md">
                    Al enviar este formulario, usted confirma que toda la información proporcionada es verdadera y correcta.
                    Su solicitud será revisada por un administrador y recibirá una notificación cuando su cuenta sea activada.
                  </Text>
                </Box>
              </Stack>
            </Stepper.Step>
          </Stepper>

          <Group justify="apart" mt="xl">
            <Button variant="default" onClick={prevStep} disabled={active === 0}>
              Atrás
            </Button>
            {active === 4 ? (
              <Button type="submit" disabled={loading}>
                {loading ? <Loader size="sm" /> : 'Enviar solicitud'}
              </Button>
            ) : (
              <Button onClick={nextStep}>
                Siguiente
              </Button>
            )}
          </Group>
        </form>
      </Paper>
    </Container>
  );
} 