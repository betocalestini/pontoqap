import { describe, expect, it } from 'vitest';
import { formatMoney } from './index';

describe('formatMoney', () => {
  it('formata centavos em BRL', () => {
    expect(formatMoney(16850)).toMatch(/168,50/);
  });

  it('formata zero', () => {
    expect(formatMoney(0)).toMatch(/0,00/);
  });
});
