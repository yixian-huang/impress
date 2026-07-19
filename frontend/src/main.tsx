import { StrictMode } from 'react'
import './i18n'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { migrateLegacyBrowserStorage } from './lib/browserStorage.ts'

migrateLegacyBrowserStorage();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
