/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API base URL; defaults to the same-origin `/api/v1`. */
  readonly VITE_API_URL?: string;
}
