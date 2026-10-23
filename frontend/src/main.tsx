import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import 'leaflet/dist/leaflet.css'
import './index.css'
import App from './App'
import { ApiError } from './lib/api'
import { AuthProvider } from './lib/auth'
import { CartProvider } from './lib/cart'
import { ToastProvider } from './components/Toast'
import { DialogProvider } from './components/Dialogs'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Do not repeat a request that failed because of the request itself (4xx). Retry once for server or network errors.
      retry: (failureCount, error) =>
        !(error instanceof ApiError && error.status >= 400 && error.status < 500) && failureCount < 1,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ToastProvider>
          <AuthProvider>
            <CartProvider>
              <DialogProvider>
                <App />
              </DialogProvider>
            </CartProvider>
          </AuthProvider>
        </ToastProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
)
