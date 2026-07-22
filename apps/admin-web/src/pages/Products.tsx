import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { AdminCategory, AdminProduct } from '@store/api-client';
import { useDialog } from '@store/ui';
import { api } from '../api';
import { ProductCategoriesPanel } from './ProductCategoriesPanel';

function slugify(name: string) {
  return name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '');
}

function formatBRL(cents: number) {
  return (cents / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

const emptyForm = () => ({
  name: '',
  slug: '',
  description: '',
  category_id: '',
  sku_code: '',
  barcode: '',
  unit: 'UN',
  margin_percent: '30',
  cost_price_cents: '',
  minimum_stock: '0',
  initial_stock: '',
  promo_active: false,
  promo_margin_percent: '',
  promo_quantity: '',
  active: true,
  visible: true,
});

function productImageSrc(url: string | undefined, cacheBust: number) {
  if (!url) return undefined;
  const sep = url.includes('?') ? '&' : '?';
  return `${url}${sep}v=${cacheBust}`;
}

const ALLOWED_IMAGE_EXT = /\.(jpe?g|png|webp|svg)$/i;
const PAGE_SIZE_OPTIONS = [25, 50, 100] as const;
const DEFAULT_PAGE_SIZE = 25;
const SEARCH_DEBOUNCE_MS = 300;
const PROMO_CATEGORY_SLUG = 'promocoes';
const UNCATEGORIZED_CATEGORY_SLUG = 'sem-categoria';

type ProductListStatus = '' | 'store' | 'inactive' | 'hidden';

function productListQueryParams(
  debouncedSearch: string,
  categorySlug: string,
  status: ProductListStatus,
): {
  search?: string;
  category?: string;
  active?: boolean;
  visible?: boolean;
} {
  const params: {
    search?: string;
    category?: string;
    active?: boolean;
    visible?: boolean;
  } = {};
  if (debouncedSearch) params.search = debouncedSearch;
  if (categorySlug) params.category = categorySlug;
  switch (status) {
    case 'store':
      params.active = true;
      params.visible = true;
      break;
    case 'inactive':
      params.active = false;
      break;
    case 'hidden':
      params.visible = false;
      break;
    default:
      break;
  }
  return params;
}

export function ProductsPage() {
  const { confirm } = useDialog();
  const [items, setItems] = useState<AdminProduct[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [imageCacheBust, setImageCacheBust] = useState(() => Date.now());
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreviewUrl, setImagePreviewUrl] = useState<string | null>(null);
  const [editingImageUrl, setEditingImageUrl] = useState<string | null>(null);
  const [imageEditorKey, setImageEditorKey] = useState(0);
  const [saving, setSaving] = useState(false);
  const [loadingEditId, setLoadingEditId] = useState<string | null>(null);
  const [defaultMargin, setDefaultMargin] = useState('30');
  const [bulkMargin, setBulkMargin] = useState('30');
  const [repriceBusy, setRepriceBusy] = useState(false);
  const [categoriesOpen, setCategoriesOpen] = useState(false);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [categorySlug, setCategorySlug] = useState('');
  const [statusFilter, setStatusFilter] = useState<ProductListStatus>('');

  useEffect(() => {
    const t = window.setTimeout(() => {
      setDebouncedSearch(search.trim());
      setPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [search]);

  const clearImageSelection = useCallback(() => {
    setImageFile(null);
    setImagePreviewUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
  }, []);

  useEffect(() => {
    return () => {
      if (imagePreviewUrl) URL.revokeObjectURL(imagePreviewUrl);
    };
  }, [imagePreviewUrl]);

  const load = useCallback(
    async (opts?: { page?: number }) => {
      const p = opts?.page ?? page;
      if (opts?.page != null) {
        setPage(opts.page);
      }
      const [prodRes, catRes] = await Promise.all([
        api.adminListProducts({
          page: p,
          page_size: pageSize,
          ...productListQueryParams(debouncedSearch, categorySlug, statusFilter),
        }),
        api.adminListCategories(),
      ]);
      setItems(prodRes.items ?? []);
      setTotal(prodRes.total ?? 0);
      setCategories(catRes.items ?? []);
    },
    [page, pageSize, debouncedSearch, categorySlug, statusFilter],
  );

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
    api.adminGetPricingSettings()
      .then((s) => {
        const m = String(s.default_margin_percent);
        setDefaultMargin(m);
        setBulkMargin(m);
      })
      .catch(() => {});
  }, [load]);

  async function applyMarginToAll() {
    const margin = parseFloat(bulkMargin.replace(',', '.'));
    if (!Number.isFinite(margin) || margin < 0) {
      setError('Informe uma margem válida');
      return;
    }
    const msg = `Definir margem de ${margin}% em todos os ${total} produtos e recalcular os preços de venda com base no custo médio dos lotes em estoque?`;
    const ok = await confirm({
      title: 'Aplicar margem em massa',
      message: msg,
      confirmLabel: 'Aplicar',
    });
    if (!ok) return;
    setRepriceBusy(true);
    setError(null);
    try {
      await api.adminRepriceAllProducts(margin);
      setDefaultMargin(String(margin));
      setNotice('Margem aplicada e preços recalculados.');
      await load({ page: 1 });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao aplicar margem');
    } finally {
      setRepriceBusy(false);
    }
  }

  function onImageFileChange(file: File | null) {
    setImageFile(file);
    setImagePreviewUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return file ? URL.createObjectURL(file) : null;
    });
  }

  function startCreate() {
    setEditingId('new');
    setForm(emptyForm());
    clearImageSelection();
    setEditingImageUrl(null);
    setError(null);
  }

  async function startEdit(id: string) {
    setError(null);
    setLoadingEditId(id);
    try {
      const p = await api.adminGetProduct(id);
      const sku = p.skus?.[0];
      setEditingId(id);
      setForm({
        name: p.name,
        slug: p.slug,
        description: p.description ?? '',
        category_id: p.category_id ?? '',
        sku_code: sku?.code ?? '',
        barcode: sku?.barcode ?? '',
        unit: sku?.unit ?? 'UN',
        margin_percent: String(p.margin_percent ?? 30),
        cost_price_cents:
          sku?.average_cost_cents != null && sku.average_cost_cents > 0
            ? String(sku.average_cost_cents / 100)
            : sku?.cost_price_cents != null
              ? String(sku.cost_price_cents / 100)
              : '',
        minimum_stock: String(sku?.minimum_stock ?? 0),
        initial_stock: '',
        promo_active: p.promo_active ?? false,
        promo_margin_percent:
          p.promo_margin_percent != null ? String(p.promo_margin_percent) : '',
        promo_quantity:
          p.promo_quantity_total != null && p.promo_quantity_total > 0
            ? String(p.promo_quantity_total)
            : '',
        active: p.active,
        visible: p.visible,
      });
      clearImageSelection();
      setEditingImageUrl(p.image_url ?? null);
      setImageEditorKey((k) => k + 1);
      requestAnimationFrame(() => {
        document.getElementById('product-edit-form')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível carregar o produto');
    } finally {
      setLoadingEditId(null);
    }
  }

  function cancelForm() {
    setEditingId(null);
    setForm(emptyForm());
    clearImageSelection();
    setEditingImageUrl(null);
  }

  async function uploadImageForProduct(productId: string, file: File): Promise<boolean> {
    try {
      const res = await api.adminUploadProductImage(productId, file);
      clearImageSelection();
      setEditingImageUrl(res.image_url);
      setImageEditorKey((k) => k + 1);
      setImageCacheBust(Date.now());
      setNotice('Foto enviada com sucesso.');
      return true;
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Falha no upload';
      setError(`Produto salvo, mas a foto falhou: ${msg}`);
      return false;
    }
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setNotice(null);
    const fileToUpload = imageFile;
    if (fileToUpload && !ALLOWED_IMAGE_EXT.test(fileToUpload.name)) {
      setError('Formato não suportado. Use JPG, PNG, WebP ou SVG.');
      setSaving(false);
      return;
    }
    try {
      const margin = parseFloat(form.margin_percent.replace(',', '.'));
      if (!Number.isFinite(margin) || margin < 0) {
        throw new Error('Margem inválida');
      }
      const costCents = form.cost_price_cents.trim()
        ? Math.round(parseFloat(form.cost_price_cents.replace(',', '.')) * 100)
        : undefined;
      const minStock = parseInt(form.minimum_stock, 10) || 0;

      if (editingId === 'new') {
        const initial = form.initial_stock.trim() ? parseInt(form.initial_stock, 10) : 0;
        if (initial > 0 && costCents == null) {
          throw new Error('Informe o custo unitário para estoque inicial (ou registre entrada depois em Estoque)');
        }
        const created = await api.adminCreateProduct({
          name: form.name,
          slug: form.slug || slugify(form.name),
          description: form.description,
          category_id: form.category_id || undefined,
          sku_code: form.sku_code,
          barcode: form.barcode || undefined,
          sale_price_cents: 0,
          cost_price_cents: costCents,
          minimum_stock: minStock,
          unit: form.unit,
          initial_stock: initial > 0 ? initial : undefined,
        });
        await api.adminUpdateProduct(created.id, { margin_percent: margin });
        const uploadOk = fileToUpload
          ? await uploadImageForProduct(created.id, fileToUpload)
          : true;
        setImageCacheBust(Date.now());
        await load({ page: 1 });
        if (uploadOk) cancelForm();
        return;
      }

      if (editingId) {
        if (form.promo_active) {
          const pm = parseFloat(form.promo_margin_percent.replace(',', '.'));
          const pq = parseInt(form.promo_quantity, 10);
          if (!Number.isFinite(pm) || pm < 0) {
            throw new Error('Margem promocional inválida');
          }
          if (!Number.isFinite(pq) || pq <= 0) {
            throw new Error('Informe a quantidade de unidades da promoção');
          }
        }
        const sku = (await api.adminGetProduct(editingId)).skus?.[0];
        if (sku) {
          await api.adminUpdateSku(sku.id, {
            code: form.sku_code,
            barcode: form.barcode,
            unit: form.unit,
            cost_price_cents: costCents ?? null,
            minimum_stock: minStock,
            active: form.active,
          });
        }
        await api.adminUpdateProduct(editingId, {
          name: form.name,
          description: form.description,
          category_id: form.category_id,
          active: form.active,
          visible: form.visible,
          margin_percent: margin,
          promo_active: form.promo_active,
          promo_margin_percent: form.promo_active
            ? parseFloat(form.promo_margin_percent.replace(',', '.'))
            : undefined,
          promo_quantity: form.promo_active
            ? parseInt(form.promo_quantity, 10) || 0
            : undefined,
        });
        const uploadOk = fileToUpload
          ? await uploadImageForProduct(editingId, fileToUpload)
          : true;
        setImageCacheBust(Date.now());
        await load({ page: 1 });
        if (uploadOk) cancelForm();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar');
    } finally {
      setSaving(false);
    }
  }

  async function removeImage(productId: string, imageId: string) {
    setError(null);
    try {
      await api.adminDeleteProductImage(productId, imageId);
      setImageEditorKey((k) => k + 1);
      if (editingId === productId) {
        const p = await api.adminGetProduct(productId);
        setEditingImageUrl(p.image_url ?? null);
      }
      await load({ page: 1 });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao remover imagem');
    }
  }

  const displayPreviewUrl = imagePreviewUrl ?? productImageSrc(editingImageUrl ?? undefined, imageCacheBust);
  const productListOffset = (page - 1) * pageSize;
  const hasActiveFilters =
    debouncedSearch.length > 0 || categorySlug !== '' || statusFilter !== '';

  return (
    <section className={`content-section products-page${editingId ? ' products-page--editing' : ''}`}>
      <div className="products-page__layout">
        <div className="page-toolbar products-page__toolbar">
          <h1>Produtos</h1>
          {!editingId && (
            <button type="button" onClick={startCreate}>
              Novo produto
            </button>
          )}
        </div>
        {error && <p className="error">{error}</p>}
        {notice && <p className="ok">{notice}</p>}

        {!editingId && (
          <div className="products-pricing-bar">
          <label>
            Margem padrão (%)
            <input
              value={bulkMargin}
              onChange={(e) => setBulkMargin(e.target.value)}
              inputMode="decimal"
            />
            <small>Padrão atual do sistema: {defaultMargin}%</small>
          </label>
          <button type="button" disabled={repriceBusy} onClick={() => void applyMarginToAll()}>
            {repriceBusy ? 'Aplicando…' : 'Aplicar a todos os produtos'}
          </button>
        </div>
        )}

        {!editingId && (
        <ProductCategoriesPanel
          categories={categories}
          open={categoriesOpen}
          onOpenChange={setCategoriesOpen}
          onReload={() => load()}
          onError={setError}
          onCategoryDeleted={(categoryId) => {
            setForm((f) => (f.category_id === categoryId ? { ...f, category_id: '' } : f));
          }}
        />
        )}

      {editingId && (
        <form id="product-edit-form" onSubmit={save} className="form product-form product-form-card">
          <header className="product-form__header">
            <h2>{editingId === 'new' ? 'Novo produto' : 'Editar produto'}</h2>
          </header>

          <div className="product-form__columns">
            <div className="product-form__column">
          <section className="product-form__section" aria-labelledby="product-section-ident">
            <h3 id="product-section-ident" className="product-form__section-title">
              Identificação
            </h3>
            <div className="product-form__grid">
              <label>
                Nome
                <input
                  value={form.name}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, name: e.target.value, slug: f.slug || slugify(e.target.value) }))
                  }
                  required
                />
              </label>
              <label>
                Slug
                <input
                  value={form.slug}
                  onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))}
                  readOnly={editingId !== 'new'}
                  title={editingId !== 'new' ? 'O slug não pode ser alterado após a criação' : undefined}
                />
                {editingId !== 'new' && <small>Identificador fixo (URL da imagem na loja).</small>}
              </label>
              <label className="form__full">
                Descrição
                <textarea
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  rows={2}
                />
              </label>
              <label className="form__full">
                Categoria
                <div className="product-form__category-row">
                  <select
                    value={form.category_id}
                    onChange={(e) => setForm((f) => ({ ...f, category_id: e.target.value }))}
                  >
                    <option value="">—</option>
                    {categories
                      .filter((c) => c.active || c.id === form.category_id)
                      .map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.name}
                          {!c.active ? ' (inativa)' : ''}
                        </option>
                      ))}
                  </select>
                  <button
                    type="button"
                    className="product-form__secondary-btn"
                    onClick={() => setCategoriesOpen(true)}
                  >
                    Categorias
                  </button>
                </div>
              </label>
            </div>
          </section>

          <section className="product-form__section" aria-labelledby="product-section-sku">
            <h3 id="product-section-sku" className="product-form__section-title">
              SKU e unidade
            </h3>
            <div className="product-form__grid product-form__grid--3">
              <label>
                Código SKU
                <input
                  value={form.sku_code}
                  onChange={(e) => setForm((f) => ({ ...f, sku_code: e.target.value }))}
                  required
                />
              </label>
              <label>
                Código de barras
                <input value={form.barcode} onChange={(e) => setForm((f) => ({ ...f, barcode: e.target.value }))} />
              </label>
              <label>
                Unidade
                <input value={form.unit} onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))} />
              </label>
            </div>
          </section>
            </div>

            <div className="product-form__column">
          <section className="product-form__section" aria-labelledby="product-section-price">
            <h3 id="product-section-price" className="product-form__section-title">
              Preço e estoque
            </h3>
            <div className="product-form__grid">
              <label>
                Margem de lucro (%)
                <input
                  value={form.margin_percent}
                  onChange={(e) => setForm((f) => ({ ...f, margin_percent: e.target.value }))}
                  inputMode="decimal"
                  required
                />
                <small>Preço = custo médio dos lotes × (1 + margem/100)</small>
              </label>
              <label>
                Estoque mínimo
                <input
                  type="number"
                  min={0}
                  value={form.minimum_stock}
                  onChange={(e) => setForm((f) => ({ ...f, minimum_stock: e.target.value }))}
                />
              </label>
              <label className="form__full">
                Custo unitário (R$)
                <input
                  value={form.cost_price_cents}
                  onChange={(e) => setForm((f) => ({ ...f, cost_price_cents: e.target.value }))}
                />
                <small>Base do preço; entradas de estoque atualizam o custo médio.</small>
              </label>
              <ProductSalePreview editingId={editingId} form={form} />
              {editingId === 'new' && (
                <label>
                  Estoque inicial (opcional)
                  <input
                    type="number"
                    min={0}
                    value={form.initial_stock}
                    onChange={(e) => setForm((f) => ({ ...f, initial_stock: e.target.value }))}
                  />
                  <small>Gera movimentação de estoque inicial.</small>
                </label>
              )}
            </div>
          </section>

          {editingId !== 'new' && (
            <section className="product-form__section" aria-labelledby="product-section-promo">
              <fieldset className="product-promo-fieldset">
                <legend id="product-section-promo">Promoção</legend>
                <div className="product-form__grid">
                  <label className="form__checkbox form__full">
                    <input
                      type="checkbox"
                      checked={form.promo_active}
                      onChange={(e) => setForm((f) => ({ ...f, promo_active: e.target.checked }))}
                    />
                    Promoção ativa
                  </label>
                  {form.promo_active && (
                    <>
                      <label>
                        Margem promocional (%)
                        <input
                          value={form.promo_margin_percent}
                          onChange={(e) => setForm((f) => ({ ...f, promo_margin_percent: e.target.value }))}
                          inputMode="decimal"
                          required
                        />
                      </label>
                      <label>
                        Unidades na promoção
                        <input
                          type="number"
                          min={1}
                          value={form.promo_quantity}
                          onChange={(e) => setForm((f) => ({ ...f, promo_quantity: e.target.value }))}
                          required
                        />
                        <small>Ao salvar, redefine a cota. Preço promocional até esgotar.</small>
                      </label>
                    </>
                  )}
                </div>
              </fieldset>
            </section>
          )}

          {editingId !== 'new' && (
            <section className="product-form__section" aria-labelledby="product-section-flags">
              <h3 id="product-section-flags" className="product-form__section-title">
                Publicação
              </h3>
              <div className="product-form__flags">
                <label className="form__checkbox">
                  <input
                    type="checkbox"
                    checked={form.active}
                    onChange={(e) => setForm((f) => ({ ...f, active: e.target.checked }))}
                  />
                  Ativo
                </label>
                <label className="form__checkbox">
                  <input
                    type="checkbox"
                    checked={form.visible}
                    onChange={(e) => setForm((f) => ({ ...f, visible: e.target.checked }))}
                  />
                  Visível na loja
                </label>
              </div>
            </section>
          )}

          <section className="product-form__section" aria-labelledby="product-section-image">
            <h3 id="product-section-image" className="product-form__section-title">
              Imagem
            </h3>
            <div className="product-form__image-block">
              <div className="file-upload">
                {displayPreviewUrl && (
                  <img src={displayPreviewUrl} alt="" className="file-upload__preview" width={160} height={120} />
                )}
                <div className="file-upload__controls">
                  <span className="file-upload__label">Foto principal</span>
                  <label className="file-upload__button">
                    Escolher arquivo
                    <input
                      type="file"
                      accept="image/*"
                      className="file-upload__input"
                      onChange={(e) => onImageFileChange(e.target.files?.[0] ?? null)}
                    />
                  </label>
                  {imageFile && <p className="file-upload__name">{imageFile.name}</p>}
                  <p className="file-upload__hint">Enviada ao clicar em Salvar.</p>
                </div>
              </div>
              {editingId !== 'new' && (
                <ProductImagesEditor key={imageEditorKey} productId={editingId} onRemove={removeImage} />
              )}
            </div>
          </section>
            </div>
          </div>

          <div className="product-form__actions form__actions">
            <button type="submit" disabled={saving}>
              {saving ? 'Salvando…' : 'Salvar'}
            </button>
            <button type="button" className="product-form__secondary-btn" onClick={cancelForm}>
              Cancelar
            </button>
          </div>
        </form>
      )}
      </div>

      {!editingId && (
        <div className="customers-list-filters products-list-filters" role="search">
          <label className="customers-list-filters__field customers-list-filters__field--search">
            <span className="customers-list-filters__label">Buscar</span>
            <input
              type="search"
              className="customers-filter-input"
              placeholder="Nome, slug, SKU ou código de barras…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </label>
          <label className="customers-list-filters__field">
            <span className="customers-list-filters__label">Categoria</span>
            <select
              className="customers-filter-input"
              value={categorySlug}
              onChange={(e) => {
                setCategorySlug(e.target.value);
                setPage(1);
              }}
            >
              <option value="">Todas</option>
              <option value={UNCATEGORIZED_CATEGORY_SLUG}>Sem categoria</option>
              <option value={PROMO_CATEGORY_SLUG}>Em promoção</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.slug}>
                  {cat.name}
                </option>
              ))}
            </select>
          </label>
          <label className="customers-list-filters__field">
            <span className="customers-list-filters__label">Situação</span>
            <select
              className="customers-filter-input"
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value as ProductListStatus);
                setPage(1);
              }}
            >
              <option value="">Todas</option>
              <option value="store">Ativo na loja</option>
              <option value="inactive">Inativo</option>
              <option value="hidden">Oculto na loja</option>
            </select>
          </label>
          {hasActiveFilters && (
            <button
              type="button"
              className="product-list-filters__clear"
              onClick={() => {
                setSearch('');
                setCategorySlug('');
                setStatusFilter('');
                setPage(1);
              }}
            >
              Limpar filtros
            </button>
          )}
        </div>
      )}

      {!editingId && <p className="form-hint">{total} produto(s)</p>}

      {!editingId && items.length === 0 && (
        <p className="form-hint">
          {hasActiveFilters ? 'Nenhum produto corresponde aos filtros.' : 'Nenhum produto cadastrado.'}
        </p>
      )}

      <div className="table-scroll products-table-desktop">
        <table>
          <thead>
            <tr>
              <th />
              <th>Nome</th>
              <th>Categoria</th>
              <th>Margem</th>
              <th>Custo médio</th>
              <th>Preço venda</th>
              <th>Promoção</th>
              <th>Estoque</th>
              <th>Status</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {items.map((p) => (
              <ProductTableRow
                key={p.id}
                product={p}
                loadingEditId={loadingEditId}
                imageCacheBust={imageCacheBust}
                onEdit={startEdit}
              />
            ))}
          </tbody>
        </table>
      </div>

      <ul className="products-table-mobile">
        {items.map((p) => (
          <ProductCard
            key={p.id}
            product={p}
            loadingEditId={loadingEditId}
            imageCacheBust={imageCacheBust}
            onEdit={startEdit}
          />
        ))}
      </ul>

      {!editingId && (
        <div className="pagination-row">
          <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Anterior
          </button>
          <span>
            {total === 0 ? 0 : productListOffset + 1}–{Math.min(productListOffset + pageSize, total)} de {total}
          </span>
          <button
            type="button"
            disabled={productListOffset + pageSize >= total}
            onClick={() => setPage((p) => p + 1)}
          >
            Próxima
          </button>
          <label>
            Por página{' '}
            <select
              value={pageSize}
              onChange={(e) => {
                setPageSize(Number(e.target.value));
                setPage(1);
              }}
            >
              {PAGE_SIZE_OPTIONS.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
        </div>
      )}
    </section>
  );
}

function ProductTableRow({
  product: p,
  loadingEditId,
  imageCacheBust,
  onEdit,
}: {
  product: AdminProduct;
  loadingEditId: string | null;
  imageCacheBust: number;
  onEdit: (id: string) => void;
}) {
  const sku = p.skus?.[0];
  const qty = sku?.available_quantity ?? 0;
  const low = sku && qty < sku.minimum_stock;
  return (
    <tr>
      <td>
        {p.image_url ? (
          <img src={productImageSrc(p.image_url, imageCacheBust)} alt="" className="product-thumb" />
        ) : (
          '—'
        )}
      </td>
      <td>{p.name}</td>
      <td>{p.category_name?.trim() ? p.category_name : '—'}</td>
      <td>{p.margin_percent != null ? `${p.margin_percent}%` : '—'}</td>
      <td>{sku?.average_cost_cents != null ? formatBRL(sku.average_cost_cents) : '—'}</td>
      <td>{sku ? formatBRL(sku.sale_price_cents) : '—'}</td>
      <td>
        {p.on_promotion
          ? `Ativa (${p.promo_quantity_remaining ?? 0} rest.)`
          : p.promo_active
            ? 'Inativa'
            : '—'}
      </td>
      <td className={low ? 'error' : undefined}>{qty}</td>
      <td>
        {!p.active && 'Inativo '}
        {!p.visible && 'Oculto'}
        {p.active && p.visible && 'Ativo'}
      </td>
      <td className="table-actions">
        <button type="button" onClick={() => void onEdit(p.id)} disabled={loadingEditId === p.id}>
          {loadingEditId === p.id ? 'Carregando…' : 'Editar'}
        </button>
        {sku && (
          <Link to={`/estoque?product_id=${p.id}`} className="table-actions__link">
            Estoque
          </Link>
        )}
      </td>
    </tr>
  );
}

function ProductCard({
  product: p,
  loadingEditId,
  imageCacheBust,
  onEdit,
}: {
  product: AdminProduct;
  loadingEditId: string | null;
  imageCacheBust: number;
  onEdit: (id: string) => void;
}) {
  const sku = p.skus?.[0];
  const qty = sku?.available_quantity ?? 0;
  const low = sku && qty < sku.minimum_stock;
  return (
    <li className="product-card">
      <div className="product-card__row">
        {p.image_url ? (
          <img src={productImageSrc(p.image_url, imageCacheBust)} alt="" className="product-thumb" />
        ) : (
          <span className="product-thumb product-thumb--empty">—</span>
        )}
        <div>
          <strong>{p.name}</strong>
          <p className="product-card__meta">
            Categoria: {p.category_name?.trim() ? p.category_name : '—'}
          </p>
          <p className="product-card__meta">
            Margem {p.margin_percent ?? '—'}% · Custo médio{' '}
            {sku?.average_cost_cents != null ? formatBRL(sku.average_cost_cents) : '—'} · Venda{' '}
            {sku ? formatBRL(sku.sale_price_cents) : '—'}
          </p>
          <p className="product-card__meta">
            Promoção:{' '}
            {p.on_promotion
              ? `ativa (${p.promo_quantity_remaining ?? 0} rest.)`
              : p.promo_active
                ? 'inativa'
                : '—'}
          </p>
          <p className="product-card__meta">
            Estoque: <span className={low ? 'error' : undefined}>{qty}</span>
          </p>
          <p className="product-card__meta">
            {!p.active && 'Inativo '}
            {!p.visible && 'Oculto'}
            {p.active && p.visible && 'Ativo'}
          </p>
        </div>
      </div>
      <div className="product-card__actions">
        <button type="button" onClick={() => void onEdit(p.id)} disabled={loadingEditId === p.id}>
          {loadingEditId === p.id ? 'Carregando…' : 'Editar'}
        </button>
        {sku && <Link to={`/estoque?product_id=${p.id}`}>Estoque</Link>}
      </div>
    </li>
  );
}

function salePriceFromCostCents(costCents: number, marginPercent: number): number {
  return Math.round(costCents * (1 + marginPercent / 100));
}

function ProductSalePreview({
  editingId,
  form,
}: {
  editingId: string;
  form: ReturnType<typeof emptyForm>;
}) {
  const [preview, setPreview] = useState<string>('—');
  useEffect(() => {
    const m = parseFloat(form.margin_percent.replace(',', '.'));
    const costFromForm = form.cost_price_cents.trim()
      ? Math.round(parseFloat(form.cost_price_cents.replace(',', '.')) * 100)
      : null;
    if (costFromForm != null && costFromForm > 0 && Number.isFinite(m)) {
      setPreview(formatBRL(salePriceFromCostCents(costFromForm, m)));
      return;
    }
    if (editingId === 'new') {
      setPreview('Informe custo e margem, ou registre entrada em Estoque');
      return;
    }
    api.adminGetProduct(editingId).then((p) => {
      const sku = p.skus?.[0];
      const cost = sku?.average_cost_cents;
      if (sku && cost != null && cost > 0 && Number.isFinite(m)) {
        setPreview(formatBRL(salePriceFromCostCents(cost, m)));
      } else if (sku) {
        setPreview(formatBRL(sku.sale_price_cents));
      }
    });
  }, [editingId, form.margin_percent, form.cost_price_cents]);
  return (
    <p className="product-form__price-preview form__full">
      Preço de venda (calculado): <strong>{preview}</strong>
    </p>
  );
}

function ProductImagesEditor({
  productId,
  onRemove,
}: {
  productId: string;
  onRemove: (productId: string, imageId: string) => void;
}) {
  const [images, setImages] = useState<{ id: string; url: string }[]>([]);

  useEffect(() => {
    api.adminGetProduct(productId).then((p) => {
      setImages(p.images ?? []);
    });
  }, [productId]);

  if (!images.length) return null;

  return (
    <div className="product-form__gallery">
      <p className="product-form__gallery-title">Imagens cadastradas</p>
      <ul className="product-form__gallery-list">
        {images.map((im) => (
          <li key={im.id} className="product-form__gallery-item">
            <img src={im.url} alt="" className="product-thumb" />
            <button type="button" className="button--danger" onClick={() => onRemove(productId, im.id)}>
              Remover
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
