export function exportSubtitle(extra?: string) {
  const base = `Gerado em ${new Date().toLocaleString('pt-BR')}`;
  return extra ? `${extra} · ${base}` : base;
}
