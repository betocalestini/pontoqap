import { useCallback, useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

type Product = {
  id: string;
  name: string;
  slug?: string;
  image_url?: string;
  image_alt?: string;
  skus?: { id: string; sale_price_cents: number; available_quantity?: number }[];
};

const PLACEHOLDER_IMAGE = '/product-placeholder.svg';
const SEARCH_DEBOUNCE_MS = 300;

function productImageSrc(product: Product): string {
  const url = product.image_url?.trim();
  if (url) {
    return url;
  }
  return PLACEHOLDER_IMAGE;
}

export function CatalogPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [guestAuth, setGuestAuth] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    const t = window.setTimeout(() => setDebouncedSearch(search.trim()), SEARCH_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [search]);

  const loadProducts = useCallback((term: string) => {
    setLoading(true);
    setError(null);
    api
      .listProducts(term ? { search: term, page_size: 50 } : { page_size: 50 })
      .then((res) => {
        setProducts((res.items ?? []) as Product[]);
        setTotal(res.total ?? 0);
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    loadProducts(debouncedSearch);
  }, [debouncedSearch, loadProducts]);

  async function add(skuId: string) {
    setMsg(null);
    setGuestAuth(false);
    setError(null);
    try {
      await api.addToCart(skuId, 1);
      setMsg('Item adicionado ao carrinho.');
    } catch (e) {
      if (isGuestAuthError(e)) {
        setGuestAuth(true);
      } else {
        setError(e instanceof Error ? e.message : 'Erro');
      }
    }
  }

  const hasSearch = debouncedSearch.length > 0;
  const showEmptyCatalog = !loading && !error && products.length === 0 && !hasSearch;
  const showNoResults = !loading && !error && products.length === 0 && hasSearch;

  return (
    <section className="content-section catalog-page">
      <h1>Catálogo</h1>

      <div className="catalog-search">
        <label className="catalog-search__label" htmlFor="catalog-search-input">
          Buscar produtos
        </label>
        <div className="catalog-search__row">
          <input
            id="catalog-search-input"
            type="search"
            className="catalog-search__input"
            placeholder="Nome do produto…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoComplete="off"
          />
          {search.length > 0 && (
            <button type="button" className="catalog-search__clear" onClick={() => setSearch('')}>
              Limpar
            </button>
          )}
        </div>
        {!loading && !error && total > 0 && (
          <p className="catalog-search__meta">
            {hasSearch
              ? `${total} produto(s) encontrado(s) para “${debouncedSearch}”`
              : `${total} produto(s) no catálogo`}
          </p>
        )}
      </div>

      {loading && <p>Carregando produtos…</p>}
      {error && <p className="error">{error}</p>}
      {guestAuth && <AuthGuestPrompt message={guestAuthMessage('catalog')} />}
      {msg && <p className="ok">{msg}</p>}
      {showEmptyCatalog && (
        <p>Nenhum produto disponível. Cadastre itens no painel admin ou rode as migrations (seed de demo).</p>
      )}
      {showNoResults && <p>Nenhum produto encontrado. Tente outro termo de busca.</p>}

      <ul className="product-grid">
        {products.map((p) => {
          const sku = p.skus?.[0];
          const imageSrc = productImageSrc(p);
          const imageAlt = p.image_alt?.trim() || p.name;
          return (
            <li key={p.id} className="product-card">
              <div className="product-card__media">
                <img
                  className="product-card__image"
                  src={imageSrc}
                  alt={imageAlt}
                  loading="lazy"
                  decoding="async"
                  onError={(e) => {
                    const img = e.currentTarget;
                    if (!img.src.includes('product-placeholder.svg')) {
                      img.src = PLACEHOLDER_IMAGE;
                    }
                  }}
                />
              </div>
              <strong className="product-card__title">{p.name}</strong>
              {sku && (
                <>
                  <p className="product-card__price">{formatMoney(sku.sale_price_cents)}</p>
                  <button type="button" onClick={() => add(sku.id)}>
                    Adicionar ao carrinho
                  </button>
                </>
              )}
            </li>
          );
        })}
      </ul>
    </section>
  );
}
