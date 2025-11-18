/// <reference types="vite/client" />

// YouGlish Widget types
interface Window {
  YG?: {
    Widget: new (container: HTMLElement, options: any) => {
      load: () => void;
    };
  };
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}