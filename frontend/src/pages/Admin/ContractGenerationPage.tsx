import { useState } from 'react';
import { useForm } from '@mantine/form';
import { Box, Button, Grid, Group, Paper, Select, Stack, Textarea, Title, NumberInput, Checkbox, SegmentedControl, Modal, Text } from '@mantine/core';
import { DateInput } from '@mantine/dates';
import { notifications } from '@mantine/notifications';
import { useQuery } from '@tanstack/react-query';
import { IconFileTypePdf, IconAlertCircle, IconCheck, IconSignature } from '@tabler/icons-react';
import { personApi, propertyApi, contractSigningApi } from '../../api/apiService';
import axios from 'axios';
import { Person, Property } from '../../types';
import { authService } from '../../services/authService';

// Get the apiClient from the apiService
const apiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token in all requests
apiClient.interceptors.request.use(
  (config) => {
    const token = authService.getToken();
    console.log('Auth token from authService:', token);
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

const ContractGenerationPage = () => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [generatedContractId, setGeneratedContractId] = useState<string | null>(null);
  const [isSignatureModalOpen, setIsSignatureModalOpen] = useState(false);
  const [isRequestingSignature, setIsRequestingSignature] = useState(false);

  // Load people and properties
  const { data: persons, isLoading: isLoadingPersons } = useQuery({
    queryKey: ['personsForContract'],
    queryFn: personApi.getAll,
  });

  const { data: properties, isLoading: isLoadingProperties } = useQuery({
    queryKey: ['propertiesForContract'],
    queryFn: propertyApi.getAll,
  });

  const form = useForm({
    initialValues: {
      renter_id: '',
      owner_id: '',
      property_id: '',
      cosigner_id: '',
      witness_id: '',
      start_date: null as Date | null,
      end_date: null as Date | null,
      contract_duration: '6', // Default to 6 months
      monthly_rent: 0,
      requires_deposit: false,
      deposit_amount: 0,
      deposit_text: 'El depósito será utilizado para cubrir cualquier renta no pagada o daños a la propiedad.',
      additional_info: '',
    },
    validate: {
      renter_id: (value) => (value ? null : 'Arrendatario es requerido'),
      owner_id: (value) => (value ? null : 'Arrendador es requerido'),
      property_id: (value) => (value ? null : 'Propiedad es requerida'),
      start_date: (value) => (!value ? 'Fecha de inicio es requerida' : null),
      monthly_rent: (value) => (value <= 0 ? 'El precio de renta mensual es requerido' : null),
      deposit_amount: (value, values) => 
        (values.requires_deposit && value <= 0) ? 'El monto del depósito es requerido' : null,
    },
  });

  // For this implementation, we'll determine person roles differently
  // Assuming we have some way to differentiate renters and owners
  // This is a placeholder - we would need to adapt based on how roles are defined in the system
  const renters = persons || [];
  const owners = persons || [];
  const allPersons = persons || [];

  // Calculate end date based on start date and duration
  const calculateEndDate = (startDate: Date | null, durationMonths: number): Date | null => {
    if (!startDate) return null;
    const endDate = new Date(startDate);
    endDate.setMonth(endDate.getMonth() + durationMonths);
    return endDate;
  };

  // Handle duration change
  const handleDurationChange = (value: string) => {
    form.setFieldValue('contract_duration', value);
    const durationMonths = parseInt(value, 10);
    const endDate = calculateEndDate(form.values.start_date, durationMonths);
    form.setFieldValue('end_date', endDate);
  };

  // Handle start date change
  const handleStartDateChange = (date: Date | null) => {
    form.setFieldValue('start_date', date);
    if (date) {
      const durationMonths = parseInt(form.values.contract_duration, 10);
      const endDate = calculateEndDate(date, durationMonths);
      form.setFieldValue('end_date', endDate);
    }
  };

  const handleSubmit = async (values: typeof form.values) => {
    setIsSubmitting(true);

    try {
      // Send request to generate PDF
      const response = await apiClient.post('/contracts/generate', values, {
        responseType: 'blob',
      });

      // Extract contract ID from response headers if available
      const contractId = response.headers['x-contract-id'];
      if (contractId) {
        setGeneratedContractId(contractId);
      }

      // Create a blob URL for the PDF
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      
      // Create a link and trigger download
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'contrato_arrendamiento.pdf');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      notifications.show({
        title: 'Contrato generado',
        message: 'El contrato ha sido generado exitosamente',
        color: 'green',
        icon: <IconCheck />,
      });

      // Don't reset form here so we can send for signature
    } catch (error) {
      console.error('Error generating contract:', error);
      let errorMessage = 'Ha ocurrido un error al generar el contrato.';
      
      if (axios.isAxiosError(error) && error.response) {
        errorMessage = error.response.data.error || errorMessage;
      }
      
      notifications.show({
        title: 'Error',
        message: errorMessage,
        color: 'red',
        icon: <IconAlertCircle />,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRequestSignature = async () => {
    if (!generatedContractId || !form.values.renter_id) {
      notifications.show({
        title: 'Error',
        message: 'No se puede solicitar firma sin generar un contrato primero',
        color: 'red',
        icon: <IconAlertCircle />,
      });
      return;
    }

    setIsRequestingSignature(true);
    try {
      await contractSigningApi.requestSignature({
        contract_id: generatedContractId,
        recipient_id: form.values.renter_id,
        expiration_days: 7
      });

      notifications.show({
        title: 'Solicitud de firma enviada',
        message: 'El correo electrónico con la solicitud de firma ha sido enviado correctamente',
        color: 'green',
        icon: <IconCheck />,
      });

      // Reset form only after successful signature request
      form.reset();
      setGeneratedContractId(null);
      setIsSignatureModalOpen(false);
    } catch (error) {
      console.error('Error requesting signature:', error);
      let errorMessage = 'Ha ocurrido un error al solicitar la firma.';
      
      if (axios.isAxiosError(error) && error.response) {
        errorMessage = error.response.data.error || errorMessage;
      }
      
      notifications.show({
        title: 'Error',
        message: errorMessage,
        color: 'red',
        icon: <IconAlertCircle />,
      });
    } finally {
      setIsRequestingSignature(false);
    }
  };

  return (
    <Box>
      <Title order={2} mb="md">Generación de Contratos de Arrendamiento</Title>
      
      <Paper shadow="xs" p="md" withBorder>
        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <Grid>
              <Grid.Col span={6}>
                <Select
                  label="Arrendatario"
                  description="Persona que alquilará la propiedad"
                  data={renters?.map((renter: Person) => ({ value: renter.id, label: `${renter.full_name} (${renter.nit})` })) || []}
                  placeholder="Seleccione el arrendatario"
                  searchable
                  withAsterisk
                  disabled={isLoadingPersons}
                  {...form.getInputProps('renter_id')}
                />
              </Grid.Col>
              
              <Grid.Col span={6}>
                <Select
                  label="Arrendador"
                  description="Propietario de la propiedad"
                  data={owners?.map((owner: Person) => ({ value: owner.id, label: `${owner.full_name} (${owner.nit})` })) || []}
                  placeholder="Seleccione el arrendador"
                  searchable
                  withAsterisk
                  disabled={isLoadingPersons}
                  {...form.getInputProps('owner_id')}
                />
              </Grid.Col>
            </Grid>
            
            <Grid>
              <Grid.Col span={12}>
                <Select
                  label="Propiedad"
                  description="Propiedad que será arrendada"
                  data={properties?.map((property: Property) => ({ value: property.id, label: `${property.address} (${property.type})` })) || []}
                  placeholder="Seleccione la propiedad"
                  searchable
                  withAsterisk
                  disabled={isLoadingProperties}
                  {...form.getInputProps('property_id')}
                />
              </Grid.Col>
            </Grid>
            
            <Grid>
              <Grid.Col span={6}>
                <Select
                  label="Deudor Solidario (Opcional)"
                  description="Persona que responderá como garantía adicional"
                  data={allPersons?.map((person: Person) => ({ value: person.id, label: `${person.full_name} (${person.nit})` })) || []}
                  placeholder="Seleccione el deudor solidario (opcional)"
                  searchable
                  disabled={isLoadingPersons}
                  clearable
                  {...form.getInputProps('cosigner_id')}
                />
              </Grid.Col>
              
              <Grid.Col span={6}>
                <Select
                  label="Testigo (Opcional)"
                  description="Persona que será testigo del contrato"
                  data={allPersons?.map((person: Person) => ({ value: person.id, label: `${person.full_name} (${person.nit})` })) || []}
                  placeholder="Seleccione el testigo (opcional)"
                  searchable
                  disabled={isLoadingPersons}
                  clearable
                  {...form.getInputProps('witness_id')}
                />
              </Grid.Col>
            </Grid>
            
            <Grid>
              <Grid.Col span={6}>
                <DateInput
                  label="Fecha de Inicio"
                  description="Fecha en que inicia el contrato"
                  placeholder="Seleccione la fecha de inicio"
                  withAsterisk
                  value={form.values.start_date}
                  onChange={handleStartDateChange}
                  error={form.errors.start_date}
                />
              </Grid.Col>
              
              <Grid.Col span={6}>
                <Stack gap="xs">
                  <Title order={6}>Duración del Contrato</Title>
                  <SegmentedControl
                    data={[
                      { label: '6 Meses', value: '6' },
                      { label: '12 Meses', value: '12' },
                      { label: '24 Meses', value: '24' },
                    ]}
                    value={form.values.contract_duration}
                    onChange={handleDurationChange}
                  />
                  <DateInput
                    label="Fecha de Finalización"
                    description="Calculada automáticamente según duración"
                    placeholder="Fecha de finalización"
                    value={form.values.end_date}
                    disabled
                  />
                </Stack>
              </Grid.Col>
            </Grid>
            
            <Grid>
              <Grid.Col span={6}>
                <NumberInput
                  label="Renta Mensual"
                  description="Monto a pagar mensualmente"
                  placeholder="Ingrese el monto de la renta"
                  withAsterisk
                  min={0}
                  {...form.getInputProps('monthly_rent')}
                />
              </Grid.Col>
              
              <Grid.Col span={6}>
                <Checkbox
                  label="Requiere Depósito"
                  description="Marque si se requiere un depósito de garantía"
                  checked={form.values.requires_deposit}
                  onChange={(event) => form.setFieldValue('requires_deposit', event.currentTarget.checked)}
                  mt="md"
                />
              </Grid.Col>
            </Grid>
            
            {form.values.requires_deposit && (
              <Grid>
                <Grid.Col span={6}>
                  <NumberInput
                    label="Monto de Depósito"
                    description="Monto del depósito de garantía"
                    placeholder="Ingrese el monto del depósito"
                    withAsterisk
                    min={0}
                    {...form.getInputProps('deposit_amount')}
                  />
                </Grid.Col>
                <Grid.Col span={6}>
                  <Textarea
                    label="Texto del Depósito"
                    description="Descripción de las condiciones del depósito"
                    placeholder="Ingrese el texto sobre las condiciones del depósito"
                    {...form.getInputProps('deposit_text')}
                  />
                </Grid.Col>
              </Grid>
            )}
            
            <Textarea
              label="Información Adicional (Opcional)"
              description="Cláusulas adicionales u observaciones para el contrato"
              placeholder="Ingrese información adicional para el contrato"
              minRows={4}
              maxRows={8}
              {...form.getInputProps('additional_info')}
            />
            
            <Grid mt="md">
              <Grid.Col>
                <Group>
                  <Button
                    leftSection={<IconFileTypePdf />}
                    type="submit"
                    loading={isSubmitting}
                    disabled={isSubmitting}
                  >
                    Generar Contrato PDF
                  </Button>
                  {generatedContractId && (
                    <Button
                      leftSection={<IconSignature />}
                      color="blue"
                      onClick={() => setIsSignatureModalOpen(true)}
                      disabled={isSubmitting}
                    >
                      Solicitar Firma
                    </Button>
                  )}
                </Group>
              </Grid.Col>
            </Grid>
          </Stack>
        </form>
      </Paper>
      
      {/* Signature Request Modal */}
      <Modal
        opened={isSignatureModalOpen}
        onClose={() => setIsSignatureModalOpen(false)}
        title="Solicitar Firma de Contrato"
      >
        <Text size="sm" mb="md">
          Está a punto de enviar una solicitud de firma electrónica al arrendatario. Se le enviará un correo 
          electrónico con un enlace para revisar y firmar el contrato.
        </Text>
        <Text size="sm" mb="md">
          El enlace de firma estará activo por 7 días.
        </Text>
        <Group justify="space-between" mt="xl">
          <Button variant="outline" onClick={() => setIsSignatureModalOpen(false)}>
            Cancelar
          </Button>
          <Button 
            loading={isRequestingSignature}
            onClick={handleRequestSignature}
          >
            Confirmar y Enviar
          </Button>
        </Group>
      </Modal>
    </Box>
  );
};

export default ContractGenerationPage;