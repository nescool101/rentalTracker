import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Table, Button, Modal, Group, Title, LoadingOverlay, Alert, Stack, Text, Paper, Badge, TextInput, NumberInput, MultiSelect } from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconTrash, IconCurrencyDollar, IconAlertCircle } from '@tabler/icons-react';

import { rentalApi } from '../../api/apiService';
import { pricingApi } from '../../api/apiService';
import type { Rental, Pricing } from '../../types';

interface RentalWithPricing extends Rental {
  pricing?: Pricing | null;
}

const RentalPricingPage: React.FC = () => {
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingPricing, setEditingPricing] = useState<Pricing | null>(null);
  const [selectedRentalForPricing, setSelectedRentalForPricing] = useState<Rental | null>(null);

  const { data: rentals, isLoading: isLoadingRentals, error: rentalsError } = useQuery<Rental[], Error>({
    queryKey: ['rentals'],
    queryFn: rentalApi.getAll,
  });

  const { data: allPricing, isLoading: isLoadingAllPricing, error: pricingError } = useQuery<Pricing[], Error>({
    queryKey: ['allPricing'],
    queryFn: pricingApi.getAll,
  });

  const rentalsWithPricing: RentalWithPricing[] = React.useMemo(() => {
    if (!rentals || !allPricing) return [];
    return rentals.map((rental: Rental) => ({
      ...rental,
      pricing: allPricing.find((p: Pricing) => p.rental_id === rental.id) || null,
    }));
  }, [rentals, allPricing]);

  const form = useForm<Omit<Pricing, 'id'> & { id?: string }>({
    initialValues: {
      rental_id: '',
      monthly_rent: 0,
      security_deposit: 0,
      utilities_included: [],
      tenant_responsible_for: [],
      late_fee: 0,
      due_day: 1,
    },
    validate: {
      rental_id: (value) => (value && value.trim() !== '' ? null : 'Rental ID is required'),
      monthly_rent: (value) => (value != null && value > 0 ? null : 'Monthly rent must be a positive number'),
      due_day: (value) => (value != null && value >= 1 && value <= 31 ? null : 'Due day must be between 1 and 31'),
      security_deposit: (value) => (value != null && value >=0 ? null : 'Security deposit cannot be negative'),
      late_fee: (value) => (value != null && value >=0 ? null : 'Late fee cannot be negative'),
    },
  });

  const commonMutationOptions = {
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['allPricing'] });
      queryClient.invalidateQueries({ queryKey: ['rentals'] });
      notifications.show({ title: 'Success', message: 'Pricing information saved successfully.', color: 'green' });
      setIsModalOpen(false);
      form.reset();
    },
    onError: (error: Error) => {
      notifications.show({ title: 'Error', message: error.message || 'An unexpected error occurred.', color: 'red' });
    },
  };

  const createPricingMutation = useMutation<Pricing, Error, Omit<Pricing, 'id'>>({
    mutationFn: (pricingData) => pricingApi.create(pricingData),
    ...commonMutationOptions,
  });

  const updatePricingMutation = useMutation<Pricing, Error, Pricing>({
    mutationFn: (pricingData) => pricingApi.update(pricingData.id!, pricingData),
    ...commonMutationOptions,
  });

  const deletePricingMutation = useMutation<void, Error, string>({
    mutationFn: (id) => pricingApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['allPricing'] });
      queryClient.invalidateQueries({ queryKey: ['rentals'] });
      notifications.show({ title: 'Success', message: 'Pricing record deleted.', color: 'green' });
    },
    onError: (error: Error) => {
      notifications.show({ title: 'Error', message: error.message || 'Failed to delete pricing record.', color: 'red' });
    },
  });

  const handleOpenModal = (rental: Rental, pricing?: Pricing | null) => {
    setSelectedRentalForPricing(rental);
    setEditingPricing(pricing || null);
    form.setValues({
      rental_id: rental.id,
      monthly_rent: pricing?.monthly_rent || 0,
      security_deposit: pricing?.security_deposit || 0,
      utilities_included: pricing?.utilities_included || [],
      tenant_responsible_for: pricing?.tenant_responsible_for || [],
      late_fee: pricing?.late_fee || 0,
      due_day: pricing?.due_day || 1,
      id: pricing?.id,
    });
    setIsModalOpen(true);
  };

  const handleSubmit = (values: Omit<Pricing, 'id'> & { id?: string }) => {
    if (!selectedRentalForPricing) return;

    const submissionData = {
      ...values,
      rental_id: selectedRentalForPricing.id,
      monthly_rent: Number(values.monthly_rent) || 0,
      security_deposit: Number(values.security_deposit) || 0,
      late_fee: Number(values.late_fee) || 0,
      due_day: Number(values.due_day) || 1,
      utilities_included: values.utilities_included || [],
      tenant_responsible_for: values.tenant_responsible_for || [],
    };

    if (editingPricing && editingPricing.id) {
      updatePricingMutation.mutate({ ...submissionData, id: editingPricing.id } as Pricing);
    } else {
      const { id, ...createData } = submissionData;
      createPricingMutation.mutate(createData as Omit<Pricing, 'id'>);
    }
  };
  
  const handleDelete = (id: string) => {
    if (window.confirm('Are you sure you want to delete this pricing record?')) {
      deletePricingMutation.mutate(id);
    }
  };

  if (isLoadingRentals || isLoadingAllPricing) {
    return <LoadingOverlay visible />;
  }

  if (rentalsError || pricingError) {
    return (
      <Alert icon={<IconAlertCircle size={16} />} title="Error!" color="red">
        {rentalsError?.message || pricingError?.message || 'Failed to load data.'}
      </Alert>
    );
  }

  const rows = rentalsWithPricing.map((item) => (
    <tr key={item.id}>
      <td>{item.id}</td>
      <td>
        {item.pricing ? (
          <Badge color="green">Configured</Badge>
        ) : (
          <Badge color="orange">Not Set</Badge>
        )}
      </td>
      <td>{item.pricing ? `$${item.pricing.monthly_rent.toFixed(2)}` : 'N/A'}</td>
      <td>{item.pricing ? item.pricing.due_day : 'N/A'}</td>
      <td>
        <Group>
          <Button
            leftSection={<IconCurrencyDollar size={14} />}
            onClick={() => handleOpenModal(item, item.pricing)}
            size="xs"
            variant="outline"
          >
            {item.pricing ? 'Edit Pricing' : 'Add Pricing'}
          </Button>
          {item.pricing && (
            <Button
              leftSection={<IconTrash size={14} />}
              onClick={() => handleDelete(item.pricing!.id)}
              color="red"
              size="xs"
              variant="outline"
              disabled={deletePricingMutation.isPending}
            >
              Delete
            </Button>
          )}
        </Group>
      </td>
    </tr>
  ));

  return (
    <Stack>
      <Title order={2}>Precios de Alquileres</Title>
      <Paper shadow="sm" p="md">
        <Table striped highlightOnHover>
          <thead>
            <tr>
              <th>Rental ID</th>
              <th>Pricing Status</th>
              <th>Monthly Rent</th>
              <th>Due Day</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>{rows.length > 0 ? rows : (<tr><td colSpan={5}><Text ta="center">No rentals found.</Text></td></tr>)}</tbody>
        </Table>
      </Paper>

      <Modal
        opened={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingPricing ? 'Edit Pricing' : 'Add Pricing'}
        size="lg"
      >
        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <TextInput
              label="Rental ID"
              disabled
              {...form.getInputProps('rental_id')}
            />
            <NumberInput
              label="Monthly Rent"
              required
              step={0.01}
              min={0}
              {...form.getInputProps('monthly_rent')}
            />
            <NumberInput
              label="Security Deposit"
              step={0.01}
              min={0}
              {...form.getInputProps('security_deposit')}
            />
            <MultiSelect 
              label="Utilities Included"
              data={['Water', 'Electricity', 'Gas', 'Internet', 'Trash', 'Other']}
              searchable
              {...form.getInputProps('utilities_included')}
            />
            <MultiSelect
              label="Tenant Responsible For"
              data={['Water', 'Electricity', 'Gas', 'Internet', 'Trash', 'Other']}
              searchable
              {...form.getInputProps('tenant_responsible_for')}
            />
            <NumberInput
              label="Late Fee"
              step={0.01}
              min={0}
              {...form.getInputProps('late_fee')}
            />
            <NumberInput
              label="Due Day of Month"
              required
              min={1}
              max={31}
              step={1}
              {...form.getInputProps('due_day')}
            />
            <Button type="submit" loading={createPricingMutation.isPending || updatePricingMutation.isPending} mt="md">
              {editingPricing ? 'Update Pricing' : 'Create Pricing'}
            </Button>
          </Stack>
        </form>
      </Modal>
    </Stack>
  );
};

export default RentalPricingPage; 