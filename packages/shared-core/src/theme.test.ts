import { describe, expect, it } from 'vitest';
import { resolveEffectiveTheme } from './theme';

describe('resolveEffectiveTheme', () => {
  it('respeita preferência salva', () => {
    expect(resolveEffectiveTheme('light', true)).toBe('light');
    expect(resolveEffectiveTheme('dark', false)).toBe('dark');
  });

  it('usa prefersDark quando não há preferência', () => {
    expect(resolveEffectiveTheme(null, true)).toBe('dark');
    expect(resolveEffectiveTheme(null, false)).toBe('light');
  });
});
