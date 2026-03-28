import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css";

// Keep VITE_SUPABASE_ANON_KEY in the production bundle when set at build time (e.g. Cloudflare Pages).
void import.meta.env.VITE_SUPABASE_ANON_KEY;

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
