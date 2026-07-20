import { Link } from 'react-router-dom';

type AuthGuestPromptProps = {
  message: string;
};

export function AuthGuestPrompt({ message }: AuthGuestPromptProps) {
  return (
    <div className="auth-guest-prompt" role="status">
      <p className="error">{message}</p>
      <p className="auth-guest-prompt__links">
        <Link to="/login">Entrar</Link>
        <span className="auth-guest-prompt__sep" aria-hidden>
          {' '}
          ·{' '}
        </span>
        <Link to="/cadastro">Criar conta</Link>
      </p>
    </div>
  );
}
