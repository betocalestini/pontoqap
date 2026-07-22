import { useCallback, useEffect, useRef, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { CatalogAddToast } from '../components/CatalogAddToast';
import { CatalogFilterSelect } from '../components/CatalogFilterSelect';

type Product = {
  id: string;
  name: string;
  slug?: string;
  image_url?: string;
  image_alt?: string;
  updated_at?: string;
  on_promotion?: boolean;
  promo_quantity_remaining?: number;
  skus?: { id: string; sale_price_cents: number; available_quantity?: number }[];
};

type CatalogCategory = {
  id: string;
  name: string;
  slug: string;
};

export type CatalogSort = '' | 'name' | 'price_asc' | 'price_desc' | 'purchases';

const PROMO_CATEGORY_SLUG = 'promocoes';

const PLACEHOLDER_IMAGE = '/product-placeholder.svg';
const SEARCH_DEBOUNCE_MS = 300;
const ADD_FEEDBACK_MS = 2500;

function CartAddIcon() {
  return (
    <svg
      className="product-card__add-icon"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <circle cx="9" cy="21" r="1" />
      <circle cx="20" cy="21" r="1" />
      <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
    </svg>
  );
}

function productImageSrc(product: Product): string {
  const url = product.image_url?.trim();
  if (!url) {
    return PLACEHOLDER_IMAGE;
  }
  if (product.updated_at) {
    const sep = url.includes('?') ? '&' : '?';
    return `${url}${sep}v=${encodeURIComponent(product.updated_at)}`;
  }
  return url;
}

function categoryLabel(slug: string, categories: CatalogCategory[]): string {
  if (slug === PROMO_CATEGORY_SLUG) return 'Promoções';
  if (!slug) return 'todas as categorias';
  return categories.find((c) => c.slug === slug)?.name ?? slug;
}

export function CatalogPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [categorySlug, setCategorySlug] = useState('');
  const [sort, setSort] = useState<CatalogSort>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [addedProductName, setAddedProductName] = useState<string | null>(null);
  const [addToastKey, setAddToastKey] = useState(0);
  const [addingSkuId, setAddingSkuId] = useState<string | null>(null);
  const addFeedbackTimerRef = useRef<number | null>(null);

  const showAddedFeedback = useCallback((productName: string) => {
    setAddedProductName(productName);
    setAddToastKey((k) => k + 1);
    if (addFeedbackTimerRef.current != null) {
      window.clearTimeout(addFeedbackTimerRef.current);
    }
    addFeedbackTimerRef.current = window.setTimeout(() => {
      setAddedProductName(null);
      addFeedbackTimerRef.current = null;
    }, ADD_FEEDBACK_MS);
  }, []);

  useEffect(() => {
    return () => {
      if (addFeedbackTimerRef.current != null) {
        window.clearTimeout(addFeedbackTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    const t = window.setTimeout(() => setDebouncedSearch(search.trim()), SEARCH_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [search]);

  useEffect(() => {
    api
      .listCatalogCategories()
      .then((res) => setCategories(res.items ?? []))
      .catch(() => setCategories([]));
  }, []);

  const loadProducts = useCallback(
    (term: string, category: string, sortBy: CatalogSort) => {
      setLoading(true);
      setError(null);
      const params: Parameters<typeof api.listProducts>[0] = { page_size: 50 };
      if (term) params.search = term;
      if (category) params.category = category;
      if (sortBy) params.sort = sortBy;
      api
        .listProducts(params)
        .then((res) => {
          setProducts((res.items ?? []) as Product[]);
          setTotal(res.total ?? 0);
        })
        .catch((e: Error) => setError(e.message))
        .finally(() => setLoading(false));
    },
    [],
  );

  useEffect(() => {
    loadProducts(debouncedSearch, categorySlug, sort);
  }, [debouncedSearch, categorySlug, sort, loadProducts]);

  async function add(skuId: string, productName: string) {
    setError(null);
    setAddingSkuId(skuId);
    try {
      await api.addToCart(skuId, 1);
      showAddedFeedback(productName);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    } finally {
      setAddingSkuId(null);
    }
  }

  const hasSearch = debouncedSearch.length > 0;
  const hasCategory = categorySlug.length > 0;
  const hasFilters = hasSearch || hasCategory;
  const showEmptyCatalog = !loading && !error && products.length === 0 && !hasFilters;
  const showNoResults = !loading && !error && products.length === 0 && hasFilters;

  const categoryOptions = [
    { value: '', label: 'Todas as categorias' },
    { value: PROMO_CATEGORY_SLUG, label: 'Promoções' },
    ...categories.map((c) => ({ value: c.slug, label: c.name })),
  ];

  const sortOptions = [
    { value: '', label: 'Padrão (recomendado)' },
    { value: 'name', label: 'Nome (A–Z)' },
    { value: 'price_asc', label: 'Preço (menor → maior)' },
    { value: 'price_desc', label: 'Preço (maior → menor)' },
    { value: 'purchases', label: 'Mais comprados por você' },
  ];

  const metaText =
    !loading && !error && total > 0
      ? hasSearch
        ? `${total} produto(s) encontrado(s) para “${debouncedSearch}”`
        : hasCategory
          ? `${total} produto(s) em ${categoryLabel(categorySlug, categories)}`
          : `${total} produto(s) no catálogo`
      : null;

  return (
    <section className="content-section catalog-page">
      <h1>Catálogo</h1>

      {metaText && <p className="catalog-page__meta">{metaText}</p>}

      <div className="catalog-toolbar" aria-label="Filtros do catálogo">
        <div className="catalog-search">
          <label className="catalog-search__label" htmlFor="catalog-search-input">
            Buscar produtos
          </label>
          <div className="catalog-search__row">
            <input
              id="catalog-search-input"
              type="search"
              className="catalog-search__input"
              placeholder="Pesquisar produtos…"
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
        </div>

        <div className="catalog-filters">
          <CatalogFilterSelect
            label="Categoria"
            ariaLabel="Filtrar por categoria"
            value={categorySlug}
            onChange={setCategorySlug}
            options={categoryOptions}
          />
          <CatalogFilterSelect
            label="Ordenar por"
            ariaLabel="Ordenar produtos"
            value={sort}
            onChange={(v) => setSort(v as CatalogSort)}
            options={sortOptions}
          />
        </div>
      </div>

      {loading && <p>Carregando produtos…</p>}
      {error && <p className="error">{error}</p>}
      {showEmptyCatalog && (
        <p>Nenhum produto disponível. Cadastre itens no painel admin ou rode as migrations (seed de demo).</p>
      )}
      {showNoResults && <p>Nenhum produto encontrado. Ajuste a busca ou os filtros.</p>}

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
                  {p.on_promotion && (p.promo_quantity_remaining ?? 0) > 0 && (
                    <p className="product-card__promo">
                      Promoção: restam {p.promo_quantity_remaining} unidades neste preço.
                    </p>
                  )}
                  <button
                    type="button"
                    className="product-card__add"
                    aria-label={`Adicionar ${p.name} ao carrinho`}
                    disabled={addingSkuId === sku.id}
                    onClick={() => void add(sku.id, p.name)}
                  >
                    <CartAddIcon />
                    Adicionar
                  </button>
                </>
              )}
            </li>
          );
        })}
      </ul>

      <CatalogAddToast productName={addedProductName} animationKey={addToastKey} />
    </section>
  );
}
