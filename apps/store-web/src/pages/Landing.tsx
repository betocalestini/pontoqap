import { Link } from 'react-router-dom';

export function LandingPage() {
  return (
    <section className="landing">
      <p className="landing__eyebrow">Compras com faturamento mensal</p>
      <h1 className="landing__title">Store</h1>
      <p className="landing__lead">
        Acesse o catálogo, monte seu carrinho e acompanhe faturas após entrar na sua conta.
      </p>
      <div className="landing__actions">
        <Link to="/login" className="landing__btn landing__btn--primary">
          Entrar
        </Link>
        <Link to="/cadastro" className="landing__btn landing__btn--secondary">
          Criar conta
        </Link>
      </div>
    </section>
  );
}
