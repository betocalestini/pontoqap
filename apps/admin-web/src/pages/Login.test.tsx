import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { LoginPage } from './Login';

vi.mock('../api', () => ({
  api: { login: vi.fn() },
}));

vi.mock('../auth/AuthProvider', () => ({
  useAuth: () => ({ completeLogin: vi.fn() }),
}));

describe('LoginPage', () => {
  it('renders admin login heading', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { name: /Painel administrativo/i })).toBeInTheDocument();
  });
});
