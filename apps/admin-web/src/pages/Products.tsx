import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { AdminCategory, AdminProduct } from '@store/api-client';
import { api } from '../api';

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
  sale_price_cents: '',
  cost_price_cents: '',
  minimum_stock: '0',
  initial_stock: '',
  active: true,
  visible: true,
});

function productImageSrc(url: string | undefined, cacheBust: number) {
  if (!url) return undefined;
  const sep = url.includes('?') ? '&' : '?';
  return `${url}${sep}v=${cacheBust}`;
}

const ALLOWED_IMAGE_EXT = /\.(jpe?g|png|webp|svg)$/i;

export function ProductsPage() {
  const [items, setItems] = useState<AdminProduct[]>([]);
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

  const load = useCallback(async () => {
    const [prodRes, catRes] = await Promise.all([api.adminListProducts(), api.adminListCategories()]);
    setItems(prodRes.items ?? []);
    setCategories(catRes.items ?? []);
  }, []);

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, [load]);

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
        sale_price_cents: sku ? String(sku.sale_price_cents / 100) : '',
        cost_price_cents: sku?.cost_price_cents != null ? String(sku.cost_price_cents / 100) : '',
        minimum_stock: String(sku?.minimum_stock ?? 0),
        initial_stock: '',
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
      const saleCents = Math.round(parseFloat(form.sale_price_cents.replace(',', '.')) * 100);
      const costCents = form.cost_price_cents.trim()
        ? Math.round(parseFloat(form.cost_price_cents.replace(',', '.')) * 100)
        : undefined;
      const minStock = parseInt(form.minimum_stock, 10) || 0;

      if (editingId === 'new') {
        const initial = form.initial_stock.trim() ? parseInt(form.initial_stock, 10) : 0;
        const created = await api.adminCreateProduct({
          name: form.name,
          slug: form.slug || slugify(form.name),
          description: form.description,
          category_id: form.category_id || undefined,
          sku_code: form.sku_code,
          barcode: form.barcode || undefined,
          sale_price_cents: saleCents,
          cost_price_cents: costCents,
          minimum_stock: minStock,
          unit: form.unit,
          initial_stock: initial > 0 ? initial : undefined,
        });
        const uploadOk = fileToUpload
          ? await uploadImageForProduct(created.id, fileToUpload)
          : true;
        setImageCacheBust(Date.now());
        await load();
        if (uploadOk) cancelForm();
        return;
      }

      if (editingId) {
        await api.adminUpdateProduct(editingId, {
          name: form.name,
          description: form.description,
          category_id: form.category_id,
          active: form.active,
          visible: form.visible,
        });
        const sku = (await api.adminGetProduct(editingId)).skus?.[0];
        if (sku) {
          await api.adminUpdateSku(sku.id, {
            code: form.sku_code,
            barcode: form.barcode,
            unit: form.unit,
            sale_price_cents: saleCents,
            cost_price_cents: costCents ?? null,
            minimum_stock: minStock,
            active: form.active,
            price_reason: 'Alteração no painel',
          });
        }
        const uploadOk = fileToUpload
          ? await uploadImageForProduct(editingId, fileToUpload)
          : true;
        setImageCacheBust(Date.now());
        await load();
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
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao remover imagem');
    }
  }

  const displayPreviewUrl = imagePreviewUrl ?? productImageSrc(editingImageUrl ?? undefined, imageCacheBust);

  return (
    <section className="content-section products-page">
      <div className="page-toolbar">
        <h1>Produtos</h1>
        {!editingId && (
          <button type="button" onClick={startCreate}>
            Novo produto
          </button>
        )}
      </div>
      {error && <p className="error">{error}</p>}
      {notice && <p className="ok">{notice}</p>}

      {editingId && (
        <form id="product-edit-form" onSubmit={save} className="form form--wide product-form">
          <h2>{editingId === 'new' ? 'Novo produto' : 'Editar produto'}</h2>
          <label className="form__full">
            Nome
            <input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value, slug: f.slug || slugify(e.target.value) }))}
              required
            />
          </label>
          <label className="form__full">
            Slug
            <input
              value={form.slug}
              onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))}
              readOnly={editingId !== 'new'}
              title={editingId !== 'new' ? 'O slug não pode ser alterado após a criação' : undefined}
            />
            {editingId !== 'new' && <small>Identificador fixo (usado na URL da imagem na loja).</small>}
          </label>
          <label className="form__full">
            Descrição
            <textarea
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              rows={3}
            />
          </label>
          <label className="form__full">
            Categoria
            <select value={form.category_id} onChange={(e) => setForm((f) => ({ ...f, category_id: e.target.value }))}>
              <option value="">—</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Código SKU
            <input value={form.sku_code} onChange={(e) => setForm((f) => ({ ...f, sku_code: e.target.value }))} required />
          </label>
          <label>
            Código de barras
            <input value={form.barcode} onChange={(e) => setForm((f) => ({ ...f, barcode: e.target.value }))} />
          </label>
          <label>
            Unidade
            <input value={form.unit} onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))} />
          </label>
          <label>
            Preço de venda (R$)
            <input value={form.sale_price_cents} onChange={(e) => setForm((f) => ({ ...f, sale_price_cents: e.target.value }))} required />
          </label>
          <label>
            Custo (R$)
            <input value={form.cost_price_cents} onChange={(e) => setForm((f) => ({ ...f, cost_price_cents: e.target.value }))} />
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
          {editingId !== 'new' && (
            <label className="form__checkbox">
              <input
                type="checkbox"
                checked={form.active}
                onChange={(e) => setForm((f) => ({ ...f, active: e.target.checked }))}
              />
              Ativo
            </label>
          )}
          {editingId !== 'new' && (
            <label className="form__checkbox">
              <input
                type="checkbox"
                checked={form.visible}
                onChange={(e) => setForm((f) => ({ ...f, visible: e.target.checked }))}
              />
              Visível na loja
            </label>
          )}
          <div className="form__full file-upload">
            <span className="file-upload__label">Foto</span>
            {displayPreviewUrl && (
              <img src={displayPreviewUrl} alt="" className="file-upload__preview" width={120} height={90} />
            )}
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
            <p className="file-upload__hint">A imagem será enviada ao clicar em Salvar.</p>
          </div>
          {editingId !== 'new' && (
            <ProductImagesEditor key={imageEditorKey} productId={editingId} onRemove={removeImage} />
          )}
          <div className="form__actions form__full">
            <button type="submit" disabled={saving}>
              {saving ? 'Salvando…' : 'Salvar'}
            </button>
            <button type="button" onClick={cancelForm}>
              Cancelar
            </button>
          </div>
        </form>
      )}

      <div className="table-scroll products-table-desktop">
        <table>
          <thead>
            <tr>
              <th />
              <th>Nome</th>
              <th>Preço</th>
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
      <td>{sku ? formatBRL(sku.sale_price_cents) : '—'}</td>
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
          <Link to={`/estoque?sku_id=${sku.id}`} className="table-actions__link">
            Histórico
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
            {sku ? formatBRL(sku.sale_price_cents) : '—'} · Estoque: <span className={low ? 'error' : undefined}>{qty}</span>
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
        {sku && <Link to={`/estoque?sku_id=${sku.id}`}>Histórico</Link>}
      </div>
    </li>
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
    <div className="form__full">
      <p>Imagens cadastradas</p>
      <ul className="data-list">
        {images.map((im) => (
          <li key={im.id}>
            <img src={im.url} alt="" className="product-thumb" />
            <button type="button" onClick={() => onRemove(productId, im.id)}>
              Remover
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
