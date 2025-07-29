import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Box, Title, Paper, Button, Group, Loader, Alert, Stack, Center, Text } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconAlertCircle, IconX, IconSignature, IconDownload } from '@tabler/icons-react';
import { contractSigningApi } from '../api/apiService';

interface SigningStatusData {
  id: string;
  contract_id: string;
  recipient_id: string;
  status: string;
  status_spanish: string;
  created_at: string;
  expires_at: string;
  signed_at?: string;
}

const ContractSigningPage = () => {
  const { signingId } = useParams<{ signingId: string }>();
  const [loading, setLoading] = useState(true);
  const [signingData, setSigningData] = useState<SigningStatusData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const [isSigning, setIsSigning] = useState(false);
  const [isRejecting, setIsRejecting] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);

  // Fetch signing status on load
  useEffect(() => {
    const fetchSigningStatus = async () => {
      if (!signingId) {
        setError('ID de firma no proporcionado');
        setLoading(false);
        return;
      }

      try {
        const data = await contractSigningApi.getSignatureStatus(signingId);
        setSigningData(data);
        
        // Set PDF URL - using public endpoint that doesn't require authentication
        setPdfUrl(`/api/public/contract-signing/pdf/${signingId}`);
        
      } catch (error) {
        console.error('Error fetching signature status:', error);
        setError('No se pudo obtener la información de firma del contrato.');
      } finally {
        setLoading(false);
      }
    };

    fetchSigningStatus();
  }, [signingId]);

  const handleSign = async () => {
    if (!signingId) return;
    
    setIsSigning(true);
    try {
      await contractSigningApi.signContract(signingId);
      
      notifications.show({
        title: 'Contrato firmado',
        message: 'El contrato ha sido firmado exitosamente.',
        color: 'green',
        icon: <IconCheck />,
      });
      
      // Refresh status
      const data = await contractSigningApi.getSignatureStatus(signingId);
      setSigningData(data);
      
      // Update PDF URL to get the signed version
      setPdfUrl(`/api/public/contract-signing/pdf/${signingId}?signed=true`);
      
    } catch (error) {
      console.error('Error signing contract:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo firmar el contrato. Por favor intente nuevamente.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    } finally {
      setIsSigning(false);
    }
  };

  const handleReject = async () => {
    if (!signingId) return;
    
    setIsRejecting(true);
    try {
      await contractSigningApi.rejectContract(signingId);
      
      notifications.show({
        title: 'Firma rechazada',
        message: 'Ha rechazado la firma del contrato.',
        color: 'yellow',
        icon: <IconX />,
      });
      
      // Refresh status
      const data = await contractSigningApi.getSignatureStatus(signingId);
      setSigningData(data);
      
    } catch (error) {
      console.error('Error rejecting contract:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo rechazar la firma. Por favor intente nuevamente.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    } finally {
      setIsRejecting(false);
    }
  };

  const handleDownloadPDF = () => {
    if (!pdfUrl) return;
    
    setIsDownloading(true);
    try {
      // Create a link and trigger download
      const link = document.createElement('a');
      link.href = pdfUrl;
      link.setAttribute('download', 'contrato.pdf');
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      notifications.show({
        title: 'Descarga iniciada',
        message: 'La descarga del contrato ha comenzado.',
        color: 'blue',
      });
    } catch (error) {
      console.error('Error downloading PDF:', error);
      notifications.show({
        title: 'Error',
        message: 'No se pudo descargar el contrato. Por favor intente nuevamente.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    } finally {
      setIsDownloading(false);
    }
  };

  // Helper to render status-dependent content
  const renderContent = () => {
    if (!signingData) return null;
    
    const status = signingData.status;
    const statusSpanish = signingData.status_spanish || 'Pendiente';
    
    if (status === 'signed') {
      return (
        <Stack>
          <Alert title="Contrato Firmado" color="green" icon={<IconCheck />}>
            Este contrato ya ha sido firmado el {new Date(signingData.signed_at!).toLocaleDateString()}.
          </Alert>
          
          {pdfUrl && (
            <>
              <Box my="md">
                <iframe 
                  src={pdfUrl} 
                  style={{ width: '100%', height: '500px', border: '1px solid #eee' }}
                  title="Contrato de Arrendamiento Firmado"
                />
              </Box>
              
              <Button 
                leftSection={<IconDownload />}
                onClick={handleDownloadPDF}
                loading={isDownloading}
              >
                Descargar Contrato Firmado
              </Button>
            </>
          )}
        </Stack>
      );
    }
    
    if (status === 'rejected') {
      return (
        <Alert title="Firma Rechazada" color="red" icon={<IconX />}>
          Usted ha rechazado la firma de este contrato.
        </Alert>
      );
    }
    
    if (status === 'expired') {
      return (
        <Alert title="Solicitud Expirada" color="yellow" icon={<IconAlertCircle />}>
          Esta solicitud de firma ha expirado. Por favor contacte al remitente para una nueva solicitud.
        </Alert>
      );
    }
    
    // If pending
    return (
      <Stack>
        <Alert title={`Estado: ${statusSpanish}`} color="blue" icon={<IconSignature />}>
          Este contrato está pendiente de su firma. Por favor revise el documento antes de firmar.
        </Alert>
        
        {pdfUrl ? (
          <Box my="md">
            <iframe 
              src={pdfUrl} 
              style={{ width: '100%', height: '500px', border: '1px solid #eee' }}
              title="Contrato de Arrendamiento"
            />
          </Box>
        ) : (
          <Text color="dimmed" ta="center" my="xl">
            No se pudo cargar la vista previa del contrato. Puede proceder a firmarlo o rechazarlo.
          </Text>
        )}
        
        <Group>
          <Button 
            color="red" 
            leftSection={<IconX />}
            onClick={handleReject}
            loading={isRejecting}
          >
            Rechazar
          </Button>
          <Button 
            color="green" 
            leftSection={<IconSignature />}
            onClick={handleSign}
            loading={isSigning}
          >
            Firmar Contrato
          </Button>
        </Group>
      </Stack>
    );
  };

  if (loading) {
    return (
      <Center p="2rem">
        <Loader size="lg" />
      </Center>
    );
  }

  if (error) {
    return (
      <Box>
        <Alert title="Error" color="red" icon={<IconAlertCircle />}>
          {error}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Title order={2} mb="md">Firma de Contrato de Arrendamiento</Title>
      
      <Paper shadow="xs" p="md" withBorder>
        {renderContent()}
      </Paper>
    </Box>
  );
};

export default ContractSigningPage; 