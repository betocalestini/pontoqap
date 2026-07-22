import { Link } from 'react-router-dom';

type AppBrandProps = {
  to?: string;
  className?: string;
};

export function AppBrand({ to = '/', className = 'app-brand' }: AppBrandProps) {
  const content = (
    <>
      Ponto <span className="app-brand__accent">QAP</span>
    </>
  );

  if (to) {
    return (
      <Link to={to} className={className}>
        {content}
      </Link>
    );
  }

  return <span className={className}>{content}</span>;
}
