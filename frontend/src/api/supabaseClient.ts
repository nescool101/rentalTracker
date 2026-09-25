import { createClient } from '@supabase/supabase-js';

// Configured at build time (frontend/.env locally, repo secrets in CI). Never hard-code keys here.
const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
const supabaseKey = import.meta.env.VITE_SUPABASE_KEY;

if (!supabaseUrl || !supabaseKey) {
  throw new Error('Missing VITE_SUPABASE_URL or VITE_SUPABASE_KEY — see frontend/.env.example');
}

// Create a single supabase client for interacting with the database
export const supabase = createClient(supabaseUrl, supabaseKey);
