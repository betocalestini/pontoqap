import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { createPortal } from 'react-dom';

export type ConfirmVariant = 'default' | 'danger';

export type ConfirmOptions = {
  title: string;
  message: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: ConfirmVariant;
};

export type PromptOptions = {
  title: string;
  message?: ReactNode;
  label?: string;
  defaultValue?: string;
  inputType?: 'text' | 'password';
  placeholder?: string;
  confirmLabel?: string;
  cancelLabel?: string;
};

type DialogState =
  | { kind: 'confirm'; options: ConfirmOptions; resolve: (v: boolean) => void }
  | { kind: 'prompt'; options: PromptOptions; resolve: (v: string | null) => void };

type DialogContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
  prompt: (options: PromptOptions) => Promise<string | null>;
};

const DialogContext = createContext<DialogContextValue | null>(null);

function DialogModal({ state, onClose }: { state: DialogState; onClose: () => void }) {
  const titleId = useId();
  const descId = useId();
  const inputRef = useRef<HTMLInputElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const [value, setValue] = useState(
    state.kind === 'prompt' ? (state.options.defaultValue ?? '') : '',
  );

  useEffect(() => {
    const t = window.setTimeout(() => {
      if (state.kind === 'prompt') {
        inputRef.current?.focus();
        inputRef.current?.select();
      } else {
        cancelRef.current?.focus();
      }
    }, 0);
    return () => window.clearTimeout(t);
  }, [state.kind]);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        e.preventDefault();
        if (state.kind === 'confirm') {
          state.resolve(false);
        } else {
          state.resolve(null);
        }
        onClose();
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [state, onClose]);

  function finishConfirm(ok: boolean) {
    if (state.kind === 'confirm') {
      state.resolve(ok);
      onClose();
    }
  }

  function finishPrompt(submit: boolean) {
    if (state.kind !== 'prompt') return;
    state.resolve(submit ? value : null);
    onClose();
  }

  const isConfirm = state.kind === 'confirm';
  const opts = state.options;
  const variant = isConfirm ? (opts as ConfirmOptions).variant ?? 'default' : 'default';

  return (
    <div className="store-dialog-backdrop" onClick={() => (isConfirm ? finishConfirm(false) : finishPrompt(false))}>
      <div
        className="store-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id={titleId} className="store-dialog__title">
          {opts.title}
        </h2>
        {opts.message != null && opts.message !== '' && (
          <div id={descId} className="store-dialog__message">
            {opts.message}
          </div>
        )}
        {!isConfirm && (
          <label className="store-dialog__field">
            {(opts as PromptOptions).label && (
              <span className="store-dialog__label">{(opts as PromptOptions).label}</span>
            )}
            <input
              ref={inputRef}
              type={(opts as PromptOptions).inputType ?? 'text'}
              className="store-dialog__input"
              value={value}
              placeholder={(opts as PromptOptions).placeholder}
              onChange={(e) => setValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  finishPrompt(true);
                }
              }}
            />
          </label>
        )}
        <div className="store-dialog__actions">
          <button
            ref={cancelRef}
            type="button"
            className="store-dialog__btn store-dialog__btn--secondary"
            onClick={() => (isConfirm ? finishConfirm(false) : finishPrompt(false))}
          >
            {opts.cancelLabel ?? 'Cancelar'}
          </button>
          <button
            type="button"
            className={`store-dialog__btn store-dialog__btn--primary${variant === 'danger' ? ' store-dialog__btn--danger' : ''}`}
            onClick={() => (isConfirm ? finishConfirm(true) : finishPrompt(true))}
          >
            {opts.confirmLabel ?? (isConfirm ? 'Confirmar' : 'OK')}
          </button>
        </div>
      </div>
    </div>
  );
}

export function DialogProvider({ children }: { children: ReactNode }) {
  const [queue, setQueue] = useState<DialogState[]>([]);
  const current = queue[0] ?? null;

  const enqueue = useCallback((item: DialogState) => {
    setQueue((q) => [...q, item]);
  }, []);

  const dequeue = useCallback(() => {
    setQueue((q) => q.slice(1));
  }, []);

  const confirm = useCallback(
    (options: ConfirmOptions) =>
      new Promise<boolean>((resolve) => {
        enqueue({ kind: 'confirm', options, resolve });
      }),
    [enqueue],
  );

  const prompt = useCallback(
    (options: PromptOptions) =>
      new Promise<string | null>((resolve) => {
        enqueue({ kind: 'prompt', options, resolve });
      }),
    [enqueue],
  );

  const value: DialogContextValue = { confirm, prompt };

  return (
    <DialogContext.Provider value={value}>
      {children}
      {current != null && createPortal(<DialogModal state={current} onClose={dequeue} />, document.body)}
    </DialogContext.Provider>
  );
}

export function useDialog(): DialogContextValue {
  const ctx = useContext(DialogContext);
  if (!ctx) {
    throw new Error('useDialog deve ser usado dentro de DialogProvider');
  }
  return ctx;
}
