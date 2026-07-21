export const THEME_STORAGE_KEY = 'store-platform-color-theme';

export type ColorTheme = 'light' | 'dark';

const META_THEME_COLORS: Record<ColorTheme, string> = {
  dark: '#1a1a1a',
  light: '#f4f4f5',
};

export function readThemePreference(): ColorTheme | null {
  if (typeof localStorage === 'undefined') return null;
  const raw = localStorage.getItem(THEME_STORAGE_KEY);
  if (raw === 'light' || raw === 'dark') return raw;
  return null;
}

export function writeThemePreference(theme: ColorTheme): void {
  if (typeof localStorage === 'undefined') return;
  localStorage.setItem(THEME_STORAGE_KEY, theme);
}

/** stored null → usa prefersDark (padrão true se indeterminado, ex. testes). */
export function resolveEffectiveTheme(stored: ColorTheme | null, prefersDark?: boolean): ColorTheme {
  if (stored === 'light' || stored === 'dark') return stored;
  if (prefersDark === undefined) {
    if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
      prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    } else {
      prefersDark = true;
    }
  }
  return prefersDark ? 'dark' : 'light';
}

export function applyDocumentTheme(theme: ColorTheme): void {
  if (typeof document === 'undefined') return;
  document.documentElement.dataset.theme = theme;
  document.documentElement.style.colorScheme = theme;
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute('content', META_THEME_COLORS[theme]);
  }
}

export function getInitialTheme(): ColorTheme {
  return resolveEffectiveTheme(readThemePreference());
}
