import { createClient } from '@supabase/supabase-js';

// Get environment variables
const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://wbnoxgtrahnlskrlhkmy.supabase.co';
const supabaseKey = import.meta.env.VITE_SUPABASE_KEY || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Indibm94Z3RyYWhubHNrcmxoa215Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDYzNzE2NzYsImV4cCI6MjA2MTk0NzY3Nn0.Y2sqQFFb6oiEwbyWACZhlNKkhk7ahSo37gW7KL1k0gs';

// Create a single supabase client for interacting with the database
export const supabase = createClient(supabaseUrl, supabaseKey); 