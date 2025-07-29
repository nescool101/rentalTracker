export type Person = {
  id: string;
  full_name: string;
  phone: string;
  nit: string;
  address?: string;
};

export type Role = {
  id: string;
  role_name: string;
};

export type PersonRole = {
  id: string;
  person_id: string;
  role_id: string;
};

export type Property = {
  id: string;
  address: string;
  apt_number: string;
  city: string;
  state: string;
  department?: string;
  department_id?: string;
  zip_code: string;
  type: string;
  resident_id: string;
  manager_ids: string[];
};

export type BankAccount = {
  id: string;
  person_id: string;
  bank_name: string;
  account_type: string;
  account_number: string;
  account_holder: string;
};

export type Rental = {
  id: string;
  property_id: string;
  renter_id: string;
  bank_account_id: string;
  start_date: string;
  end_date: string;
  payment_terms: string;
  unpaid_months: number;
};

export type Pricing = {
  id: string;
  rental_id: string;
  monthly_rent: number;
  security_deposit: number;
  utilities_included: string[];
  tenant_responsible_for: string[];
  late_fee: number;
  due_day: number;
};

export type PaymentSchedule = {
  id: string;
  rental_id: string;
  due_date: string;
  expected_amount: number;
  is_paid: boolean;
  paid_date: string;
  reminder_sent: boolean;
  recurrence: string;
};

export type RentPayment = {
  id: string;
  rental_id: string;
  payment_date: string;
  amount_paid: number;
  paid_on_time: boolean;
};

export type Document = {
  id: string;
  rental_id: string;
  file_url: string;
  document_type: string;
  upload_date: string;
  uploader_id: string;
  description: string;
};

export type RentalHistory = {
  id: string;
  person_id: string;
  rental_id: string;
  status: string;
  end_reason: string;
  end_date: string;
};

export type MaintenanceRequest = {
  id: string;
  property_id: string;
  renter_id: string;
  description: string;
  request_date: string;
  status: string;
};

export type AuditLog = {
  id: string;
  action: string;
  entity: string;
  entity_id: string;
  changed_by: string;
  timestamp: string;
  details: any;
};

export type User = {
  id: string;
  email: string;
  role: string;
  person_id?: string;
  password_base64?: string;
  status?: string;
}; 