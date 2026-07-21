import { useCallback, useEffect, useState } from 'react';
import {
  applyDocumentTheme,
  getInitialTheme,
  writeThemePreference,
  type ColorTheme,
} from '@store/shared-core';

export function useTheme() {
  const [theme, setThemeState] = useState<ColorTheme>(() => getInitialTheme());

  useEffect(() => {
    applyDocumentTheme(theme);
  }, [theme]);

  const setTheme = useCallback((next: ColorTheme) => {
    writeThemePreference(next);
    setThemeState(next);
    applyDocumentTheme(next);
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme(theme === 'dark' ? 'light' : 'dark');
  }, [theme, setTheme]);

  return { theme, setTheme, toggleTheme, isDark: theme === 'dark' };
}
