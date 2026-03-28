/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_API_URL?: string;
  /** Public Supabase anon key (safe in bundle); used when wiring @supabase/supabase-js. */
  readonly VITE_SUPABASE_ANON_KEY?: string;
  readonly VITE_MOCK_AUTH?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
