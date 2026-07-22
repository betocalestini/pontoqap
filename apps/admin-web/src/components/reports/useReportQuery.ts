import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

export function useReportQuery(defaults?: Record<string, string>) {
  const [searchParams, setSearchParams] = useSearchParams();

  const values = useMemo(() => {
    const out: Record<string, string> = { ...defaults };
    searchParams.forEach((v, k) => {
      out[k] = v;
    });
    return out;
  }, [searchParams, defaults]);

  const setField = useCallback(
    (key: string, value: string) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (!value) next.delete(key);
        else next.set(key, value);
        next.delete('offset');
        return next;
      });
    },
    [setSearchParams],
  );

  const setOffset = useCallback(
    (offset: number) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        next.set('offset', String(offset));
        return next;
      });
    },
    [setSearchParams],
  );

  const queryParams = useMemo(() => {
    const p: Record<string, string | number | boolean | undefined> = {};
    for (const [k, v] of Object.entries(values)) {
      if (v === '') continue;
      if (k === 'limit' || k === 'offset' || k === 'year' || k === 'month') {
        p[k] = Number(v);
      } else if (v === 'true' || v === 'false') {
        p[k] = v === 'true';
      } else {
        p[k] = v;
      }
    }
    if (p.limit == null) p.limit = 50;
    if (p.offset == null) p.offset = 0;
    return p;
  }, [values]);

  return { values, setField, setOffset, queryParams };
}
