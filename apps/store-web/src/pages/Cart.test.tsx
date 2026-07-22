import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { DialogProvider } from '@store/ui';
import { CartPage } from './Cart';

vi.mock('../api', () => ({
  api: {
    getCart: vi.fn(async () => ({ items: [], subtotal_cents: 0 })),
  },
}));

vi.mock('../auth/StoreAuthProvider', () => ({
  useStoreAuth: () => ({ expireSession: vi.fn() }),
}));

describe('CartPage', () => {
  it('shows empty cart message when there are no items', async () => {
    render(
      <MemoryRouter>
        <DialogProvider>
          <CartPage />
        </DialogProvider>
      </MemoryRouter>,
    );
    expect(await screen.findByText(/Seu carrinho está vazio/i)).toBeInTheDocument();
  });
});
