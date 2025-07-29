# rentManager
rentManager by nescao


SUPABASE QUERIES:
Here is the complete set of SQL queries for your Supabase rental management schema, ready to include in your README.md as a setup/backup reference:

✅ Table Creation Queries
create table person (
id uuid primary key default gen_random_uuid(),
full_name text not null,
email text,
phone text,
nit text
);

create table role (
id uuid primary key default gen_random_uuid(),
role_name text not null check (role_name in ('owner', 'renter', 'admin'))
);

create table person_role (
id uuid primary key default gen_random_uuid(),
person_id uuid references person(id),
role_id uuid references role(id)
);

create table property (
id uuid primary key default gen_random_uuid(),
address text not null,
apt_number text,
city text,
state text,
zip_code text,
type text not null,
resident_id uuid references person(id)
);

create table bank_account (
id uuid primary key default gen_random_uuid(),
person_id uuid references person(id),
bank_name text not null,
account_type text not null,
account_number text not null,
account_holder text not null
);

create table rental (
id uuid primary key default gen_random_uuid(),
property_id uuid references property(id),
renter_id uuid references person(id),
bank_account_id uuid references bank_account(id),
start_date date not null,
end_date date not null,
payment_terms text,
unpaid_months integer
);

create table pricing (
id uuid primary key default gen_random_uuid(),
rental_id uuid references rental(id),
monthly_rent numeric not null,
security_deposit numeric,
utilities_included text[],
tenant_responsible_for text[],
late_fee numeric,
due_day integer
);

create table payment_schedule (
id uuid primary key default gen_random_uuid(),
rental_id uuid references rental(id),
due_date date not null,
expected_amount numeric not null,
is_paid boolean default false,
paid_date date,
reminder_sent boolean default false,
recurrence text default 'yearly'
);

-- Add missing property_managers table (used for many-to-many relationship)
create table property_managers (
property_id uuid references property(id) on delete cascade,
manager_person_id uuid references person(id) on delete cascade,
primary key (property_id, manager_person_id)
);

create table rent_payment (
id uuid primary key default gen_random_uuid(),
rental_id uuid references rental(id),
payment_date date not null,
amount_paid numeric not null,
paid_on_time boolean
);

create table document (
id uuid primary key default gen_random_uuid(),
rental_id uuid references rental(id),
file_url text not null,
document_type text,
upload_date timestamp default now(),
uploader_id uuid references person(id),
description text
);

create table rental_history (
id uuid primary key default gen_random_uuid(),
person_id uuid references person(id),
rental_id uuid references rental(id),
status text,
end_reason text,
end_date date
);

create table maintenance_request (
id uuid primary key default gen_random_uuid(),
property_id uuid references property(id),
renter_id uuid references person(id),
description text not null,
request_date date default current_date,
status text default 'open'
);

create table audit_log (
id uuid primary key default gen_random_uuid(),
action text,
entity text,
entity_id uuid,
changed_by uuid references person(id),
timestamp timestamp default now(),
details jsonb
);
✅ Example Insert for Bank Account
insert into bank_account (
id, person_id, bank_name, account_type, account_number, account_holder
) values (
gen_random_uuid(),
'b1c28697-aafb-43b8-8a52-e21e7f3c906d',
'Banco Davivienda',
'Corriente',
'1234567890',
'Nestor Fernando Alvarez Gomez'
);

## Recent Updates

### User Authentication Update

We've implemented the following changes to support user authentication:

1. Updated the database schema:
   - Added `person_id` and `email` fields to the `users` table
   - Removed `username` from the `users` table
   - Moved email field from `person` to `users` table

2. Added a new endpoint: `/api/users/login` for authentication

#### Authentication Example

You can test the authentication using:

```
curl -X POST http://localhost:8081/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"adminscao@rentalmanager.com","password":"Nesc@02025."}'
```

#### Authentication Response

A successful authentication will return:

```json
{
  "success": true,
  "user": {
    "id": "user-uuid",
    "email": "adminscao@rentalmanager.com",
    "role": "admin",
    "person_id": "person-uuid"
  },
  "token": "jwt-token-would-go-here"
}
```

The frontend can then use this information to:
1. Store the authentication token
2. Load additional user information using the person_id if needed
3. Set up user permissions based on the role