import { useState, useEffect } from 'react';
import { 
  Container, 
  Paper, 
  Title, 
  TextInput, 
  Button, 
  Group, 
  Stack,
  Stepper,
  Box,
  Loader,
  Alert,
  PasswordInput,
  Text,
  Divider,
  Textarea,
  NumberInput,
  Checkbox,
  Select
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconAlertCircle } from '@tabler/icons-react';
import { useAuth } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || '';

export default function OnboardingStepper() {
  const [active, setActive] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [departments, setDepartments] = useState<{value: string, label: string}[]>([]);
  const [loadingDepartments, setLoadingDepartments] = useState(false);
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!user) {
      navigate('/login', { replace: true });
    }
  }, [user, navigate]);

  // Fetch departments from Colombia API
  useEffect(() => {
    const fetchDepartments = async () => {
      setLoadingDepartments(true);
      try {
        const response = await axios.get('https://api-colombia.com/api/v1/Department');
        const departmentOptions = response.data.map((dept: any) => ({
          value: dept.id.toString(),
          label: dept.name
        }));
        setDepartments(departmentOptions);
      } catch (error) {
        console.error('Error fetching departments:', error);
        notifications.show({
          title: 'Error',
          message: 'No se pudieron cargar los departamentos. Por favor, intenta nuevamente.',
          color: 'red',
        });
      } finally {
        setLoadingDepartments(false);
      }
    };

    fetchDepartments();
  }, []);

  // If not authenticated, don't render the stepper
  if (!user) {
    return null;
  }

  const form = useForm({
    initialValues: {
      // Personal information
      fullName: '',
      phone: '',
      identificationType: 'CC', // Default to Cédula de Ciudadanía
      identificationNumber: '',
      address: '',
      
      // New password
      password: '',
      confirm_password: '',
      
      // Property information
      propertyName: '',
      propertyAddress: '',
      propertyCity: '',
      propertyDescription: '',
      propertyType: '',
      propertyDepartment: '',
      
      // Bank account information
      bankName: '',
      accountNumber: '',
      accountType: '',
      accountHolderName: '',
      
      // Pricing 
      baseRentalPrice: 0,
      depositAmount: 0,
      lateFeePercentage: 0,
      
      // Terms and conditions
      termsAccepted: false
    },
    validate: {
      fullName: (value) => (value.trim().length < 2 ? 'Ingrese su nombre completo' : null),
      phone: (value) => (value.trim().length < 6 ? 'Ingrese un número de teléfono válido' : null),
      password: (value) => (value.length < 8 ? 'La contraseña debe tener al menos 8 caracteres' : null),
      confirm_password: (value, values) => (value !== values.password ? 'Las contraseñas no coinciden' : null),
      propertyAddress: (value) => (value.trim().length < 5 ? 'Ingrese una dirección válida' : null),
      accountNumber: (value) => (value.trim().length < 5 ? 'Ingrese un número de cuenta válido' : null),
      termsAccepted: (value) => (value === false ? 'Debe aceptar los términos y condiciones' : null),
    },
  });

  const validateStep = (step: number): boolean => {
    switch (step) {
      case 0: // Welcome screen
        return true;
      case 1: // Personal information
        return !form.validateField('fullName').hasError &&
               !form.validateField('phone').hasError;
      case 2: // Password
        return !form.validateField('password').hasError &&
               !form.validateField('confirm_password').hasError;
      case 3: // Property information
        return !form.validateField('propertyAddress').hasError;
      case 4: // Bank account
        return !form.validateField('accountNumber').hasError;
      case 5: // Pricing
        return true;
      case 6: // Terms and conditions
        return !form.validateField('termsAccepted').hasError;
      default:
        return true;
    }
  };

  const nextStep = () => {
    if (validateStep(active) && active < 7) {
      setActive((current) => current + 1);
    }
  };

  const prevStep = () => {
    setActive((current) => (current > 0 ? current - 1 : current));
  };

  const handleSubmit = async (values: typeof form.values) => {
    if (!validateStep(active)) {
      return;
    }
    
    if (active < 7) {
      nextStep();
      return;
    }
    
    // Final step - submit all data
    setLoading(true);
    setError('');
    
    try {
      if (!user || !user.id || !user.token) {
        setError('Se ha perdido la sesión. Por favor, vuelva a iniciar sesión.');
        return;
      }

      // First create/update person record
      try {
        console.log('Creating or updating person for user:', user.email);
        // Create a person object from the form data
        const personData = {
          full_name: values.fullName,
          phone: values.phone,
          identification_type: values.identificationType,
          identification_number: values.identificationNumber,
          address: values.address
        };
        
        if (user.person_id) {
          // Update existing person
          console.log('Updating existing person with ID:', user.person_id);
          await axios.put(
            `${API_URL}/api/persons/${user.person_id}`,
            personData,
            {
              headers: {
                Authorization: `Bearer ${user.token}`,
              },
            }
          );
        } else {
          // Create new person and link to user
          console.log('Creating new person record');
          const response = await axios.post(
            `${API_URL}/api/persons`,
            personData,
            {
              headers: {
                Authorization: `Bearer ${user.token}`,
              },
            }
          );
          
          // Get the person ID from the response
          const personId = response.data.id;
          console.log('New person created with ID:', personId);
          
          // Update user with person_id
          await axios.put(
            `${API_URL}/api/users/${user.id}`,
            {
              id: user.id,
              email: user.email,
              person_id: personId,
              role: user.role
            },
            {
              headers: {
                Authorization: `Bearer ${user.token}`,
              },
            }
          );
        }
      } catch (err) {
        console.error('Error creating/updating person:', err);
        if (axios.isAxiosError(err) && err.response) {
          console.error('Person creation/update error details:', err.response.data);
        }
        // Continue even if person update fails
      }

      // Then create property if needed
      if (values.propertyAddress) {
        try {
          console.log('Creating property for user:', user.email);
          
          // Find the department name from the departments array
          const selectedDepartment = departments.find(dept => dept.value === values.propertyDepartment);
          const departmentName = selectedDepartment ? selectedDepartment.label : '';
          
          const propertyData = {
            name: values.propertyName || 'Mi Propiedad',
            address: values.propertyAddress,
            city: values.propertyCity || '',
            description: values.propertyDescription || '',
            type: values.propertyType || 'Apartment',
            department: departmentName,
            department_id: values.propertyDepartment || '',
            manager_id: user.person_id,
            resident_id: user.person_id,
            status: 'available'
          };
          console.log('Property data:', JSON.stringify(propertyData));
          
          await axios.post(
            `${API_URL}/api/properties`,
            propertyData,
            {
              headers: {
                Authorization: `Bearer ${user.token}`,
              },
            }
          );
          
          console.log('Property created successfully');
        } catch (err) {
          console.error('Error creating property:', err);
          if (axios.isAxiosError(err) && err.response) {
            console.error('Property creation error details:', err.response.data);
          }
          // Continue even if property creation fails
        }
      }

      // Then create bank account if needed
      if (values.accountNumber) {
        try {
          console.log('Creating bank account for person ID:', user.person_id);
          const bankAccountData = {
            bank_name: values.bankName,
            account_number: values.accountNumber,
            account_type: values.accountType || 'Checking',
            account_holder: values.accountHolderName || values.fullName,
            person_id: user.person_id  // Use person_id for bank accounts
          };
          console.log('Bank account data:', JSON.stringify(bankAccountData));
          
          const bankResponse = await axios.post(
            `${API_URL}/api/bank-accounts`,
            bankAccountData,
            {
              headers: {
                Authorization: `Bearer ${user.token}`,
              },
            }
          );
          
          console.log('Bank account created successfully:', bankResponse.data);
        } catch (err) {
          console.error('Error creating bank account:', err);
          if (axios.isAxiosError(err) && err.response) {
            console.error('Bank account creation error details:', err.response.data);
          }
          // Continue even if bank account creation fails
        }
      }

      // Finally, update user with new status and password
      try {
        console.log('Updating user with ID:', user.id);
        const encodedPassword = btoa(values.password);
        console.log('Password to be sent (encoded length):', encodedPassword.length);
        
        const userData = {
          id: user.id,
          email: user.email,
          person_id: user.person_id,
          role: user.role,
          status: 'activenopaid',  // Change from 'newuser' to 'activenopaid'
          password_base64: encodedPassword,
        };
        
        console.log('User update payload:', JSON.stringify({
          ...userData,
          password_base64: '***REDACTED***'
        }));
        
        const response = await axios.put(
          `${API_URL}/api/users/${user.id}`,
          userData,
          {
            headers: {
              Authorization: `Bearer ${user.token}`,
            },
          }
        );

        console.log('User update response status:', response.status);

        if (response.status === 200) {
          // Set up success message
          const successMessage = (
            <>
              <Title order={4} mb="sm">¡Felicidades!</Title>
              <Text>Has creado tu usuario Manager y toda la información requerida.</Text>
              <Text mt="sm">Serás redirigido a la página principal en unos segundos.</Text>
            </>
          );
          
          notifications.show({
            title: '¡Configuración completada!',
            message: 'Su cuenta ha sido actualizada exitosamente.',
            color: 'green',
            icon: <IconCheck size={16} />,
          });
          
          // Show large success notification
          notifications.show({
            title: 'FELICIDADES',
            message: successMessage,
            color: 'green',
            icon: <IconCheck size={20} />,
            autoClose: 5000, // 5 seconds
            withCloseButton: true,
            styles: { 
              root: { maxWidth: '450px' },
              title: { fontSize: '18px', fontWeight: 'bold' }
            },
          });

          // Logout the user
          await logout();
          
          // Redirect to home page directly without attempting login
          setTimeout(() => {
            window.location.href = '/?newManager=true';
          }, 3000);
        } else {
          setError('Hubo un problema al actualizar su información. Código de respuesta: ' + response.status);
        }
      } catch (err) {
        console.error('Error updating user password:', err);
        if (axios.isAxiosError(err) && err.response) {
          const errorMsg = err.response.data.error || 'Error al actualizar la información de usuario.';
          console.error('Error details:', err.response.data);
          setError(`Error al actualizar su contraseña: ${errorMsg}. Por favor, intente una contraseña diferente.`);
        } else {
          setError('Error de conexión al actualizar su contraseña. Verifique su conexión a internet e intente de nuevo.');
        }
      }
    } catch (err) {
      console.error('Error in account setup process:', err);
      setError('Ocurrió un error al configurar su cuenta. Por favor, intente nuevamente.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container size="lg" my={40}>
      <Paper radius="md" p="xl" withBorder>
        <Title order={2} ta="center" mb="lg">
          Complete su perfil
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
          <Stepper active={active} mb="xl">
            <Stepper.Step label="Bienvenido" description="Primeros pasos">
              <Stack gap="md" mt="md">
                <Text size="lg">
                  ¡Bienvenido a la plataforma de gestión de propiedades!
                </Text>
                <Text>
                  Vamos a ayudarle a configurar su cuenta para que pueda comenzar a utilizar
                  todas las funcionalidades de la plataforma.
                </Text>
                <Text>
                  Este proceso solo tomará unos minutos. Necesitaremos completar los siguientes datos:
                </Text>
                <Box pl="md">
                  <Text>• Información personal</Text>
                  <Text>• Contraseña de acceso</Text>
                  <Text>• Información de propiedades</Text>
                  <Text>• Datos bancarios</Text>
                  <Text>• Configuración de precios</Text>
                  <Text>• Términos y condiciones</Text>
                </Box>
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Su usuario ha sido creado con una contraseña temporal. Por seguridad,
                  deberá cambiarla para continuar usando la plataforma.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Información personal" description="Datos personales">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Nombre completo"
                  placeholder="Juan Pérez"
                  required
                  {...form.getInputProps('fullName')}
                />
                <TextInput
                  label="Teléfono"
                  placeholder="+57 300 123 4567"
                  required
                  value={form.values.phone}
                  onChange={(event) => {
                    // Only allow numbers, spaces, and common phone characters
                    const value = event.currentTarget.value.replace(/[^0-9+\-\s()]/g, '');
                    form.setFieldValue('phone', value);
                  }}
                  error={form.errors.phone}
                />
                <Group grow>
                  <Select
                    label="Tipo de identificación"
                    placeholder="Seleccione tipo"
                    data={[
                      { value: 'RC', label: 'RC - Registro Civil' },
                      { value: 'TI', label: 'TI - Tarjeta de Identidad' },
                      { value: 'CC', label: 'CC - Cédula de Ciudadanía' },
                      { value: 'CE', label: 'CE - Cédula de Extranjería' },
                      { value: 'PP', label: 'PP - Pasaporte' },
                      { value: 'PEP', label: 'PEP - Permiso Especial de Permanencia' },
                      { value: 'NIT', label: 'NIT - Número de Identificación Tributaria' },
                    ]}
                    {...form.getInputProps('identificationType')}
                  />
                  <TextInput
                    label="Número de identificación"
                    placeholder="12345678"
                    {...form.getInputProps('identificationNumber')}
                  />
                </Group>
                <TextInput
                  label="Dirección"
                  placeholder="Calle 123 #45-67"
                  {...form.getInputProps('address')}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Esta información nos ayudará a personalizar su experiencia y a contactarle
                  cuando sea necesario.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Seguridad" description="Contraseña">
              <Stack gap="md" mt="md">
                <PasswordInput
                  label="Nueva contraseña"
                  placeholder="Su nueva contraseña"
                  required
                  {...form.getInputProps('password')}
                />
                <PasswordInput
                  label="Confirmar contraseña"
                  placeholder="Confirme su nueva contraseña"
                  required
                  {...form.getInputProps('confirm_password')}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Su contraseña debe tener al menos 8 caracteres. Use una combinación
                  de letras, números y símbolos para mayor seguridad.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Propiedades" description="Información de propiedad">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Nombre de la propiedad"
                  placeholder="Apartamento Centro"
                  {...form.getInputProps('propertyName')}
                />
                <TextInput
                  label="Dirección de la propiedad"
                  placeholder="Calle 123 #45-67"
                  required
                  {...form.getInputProps('propertyAddress')}
                />
                <Group grow>
                  <Select
                    label="Departamento"
                    placeholder={loadingDepartments ? "Cargando departamentos..." : "Seleccione el departamento"}
                    data={departments}
                    disabled={loadingDepartments}
                    searchable
                    {...form.getInputProps('propertyDepartment')}
                  />
                  <TextInput
                    label="Ciudad"
                    placeholder="Ingrese ciudad"
                    {...form.getInputProps('propertyCity')}
                  />
                </Group>
                <Select
                  label="Tipo de propiedad"
                  placeholder="Seleccione el tipo"
                  data={[
                    { value: 'Apartment', label: 'Apartamento' },
                    { value: 'House', label: 'Casa' },
                    { value: 'Room', label: 'Habitación' },
                    { value: 'Office', label: 'Oficina' },
                    { value: 'Commercial', label: 'Local comercial' },
                    { value: 'Other', label: 'Otro' },
                  ]}
                  {...form.getInputProps('propertyType')}
                />
                <Textarea
                  label="Descripción"
                  placeholder="Apartamento de 2 habitaciones, 1 baño..."
                  minRows={3}
                  {...form.getInputProps('propertyDescription')}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Puede agregar más propiedades más adelante desde el panel de administración.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Cuenta bancaria" description="Datos bancarios">
              <Stack gap="md" mt="md">
                <TextInput
                  label="Nombre del banco"
                  placeholder="Banco de Colombia"
                  {...form.getInputProps('bankName')}
                />
                <TextInput
                  label="Número de cuenta"
                  placeholder="1234567890"
                  required
                  {...form.getInputProps('accountNumber')}
                />
                <Select
                  label="Tipo de cuenta"
                  placeholder="Seleccione el tipo"
                  data={[
                    { value: 'Checking', label: 'Cuenta Corriente' },
                    { value: 'Savings', label: 'Cuenta de Ahorros' },
                  ]}
                  {...form.getInputProps('accountType')}
                />
                <TextInput
                  label="Titular de la cuenta"
                  placeholder="Nombre del titular"
                  {...form.getInputProps('accountHolderName')}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Esta cuenta se utilizará para los pagos y transferencias relacionados con sus propiedades.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Precios" description="Configuración de tarifas">
              <Stack gap="md" mt="md">
                <NumberInput
                  label="Precio base de alquiler"
                  placeholder="1000000"
                  suffix=" COP"
                  {...form.getInputProps('baseRentalPrice')}
                />
                <NumberInput
                  label="Monto de depósito"
                  placeholder="2000000"
                  suffix=" COP"
                  {...form.getInputProps('depositAmount')}
                />
                <NumberInput
                  label="Porcentaje de mora por pagos tardíos"
                  placeholder="5"
                  suffix="%"
                  min={0}
                  max={100}
                  {...form.getInputProps('lateFeePercentage')}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Estos valores se utilizarán como referencia para sus propiedades.
                  Puede personalizar los precios para cada propiedad más adelante.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Términos y condiciones" description="Aceptar términos">
              <Stack gap="md" mt="md">
                <Text size="lg">
                  Términos y Condiciones de Uso
                </Text>
                <Box
                  style={{ 
                    maxHeight: '300px', 
                    overflow: 'auto', 
                    padding: '15px',
                    border: '1px solid #eee',
                    borderRadius: '8px',
                    backgroundColor: '#f9f9f9'
                  }}
                >
                  <Text size="sm">
                    <Text component="div" fw={700} mb={8}>Términos y Condiciones de Uso</Text>
                    
                    <Text component="div" mb={16}>
                      Al utilizar nuestra plataforma de gestión de propiedades, usted acepta cumplir con
                      nuestros términos y condiciones, que incluyen:
                    </Text>
                    
                    <Box component="ul" pl={16} mb={16}>
                      <Box component="li" mb={4}>Proporcionar información precisa y veraz sobre usted y sus propiedades.</Box>
                      <Box component="li" mb={4}>Mantener la confidencialidad de su cuenta y contraseña.</Box>
                      <Box component="li" mb={4}>Notificar inmediatamente sobre cualquier uso no autorizado de su cuenta.</Box>
                      <Box component="li" mb={4}>No utilizar la plataforma para actividades ilegales o fraudulentas.</Box>
                      <Box component="li" mb={4}>Cumplir con todas las leyes y regulaciones aplicables.</Box>
                    </Box>
                    
                    <Text component="div" mb={16}>
                      Nos reservamos el derecho de suspender o terminar su acceso a la plataforma si consideramos 
                      que ha violado estos términos. El uso de esta plataforma implica el pago de tarifas según el 
                      plan seleccionado. Los pagos no son reembolsables una vez procesados.
                    </Text>
                    
                    <Text component="div" fw={700} mb={8}>Consulta y Reporte en Centrales de Riesgo</Text>
                    
                    <Text component="div" mb={16}>
                      Usted autoriza expresamente a nuestra empresa para consultar y reportar su información financiera, 
                      crediticia, comercial y de servicios ante las centrales de riesgo autorizadas en Colombia, tales como 
                      DataCrédito y TransUnion, en los términos establecidos por la Ley 1266 de 2008 y la Ley 2157 de 2021.
                    </Text>
                    
                    <Text component="div" mb={16}>
                      Esta autorización es necesaria para evaluar su comportamiento crediticio y cumplir con las obligaciones 
                      legales relacionadas con la prevención del fraude y la gestión del riesgo.
                    </Text>
                    
                    <Text component="div" mb={16}>
                      En cumplimiento de la normativa vigente, se le informará previamente mediante comunicación escrita con 
                      al menos veinte (20) días de antelación antes de realizar cualquier reporte negativo a las centrales de 
                      riesgo, otorgándole la oportunidad de ejercer sus derechos de defensa y rectificación, conforme a lo 
                      dispuesto en el artículo 12 de la Ley 1266 de 2008.
                    </Text>
                    
                    <Text component="div" mb={16}>
                      Usted tiene derecho a conocer, actualizar y rectificar su información personal contenida en las bases 
                      de datos de las centrales de riesgo, de acuerdo con lo establecido en la Ley 1266 de 2008 y la Ley 1581 de 2012.
                    </Text>
                    
                    <Text component="div" fw={700} mb={8}>Superintendencia de Industria y Comercio</Text>
                    
                    <Text component="div" mb={16}>
                      Para más información sobre sus derechos y las políticas de tratamiento de datos personales, le 
                      recomendamos consultar la página oficial de la Superintendencia de Industria y Comercio: 
                      <Text component="a" href="https://www.sic.gov.co/sobre-el-habeas-data-financiero" target="_blank" c="blue"> 
                        https://www.sic.gov.co/sobre-el-habeas-data-financiero
                      </Text>.
                    </Text>
                  </Text>
                </Box>
                <Checkbox
                  label="He leído y acepto los términos y condiciones"
                  required
                  {...form.getInputProps('termsAccepted', { type: 'checkbox' })}
                />
                <Divider my="md" />
                <Text size="sm" color="dimmed">
                  Es necesario aceptar los términos y condiciones para utilizar la plataforma.
                </Text>
              </Stack>
            </Stepper.Step>
            
            <Stepper.Step label="Finalizar" description="Completar registro">
              <Stack gap="md" mt="md">
                <Text size="lg">
                  ¡Todo listo!
                </Text>
                <Text>
                  Ha completado todos los pasos necesarios para configurar su cuenta.
                  Pulse el botón "Completar" para finalizar y comenzar a utilizar la plataforma.
                </Text>
                <Alert color="blue">
                  <Text>
                    Después de completar este proceso, su cuenta quedará en estado de espera
                    de aprobación por parte del administrador. 
                  </Text>
                  <Text fw={500} mt="xs">
                    Podrá acceder al sistema, pero algunas funcionalidades estarán limitadas
                    hasta que su cuenta sea activada completamente.
                  </Text>
                </Alert>
                <Box mt="xl">
                  <Divider my="md" />
                  <Text size="sm" color="dimmed" ta="center" mb="md">
                    Para cualquier consulta, comuníquese con nuestro equipo de soporte
                    a través de nescool101@gmail.com
                  </Text>
                </Box>
              </Stack>
            </Stepper.Step>
          </Stepper>

          <Group justify="apart" mt="xl">
            <Button variant="default" onClick={prevStep} disabled={active === 0}>
              Atrás
            </Button>
            {active === 7 ? (
              <Button type="submit" disabled={loading}>
                {loading ? <Loader size="sm" /> : 'Completar'}
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