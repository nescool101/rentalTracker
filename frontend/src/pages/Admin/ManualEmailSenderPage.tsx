import React from 'react';
import { Button, Stack, Title, Paper, Text, Textarea, Select, LoadingOverlay } from '@mantine/core';
import { useForm } from '@mantine/form';
import { useMutation, useQuery } from '@tanstack/react-query';
import { notifications } from '@mantine/notifications';
import { IconMail, IconAlertCircle, IconCalendarEvent, IconSend, IconRefreshAlert } from '@tabler/icons-react';
import { utilityApi, personApi } from '../../api/apiService';
import type { Person } from '../../types';

const ManualEmailSenderPage: React.FC = () => {

  const triggerMonthlyRemindersMutation = useMutation<unknown, Error, void>({
    mutationFn: () => utilityApi.triggerEmailNotifications(),
    onSuccess: () => {
      notifications.show({
        title: 'Success',
        message: 'Monthly rent reminder process triggered successfully. Emails will be sent in the background.',
        color: 'green',
        icon: <IconMail />,
      });
    },
    onError: (error) => {
      notifications.show({
        title: 'Error',
        message: error.message || 'Failed to trigger monthly reminders.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    },
  });

  const handleTriggerMonthlyReminders = () => {
    triggerMonthlyRemindersMutation.mutate();
  };

  const annualRenewalForm = useForm<{ optional_message: string }>({
    initialValues: {
      optional_message: '',
    },
  });

  const triggerAnnualRemindersMutation = useMutation<{ message: string }, Error, { optional_message?: string }>({
    mutationFn: (payload) => utilityApi.triggerAnnualRenewalReminders(payload),
    onSuccess: (data) => {
      notifications.show({
        title: 'Process Triggered',
        message: data.message || 'Annual renewal reminder process triggered. Check logs for details.',
        color: 'green',
        icon: <IconRefreshAlert />,
      });
      annualRenewalForm.reset();
    },
    onError: (error) => {
      notifications.show({
        title: 'Error',
        message: error.message || 'Failed to trigger annual renewal reminders.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    },
  });

  const handleSendAnnualRenewal = (values: { optional_message: string }) => {
    triggerAnnualRemindersMutation.mutate({ optional_message: values.optional_message });
  };

  const { data: persons, isLoading: isLoadingPersons } = useQuery<Person[], Error>({
    queryKey: ['personsForEmail'],
    queryFn: personApi.getAll,
  });

  const customEmailForm = useForm<{ recipient_person_id: string; subject: string; body: string }>({
    initialValues: {
      recipient_person_id: '',
      subject: '',
      body: '',
    },
    validate: {
      recipient_person_id: (value) => (value ? null : 'Recipient is required'),
      subject: (value) => (value.trim() ? null : 'Subject is required'),
      body: (value) => (value.trim() ? null : 'Email body is required'),
    },
  });

  const sendCustomEmailMutation = useMutation<{ message: string }, Error, { recipient_person_id: string; subject: string; body: string }>({
    mutationFn: (payload) => utilityApi.sendCustomEmail(payload),
    onSuccess: (data) => {
      notifications.show({
        title: 'Email Sent',
        message: data.message || 'Custom email sent successfully!',
        color: 'green',
        icon: <IconSend />,
      });
      customEmailForm.reset();
    },
    onError: (error) => {
      notifications.show({
        title: 'Error Sending Email',
        message: error.message || 'Failed to send custom email.',
        color: 'red',
        icon: <IconAlertCircle />,
      });
    },
  });

  const handleSendCustomEmail = (values: { recipient_person_id: string; subject: string; body: string }) => {
    sendCustomEmailMutation.mutate(values);
  };

  const personOptions = persons?.map(p => ({ 
    value: p.id, 
    label: `${p.full_name} (ID: ${p.id})` 
  })) || [];

  return (
    <Stack gap="xl">
      <Title order={2}>Manual Email Sender</Title>

      <Paper shadow="sm" p="md" withBorder>
        <Title order={3} mb="md">Monthly Rent Reminders</Title>
        <Text mb="sm">Trigger the process to send monthly rent payment reminders to all relevant tenants.</Text>
        <Text size="xs" c="dimmed" mb="md">This uses the same underlying notification system as the automated scheduler.</Text>
        <Button 
          leftSection={<IconCalendarEvent size={18} />}
          onClick={handleTriggerMonthlyReminders}
          loading={triggerMonthlyRemindersMutation.isPending}
        >
          Trigger Monthly Rent Reminders
        </Button>
      </Paper>

      <Paper shadow="sm" p="md" withBorder>
        <Title order={3} mb="md">Annual Contract Renewal Reminders</Title>
        <Text mb="sm">Trigger the process to send renewal reminders to tenants whose contracts are ending in approximately one month.</Text>
        <form onSubmit={annualRenewalForm.onSubmit(handleSendAnnualRenewal)}>
          <Stack>
            <Textarea
              label="Optional Custom Message Addon"
              placeholder="Include an additional message to append to the standard renewal reminder..."
              {...annualRenewalForm.getInputProps('optional_message')}
              minRows={3}
            />
            <Button 
              type="submit" 
              leftSection={<IconRefreshAlert size={18} />} 
              mt="sm" 
              loading={triggerAnnualRemindersMutation.isPending}
            >
              Trigger Annual Renewal Reminders
            </Button>
          </Stack>
        </form>
      </Paper>

      <Paper shadow="sm" p="md" withBorder>
        <Title order={3} mb="md">Send Custom Email</Title>
        {isLoadingPersons && <LoadingOverlay visible />}
        {!isLoadingPersons && (
          <form onSubmit={customEmailForm.onSubmit(handleSendCustomEmail)}>
            <Stack>
              <Select
                label="Select Recipient"
                placeholder="Choose a person (tenant, manager, user)"
                data={personOptions}
                searchable
                clearable
                nothingFoundMessage="No persons found"
                {...customEmailForm.getInputProps('recipient_person_id')}
                required
              />
              <Textarea
                label="Subject"
                placeholder="Email subject"
                required
                {...customEmailForm.getInputProps('subject')}
              />
              <Textarea
                label="Email Body (HTML allowed)"
                placeholder="Compose your email message... You can use HTML tags for formatting."
                required
                minRows={6}
                {...customEmailForm.getInputProps('body')}
              />
              <Button type="submit" leftSection={<IconSend size={18} />} mt="sm" loading={sendCustomEmailMutation.isPending}>
                Send Custom Email
              </Button>
            </Stack>
          </form>
        )}
      </Paper>

    </Stack>
  );
};

export default ManualEmailSenderPage; 