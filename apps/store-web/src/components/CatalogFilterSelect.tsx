import { useEffect, useId, useRef, useState } from 'react';

export type CatalogFilterOption = {
  value: string;
  label: string;
  disabled?: boolean;
  title?: string;
};

type Props = {
  label: string;
  ariaLabel: string;
  value: string;
  onChange: (value: string) => void;
  options: CatalogFilterOption[];
};

export function CatalogFilterSelect({ label, ariaLabel, value, onChange, options }: Props) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const listId = useId();
  const selected = options.find((o) => o.value === value) ?? options[0];

  useEffect(() => {
    if (!open) return;
    function onDoc(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false);
    }
    document.addEventListener('mousedown', onDoc);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDoc);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  function choose(next: string) {
    onChange(next);
    setOpen(false);
  }

  return (
    <div className="catalog-filters__field catalog-filter-select" ref={rootRef}>
      <span className="catalog-filter-select__label">{label}</span>
      <button
        type="button"
        className="catalog-filter-select__trigger"
        aria-label={ariaLabel}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        onClick={() => setOpen((v) => !v)}
      >
        {selected?.label ?? '—'}
      </button>
      {open && (
        <ul id={listId} className="catalog-filter-select__panel" role="listbox" aria-label={ariaLabel}>
          {options.map((opt) => {
            const isSelected = opt.value === value;
            return (
              <li key={opt.value || '__empty__'} role="presentation">
                <button
                  type="button"
                  role="option"
                  aria-selected={isSelected}
                  className={
                    isSelected
                      ? 'catalog-filter-select__option catalog-filter-select__option--selected'
                      : 'catalog-filter-select__option'
                  }
                  disabled={opt.disabled}
                  title={opt.title}
                  onClick={() => {
                    if (!opt.disabled) choose(opt.value);
                  }}
                >
                  {opt.label}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
