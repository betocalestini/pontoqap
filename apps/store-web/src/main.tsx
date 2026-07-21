import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { DialogProvider } from '@store/ui';
import '@store/ui/dialog.css';
import App from './App';
import { StoreAuthProvider } from './auth/StoreAuthProvider';
import './index.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <DialogProvider>
        <StoreAuthProvider>
          <App />
        </StoreAuthProvider>
      </DialogProvider>
    </BrowserRouter>
  </StrictMode>,
);
