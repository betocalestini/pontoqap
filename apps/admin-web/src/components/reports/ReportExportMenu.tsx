import { useState } from 'react';
import type { ExportFormat, ExportTable } from './exportReport';
import { downloadExport } from './exportReport';

type Props = {
  buildTable: () => Promise<ExportTable> | ExportTable;
  disabled?: boolean;
};

export function ReportExportMenu({ buildTable, disabled }: Props) {
  const [busy, setBusy] = useState(false);

  async function run(format: ExportFormat) {
    setBusy(true);
    try {
      const table = await buildTable();
      await downloadExport(format, table);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="report-export-menu">
      <span className="report-export-menu__label">{busy ? 'Exportando…' : 'Exportar:'}</span>
      <button type="button" disabled={disabled || busy} onClick={() => void run('csv')}>
        CSV
      </button>
      <button type="button" disabled={disabled || busy} onClick={() => void run('xlsx')}>
        Excel
      </button>
      <button type="button" disabled={disabled || busy} onClick={() => void run('pdf')}>
        PDF
      </button>
    </div>
  );
}
