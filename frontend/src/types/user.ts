export interface User {
    id: string;
    email: string;
    role: string;
    status: string; // Values can be: 'active', 'pending', 'disabled', 'newuser'
    person_id: string;
    token: string;
} 