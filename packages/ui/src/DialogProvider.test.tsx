import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DialogProvider, useDialog } from './DialogProvider';
import { useEffect } from 'react';

function ConfirmOnMount() {
  const { confirm } = useDialog();
  useEffect(() => {
    void confirm({ title: 'Título teste', message: 'Mensagem', confirmLabel: 'OK' });
  }, [confirm]);
  return null;
}

describe('DialogProvider', () => {
  it('renders confirm dialog with title', async () => {
    render(
      <DialogProvider>
        <ConfirmOnMount />
      </DialogProvider>,
    );
    expect(await screen.findByText('Título teste')).toBeInTheDocument();
    expect(screen.getByText('Mensagem')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'OK' })).toBeInTheDocument();
  });
});
