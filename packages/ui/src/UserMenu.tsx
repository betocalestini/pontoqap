import { useEffect, useId, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

export type UserMenuProps = {
  name: string;
  email: string;
  profilePath?: string;
  onSignOut: () => void | Promise<void>;
};

function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function UserMenu({ name, email, profilePath = '/perfil', onSignOut }: UserMenuProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

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

  return (
    <div className="user-menu" ref={rootRef}>
      <button
        type="button"
        className="user-menu__trigger"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        onClick={() => setOpen((v) => !v)}
        title={name}
      >
        <span className="user-menu__avatar" aria-hidden>
          {initialsFromName(name)}
        </span>
        <span className="user-menu__name">{name}</span>
      </button>
      {open && (
        <div id={menuId} className="user-menu__panel" role="menu">
          <div className="user-menu__header">
            <strong>{name}</strong>
            <span className="user-menu__email">{email}</span>
          </div>
          <Link to={profilePath} className="user-menu__item" role="menuitem" onClick={() => setOpen(false)}>
            Meu perfil
          </Link>
          <button
            type="button"
            className="user-menu__item user-menu__item--action"
            role="menuitem"
            onClick={() => {
              setOpen(false);
              void onSignOut();
            }}
          >
            Sair
          </button>
        </div>
      )}
    </div>
  );
}
