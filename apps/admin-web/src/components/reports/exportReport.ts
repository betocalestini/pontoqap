export type ExportFormat = 'csv' | 'xlsx' | 'pdf';

export type ExportTable = {
  title: string;
  subtitle?: string;
  headers: string[];
  rows: string[][];
  filenameBase: string;
};

const EXPORT_ROW_CAP = 10_000;

export async function fetchAllPages<T>(
  fetchPage: (offset: number, limit: number) => Promise<{ items: T[]; total: number }>,
  pageSize = 500,
): Promise<T[]> {
  const all: T[] = [];
  let offset = 0;
  let total = 0;
  do {
    const res = await fetchPage(offset, pageSize);
    total = res.total;
    all.push(...(res.items ?? []));
    offset += pageSize;
    if (all.length >= EXPORT_ROW_CAP) {
      return all.slice(0, EXPORT_ROW_CAP);
    }
  } while (offset < total);
  return all;
}

export async function downloadExport(format: ExportFormat, table: ExportTable) {
  switch (format) {
    case 'csv':
      downloadCsv(table);
      break;
    case 'xlsx':
      await downloadXlsx(table);
      break;
    case 'pdf':
      await downloadPdf(table);
      break;
  }
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function escapeCsvCell(v: string) {
  if (/[",\n\r]/.test(v)) {
    return `"${v.replace(/"/g, '""')}"`;
  }
  return v;
}

function downloadCsv(table: ExportTable) {
  const lines = [
    table.headers.map(escapeCsvCell).join(','),
    ...table.rows.map((row) => row.map((c) => escapeCsvCell(String(c ?? ''))).join(',')),
  ];
  const blob = new Blob(['\uFEFF' + lines.join('\n')], { type: 'text/csv;charset=utf-8' });
  downloadBlob(blob, `${table.filenameBase}.csv`);
}

async function downloadXlsx(table: ExportTable) {
  const XLSX = await import('xlsx');
  const ws = XLSX.utils.aoa_to_sheet([table.headers, ...table.rows]);
  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, ws, 'Dados');
  const out = XLSX.write(wb, { bookType: 'xlsx', type: 'array' });
  downloadBlob(
    new Blob([out], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }),
    `${table.filenameBase}.xlsx`,
  );
}

async function downloadPdf(table: ExportTable) {
  const { jsPDF } = await import('jspdf');
  const autoTable = (await import('jspdf-autotable')).default;
  const doc = new jsPDF({ orientation: 'landscape' });
  doc.setFontSize(14);
  doc.text(table.title, 14, 16);
  if (table.subtitle) {
    doc.setFontSize(9);
    doc.text(table.subtitle, 14, 22);
  }
  autoTable(doc, {
    head: [table.headers],
    body: table.rows,
    startY: table.subtitle ? 26 : 20,
    styles: { fontSize: 8 },
  });
  doc.save(`${table.filenameBase}.pdf`);
}
