import { createPortal } from 'react-dom';

type CatalogAddToastProps = {
  productName: string | null;
  animationKey: number;
};

function CheckIcon() {
  return (
    <svg
      className="catalog-add-toast__icon"
      width="28"
      height="28"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

export function CatalogAddToast({ productName, animationKey }: CatalogAddToastProps) {
  if (!productName) return null;

  return createPortal(
    <div className="catalog-add-toast" aria-hidden>
      <div
        key={animationKey}
        className="catalog-add-toast__card"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        <CheckIcon />
        <p className="catalog-add-toast__title">Adicionado ao carrinho</p>
        <p className="catalog-add-toast__product">{productName}</p>
      </div>
    </div>,
    document.body,
  );
}
