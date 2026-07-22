import type { ReactNode } from 'react';
import type { ExportTable } from './exportReport';
import { ReportExportMenu } from './ReportExportMenu';

type Props = {
  title: string;
  description: string;
  children?: ReactNode;
  exportTable?: () => Promise<ExportTable> | ExportTable;
  exportDisabled?: boolean;
};

export function ReportPageHeader({ title, description, children, exportTable, exportDisabled }: Props) {
  return (
    <header className="report-page-header">
      <div>
        <h1>{title}</h1>
        <p className="form-hint">{description}</p>
      </div>
      <div className="report-page-header__actions">
        {children}
        {exportTable ? <ReportExportMenu buildTable={exportTable} disabled={exportDisabled} /> : null}
      </div>
    </header>
  );
}
