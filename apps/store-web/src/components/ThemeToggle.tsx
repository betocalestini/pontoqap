import { useTheme } from '../hooks/useTheme';

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme();

  const isDark = theme === 'dark';
  const label = isDark ? 'Ativar aparência clara' : 'Ativar aparência escura';

  return (
    <button
      type="button"
      className="theme-toggle"
      onClick={toggleTheme}
      aria-pressed={isDark}
      aria-label={label}
      title={label}
    >
      <span aria-hidden>{isDark ? '☀' : '☽'}</span>
      <span className="visually-hidden">{isDark ? 'Modo claro' : 'Modo escuro'}</span>
    </button>
  );
}
