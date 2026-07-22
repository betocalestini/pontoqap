import { Link } from 'react-router-dom';

const services = [
  {
    icon: '⚡',
    title: 'Recarga Rápida',
    description: 'Bebidas, café e lanches rápidos para manter o seu QAP em dia.',
  },
  {
    icon: '🛡️',
    title: 'Material Tático',
    description: 'Coldres, cintos e acessórios resistentes para a linha de frente.',
  },
  {
    icon: '🥾',
    title: 'Uniforme & Manutenção',
    description: 'Itens essenciais e peças de reposição para manter o fardamento alinhado.',
  },
] as const;

export function LandingPage() {
  return (
    <div className="landing-page">
      <div className="landing-page__container">
        <section className="landing-page__hero">
          <h1 className="landing-page__title">Sempre na Escuta</h1>
          <p className="landing-page__lead">
            Sua base de apoio tático e conveniência. Equipamentos padrão e energia rápida para manter você
            pronto para o serviço.
          </p>
          <div className="landing-page__actions">
            <Link to="/login" className="landing-page__btn landing-page__btn--primary">
              Entrar
            </Link>
            <Link to="/cadastro" className="landing-page__btn landing-page__btn--secondary">
              Solicitar acesso
            </Link>
          </div>
        </section>

        <section className="landing-page__services" aria-label="Categorias">
          {services.map(({ icon, title, description }) => (
            <article key={title} className="landing-page__card">
              <div className="landing-page__card-icon" aria-hidden>
                {icon}
              </div>
              <h3>{title}</h3>
              <p>{description}</p>
            </article>
          ))}
        </section>

        <footer className="landing-page__footer">
          <p>
            <strong>PONTO QAP</strong> — Apoio logístico e conveniência.
          </p>
        </footer>
      </div>
    </div>
  );
}
