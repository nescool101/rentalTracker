import axios from 'axios';
import type { Person, Property, Rental, BankAccount, MaintenanceRequest, RentPayment, RentalHistory, User, Pricing } from '../types';

// Base API URL - automatically proxied through Vite to backend in development
// In production, use the actual backend URL deployed on Fly.io
const API_URL = import.meta.env.PROD 
  ? (import.meta.env.VITE_API_URL || 'https://rentalfullnescao.fly.dev/api')
  : '/api';

// Token key must match the one in authService.ts
const TOKEN_KEY = 'auth_token';

// Create axios instance
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Create a public API client that doesn't require authentication
const publicApiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for authentication
apiClient.interceptors.request.use(
  (config) => {
    // Get token from localStorage
    const token = localStorage.getItem(TOKEN_KEY);
    
    // If token exists, add to headers
    if (token) {
      config.headers = config.headers || {};
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    return config;
  },
  (error) => {
    console.error('Auth interceptor error:', error);
    return Promise.reject(error);
  }
);

// Add request interceptor for debugging
apiClient.interceptors.request.use(
  (config) => {
    console.log('API Request:', {
      method: config.method,
      url: config.url,
      headers: config.headers,
    });
    return config;
  },
  (error) => {
    console.error('API Request Error:', error);
    return Promise.reject(error);
  }
);

// Add response interceptor for debugging
apiClient.interceptors.response.use(
  (response) => {
    console.log('API Response:', {
      status: response.status,
      url: response.config.url,
      data: response.data ? (Array.isArray(response.data) ? `Array with ${response.data.length} items` : response.data) : null,
    });
    return response;
  },
  (error) => {
    console.error('API Response Error:', {
      message: error.message,
      response: error.response ? {
        status: error.response.status,
        data: error.response.data
      } : 'No response',
      config: error.config ? {
        url: error.config.url,
        method: error.config.method
      } : 'No config',
    });
    return Promise.reject(error);
  }
);

// Person API
export const personApi = {
  getAll: async (): Promise<Person[]> => {
    const response = await apiClient.get('/persons');
    return response.data;
  },
  getById: async (id: string): Promise<Person> => {
    const response = await apiClient.get(`/persons/${id}`);
    return response.data;
  },
  getByRole: async (role: string): Promise<Person[]> => {
    const response = await apiClient.get(`/persons/role/${role}`);
    return response.data;
  },
  create: async (person: Omit<Person, 'id'>): Promise<Person> => {
    const response = await apiClient.post('/persons', person);
    return response.data;
  },
  update: async (id: string, person: Partial<Person>): Promise<Person> => {
    const response = await apiClient.put(`/persons/${id}`, person);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/persons/${id}`);
  },
};

// Property API
export const propertyApi = {
  getAll: async (): Promise<Property[]> => {
    const response = await apiClient.get('/properties');
    return response.data;
  },
  
  getForCurrentUser: async (user: User | null): Promise<Property[]> => {
    if (!user) {
      console.error('API: getForCurrentUser (Property) called with no user. Returning empty.');
      return [];
    }
    return propertyApi.getAll(); 
  },
  
  getById: async (id: string): Promise<Property> => {
    const response = await apiClient.get(`/properties/${id}`);
    return response.data;
  },

  getByResidentId: async (residentId: string): Promise<Property[]> => {
    const response = await apiClient.get(`/properties/resident/${residentId}`);
    return response.data;
  },

  getByManagerId: async (managerId: string): Promise<Property[]> => {
    const response = await apiClient.get(`/properties/manager/${managerId}`);
    return response.data;
  },

  create: async (property: Omit<Property, 'id'>): Promise<Property> => {
    const response = await apiClient.post('/properties', property);
    return response.data;
  },
  update: async (id: string, property: Partial<Property>): Promise<Property> => {
    const response = await apiClient.put(`/properties/${id}`, property);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/properties/${id}`);
  },
};

// Maintenance Request API
export const maintenanceRequestApi = {
  getAll: async (): Promise<MaintenanceRequest[]> => {
    const response = await apiClient.get('/maintenance-requests');
    return response.data;
  },
  
  getForCurrentUser: async (user: User | null): Promise<MaintenanceRequest[]> => {
    if (!user) {
      console.error('API: User not available for fetching maintenance requests');
      return [];
    }

    if (user.role === 'admin') {
      return maintenanceRequestApi.getAll();
    }

    const effectiveRole = user.role === 'user' ? 'resident' : user.role;

    if (effectiveRole === 'manager') {
      if (!user.person_id) {
        console.error('API (Manager): person_id not available for manager. Cannot fetch requests.');
        return [];
      }
      try {
        const properties = await propertyApi.getByManagerId(user.person_id);
        if (!properties || properties.length === 0) return [];
        const propertyIds = properties.map(p => p.id).filter(id => !!id);
        if (propertyIds.length === 0) return [];
        return maintenanceRequestApi.getByPropertyIds(propertyIds);
      } catch (error) {
        console.error('API (Manager): Error fetching maintenance requests:', error);
        return [];
      }
    }

    if (effectiveRole === 'resident') {
      if (!user.person_id) {
        console.error('API (Resident/User): person_id not available. Cannot fetch requests.');
        return [];
      }
      try {
        const rentals = await rentalApi.getByRenterId(user.person_id);
        if (!rentals || rentals.length === 0) {
          const residentProperties = await propertyApi.getByResidentId(user.person_id);
          if (!residentProperties || residentProperties.length === 0) return [];
          const propertyIds = residentProperties.map(p => p.id).filter(id => !!id);
          if (propertyIds.length === 0) return [];
          return maintenanceRequestApi.getByPropertyIds(propertyIds);
        }
        const propertyIds = Array.from(new Set(rentals.map(r => r.property_id).filter(id => !!id)));
        if (propertyIds.length === 0) return [];
        return maintenanceRequestApi.getByPropertyIds(propertyIds);
      } catch (error) {
        console.error('API (Resident): Error fetching maintenance requests:', error);
        return [];
      }
    }
    return [];
  },
  
  getById: async (id: string): Promise<MaintenanceRequest> => {
    const response = await apiClient.get(`/maintenance-requests/${id}`);
    return response.data;
  },
  
  getByPropertyId: async (propertyId: string): Promise<MaintenanceRequest[]> => {
    const response = await apiClient.get(`/maintenance-requests/property/${propertyId}`);
    return response.data;
  },

  getByPropertyIds: async (propertyIds: string[]): Promise<MaintenanceRequest[]> => {
    if (propertyIds.length === 0) return [];
    const response = await apiClient.post('/maintenance-requests/property-ids', propertyIds);
    return response.data;
  },
  
  getByRenterId: async (renterId: string): Promise<MaintenanceRequest[]> => {
    const response = await apiClient.get(`/maintenance-requests/renter/${renterId}`);
    return response.data;
  },

  getByStatus: async (status: string): Promise<MaintenanceRequest[]> => {
    const response = await apiClient.get(`/maintenance-requests/status/${status}`);
    return response.data;
  },
  
  create: async (request: Omit<MaintenanceRequest, 'id'>): Promise<MaintenanceRequest> => {
    const response = await apiClient.post('/maintenance-requests', request);
    return response.data;
  },
  
  update: async (id: string, request: Partial<MaintenanceRequest>): Promise<MaintenanceRequest> => {
    const response = await apiClient.put(`/maintenance-requests/${id}`, request);
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/maintenance-requests/${id}`);
  },
};

// Rental API
export const rentalApi = {
  getAll: async (): Promise<Rental[]> => {
    const response = await apiClient.get('/rentals');
    return response.data;
  },
  getById: async (id: string): Promise<Rental> => {
    const response = await apiClient.get(`/rentals/${id}`);
    return response.data;
  },
  getByPropertyId: async (propertyId: string): Promise<Rental[]> => {
    const response = await apiClient.get(`/rentals/property/${propertyId}`);
    return response.data;
  },
  getByRenterId: async (renterId: string): Promise<Rental[]> => {
    const response = await apiClient.get(`/rentals/renter/${renterId}`);
    return response.data;
  },
  getActive: async (): Promise<Rental[]> => {
    const response = await apiClient.get('/rentals/active');
    return response.data;
  },
  create: async (rental: Omit<Rental, 'id'>): Promise<Rental> => {
    const response = await apiClient.post('/rentals', rental);
    return response.data;
  },
  update: async (id: string, rental: Partial<Rental>): Promise<Rental> => {
    const response = await apiClient.put(`/rentals/${id}`, rental);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/rentals/${id}`);
  },
};

// Bank Account API
export const bankAccountApi = {
  getAll: async (): Promise<BankAccount[]> => {
    const response = await apiClient.get('/bank-accounts');
    return response.data;
  },
  getById: async (id: string): Promise<BankAccount> => {
    const response = await apiClient.get(`/bank-accounts/${id}`);
    return response.data;
  },
  getByPersonId: async (personId: string): Promise<BankAccount[]> => {
    const response = await apiClient.get(`/bank-accounts/person/${personId}`);
    return response.data;
  },
  create: async (bankAccount: Omit<BankAccount, 'id'>): Promise<BankAccount> => {
    const response = await apiClient.post('/bank-accounts', bankAccount);
    return response.data;
  },
  update: async (id: string, bankAccount: Partial<BankAccount>): Promise<BankAccount> => {
    const response = await apiClient.put(`/bank-accounts/${id}`, bankAccount);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/bank-accounts/${id}`);
  },
};

// Rent Payment API
export const rentPaymentApi = {
  getAll: async (): Promise<RentPayment[]> => {
    const response = await apiClient.get('/payments');
    return response.data;
  },

  getForCurrentUser: async (user: User | null): Promise<RentPayment[]> => {
    if (!user) {
      console.error('API (Payments): User not available.');
      return [];
    }

    if (user.role === 'admin') {
      return rentPaymentApi.getAll();
    }

    if (!user.person_id) {
      console.error(`API (Payments): User ${user.id} (role ${user.role}) has no person_id. Cannot fetch payments.`);
      return [];
    }

    try {
      let relevantRentalIds: string[] = [];

      if (user.role === 'manager') {
        const managerProperties = await propertyApi.getByManagerId(user.person_id);
        if (managerProperties.length === 0) return [];
        const managerPropertyIds = managerProperties.map(p => p.id);
        
        const rentalsPromises = managerPropertyIds.map(propId => rentalApi.getByPropertyId(propId));
        const rentalsForManagerProperties = (await Promise.all(rentalsPromises)).flat();
        relevantRentalIds = rentalsForManagerProperties.map(r => r.id);

      } else if (user.role === 'resident' || user.role === 'user') {
        const rentalsForResident = await rentalApi.getByRenterId(user.person_id);
        relevantRentalIds = rentalsForResident.map(r => r.id);
      }

      if (relevantRentalIds.length === 0) {
        return [];
      }
      
      const allPayments = await rentPaymentApi.getAll(); 
      return allPayments.filter(p => relevantRentalIds.includes(p.rental_id));

    } catch (error) {
      console.error(`API (Payments): Error fetching payments for user ${user.person_id}:`, error);
      return [];
    }
  },
  
  getById: async (id: string): Promise<RentPayment> => {
    const response = await apiClient.get(`/payments/${id}`);
    return response.data;
  },
  
  getByRentalId: async (rentalId: string): Promise<RentPayment[]> => {
    const response = await apiClient.get(`/payments/rental/${rentalId}`);
    return response.data;
  },

  getByRentalIds: async (rentalIds: string[]): Promise<RentPayment[]> => {
    if (rentalIds.length === 0) return [];
    const response = await apiClient.post('/payments/rental-ids', rentalIds);
    return response.data;
  },
  
  getLatePayments: async (): Promise<RentPayment[]> => {
    const response = await apiClient.get('/payments/late');
    return response.data;
  },
  
  getByDateRange: async (startDate: string, endDate: string): Promise<RentPayment[]> => {
    const response = await apiClient.get(`/payments/date-range?start_date=${startDate}&end_date=${endDate}`);
    return response.data;
  },
  
  create: async (payment: Omit<RentPayment, 'id'>): Promise<RentPayment> => {
    const response = await apiClient.post('/payments', payment);
    return response.data;
  },
  
  update: async (id: string, payment: Partial<RentPayment>): Promise<RentPayment> => {
    const response = await apiClient.put(`/payments/${id}`, payment);
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/payments/${id}`);
  },
};

// Rental History API
export const rentalHistoryApi = {
  getAll: async (filters?: { status?: string; startDate?: string; endDate?: string }): Promise<RentalHistory[]> => {
    const response = await apiClient.get('/rental-history', { params: filters });
    return response.data;
  },
  
  getByPersonId: async (personId: string): Promise<RentalHistory[]> => {
    const response = await apiClient.get(`/rental-history/person/${personId}`);
    return response.data;
  },

  getForCurrentUser: async (user: User | null, adminFilters?: { status?: string; startDate?: string; endDate?: string }): Promise<RentalHistory[]> => {
    if (!user) return [];
    
    if (user.role === 'admin') {
      return rentalHistoryApi.getAll(adminFilters);
    }
    
    if (user.person_id) {
      return rentalHistoryApi.getByPersonId(user.person_id);
    } else {
      console.error(`API (RentalHistory): User ${user.id} (role ${user.role}) has no person_id. Cannot fetch rental history.`);
      return [];
    }
  },
  
  getById: async (id: string): Promise<RentalHistory> => {
    const response = await apiClient.get(`/rental-history/${id}`);
    return response.data;
  },
  
  getByRentalId: async (rentalId: string): Promise<RentalHistory[]> => {
    const response = await apiClient.get(`/rental-history/rental/${rentalId}`);
    return response.data;
  },
  
  getByStatus: async (status: string): Promise<RentalHistory[]> => {
    const response = await apiClient.get(`/rental-history/status/${status}`);
    return response.data;
  },
  
  getByDateRange: async (startDate: string, endDate: string): Promise<RentalHistory[]> => {
    const response = await apiClient.get(`/rental-history/date-range?start_date=${startDate}&end_date=${endDate}`);
    return response.data;
  },
  
  getByRentalIds: async (rentalIds: string[]): Promise<RentalHistory[]> => {
    if (rentalIds.length === 0) {
      return [];
    }
    const response = await apiClient.post('/rental-history/for-rentals', { rental_ids: rentalIds }); 
    return response.data;
  },
  
  create: async (history: Omit<RentalHistory, 'id'>): Promise<RentalHistory> => {
    const response = await apiClient.post('/rental-history', history);
    return response.data;
  },
  
  update: async (id: string, history: Partial<RentalHistory>): Promise<RentalHistory> => {
    const response = await apiClient.put(`/rental-history/${id}`, history);
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/rental-history/${id}`);
  },
};

// User API
export const userApi = {
  getAll: async (): Promise<User[]> => {
    const response = await apiClient.get('/users');
    return response.data;
  },
  getById: async (id: string): Promise<User> => {
    const response = await apiClient.get(`/users/${id}`);
    return response.data;
  },
  create: async (user: Omit<User, 'id'>): Promise<User> => {
    const response = await apiClient.post('/users', user);
    return response.data;
  },
  update: async (id: string, user: Partial<User>): Promise<User> => {
    const response = await apiClient.put(`/users/${id}`, user);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/users/${id}`);
  },
};

// Pricing API (Admin)
const PRICING_API_BASE = '/pricing';

export const pricingApi = {
  create: async (data: Omit<Pricing, 'id'>): Promise<Pricing> => {
    const response = await apiClient.post(`${PRICING_API_BASE}`, data);
    return response.data;
  },
  getAll: async (): Promise<Pricing[]> => {
    const response = await apiClient.get(`${PRICING_API_BASE}`);
    return response.data;
  },
  getById: async (id: string): Promise<Pricing> => {
    const response = await apiClient.get(`${PRICING_API_BASE}/${id}`);
    return response.data;
  },
  getByRentalId: async (rentalId: string): Promise<Pricing | null> => {
    try {
      const response = await apiClient.get(`${PRICING_API_BASE}/rental/${rentalId}`);
      return response.data;
    } catch (error: any) {
      if (error.response && error.response.status === 404) {
        return null; // No pricing found for this rental is a valid state
      }
      throw error; // Re-throw other errors
    }
  },
  update: async (id: string, data: Partial<Pricing>): Promise<Pricing> => {
    const response = await apiClient.put(`${PRICING_API_BASE}/${id}`, data);
    return response.data;
  },
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`${PRICING_API_BASE}/${id}`);
  },
};

// Utility for health check, not directly used by components but good for testing
export const utilityApi = {
  healthCheck: async (): Promise<{ status: string }> => {
    const response = await apiClient.get('/health');
    return response.data;
  },
  // New function to trigger email validation and notification process
  triggerEmailNotifications: async (): Promise<{ status: string }> => {
    // The actual API path is /api/validate_email, but apiClient has /api as baseURL
    // So, we call /validate_email relative to that.
    const response = await apiClient.get('/validate_email'); 
    return response.data; 
  },
  // New function to send a custom email
  sendCustomEmail: async (payload: { recipient_person_id: string; subject: string; body: string }): Promise<{ message: string }> => {
    // Path will be /api/emails/custom because apiClient.baseURL is /api
    // and EmailController registers /emails/custom from the base /api admin group.
    const response = await apiClient.post('/emails/custom', payload);
    return response.data;
  },
  // New function to trigger annual renewal reminders
  triggerAnnualRenewalReminders: async (payload: { optional_message?: string }): Promise<{ message: string }> => {
    const response = await apiClient.post('/emails/annual-renewal-reminders', payload);
    return response.data;
  }
};

// Contract Signing API
export const contractSigningApi = {
  // Request a signature for a contract (requires authentication)
  requestSignature: async (data: { contract_id: string, recipient_id: string, expiration_days?: number }) => {
    const response = await apiClient.post('/contract-signing/request', data);
    return response.data;
  },
  
  // Public endpoints - no authentication required
  // Get status of a signature request
  getSignatureStatus: async (signingId: string) => {
    const response = await publicApiClient.get(`/public/contract-signing/status/${signingId}`);
    return response.data;
  },
  
  // Sign a contract
  signContract: async (signingId: string) => {
    const response = await publicApiClient.post(`/public/contract-signing/sign/${signingId}`);
    return response.data;
  },
  
  // Reject a contract
  rejectContract: async (signingId: string) => {
    const response = await publicApiClient.post(`/public/contract-signing/reject/${signingId}`);
    return response.data;
  }
}; 