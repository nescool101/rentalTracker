import { Container, Title, Text, Alert } from '@mantine/core';
import { IconAlertCircle } from '@tabler/icons-react';

export default function Owners() {
  return (
    <Container size="xl">
      <Title order={1} mb="lg">Owners</Title>
      
      <Alert icon={<IconAlertCircle size={16} />} title="Coming Soon" color="blue">
        <Text>The Owners management page is currently under development and will be available soon.</Text>
      </Alert>
    </Container>
  );
} 