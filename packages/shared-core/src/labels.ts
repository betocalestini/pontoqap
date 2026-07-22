const TOKEN_PT: Record<string, string> = {
  active: 'ativa',
  adjustment: 'ajuste',
  admin: 'admin',
  approved: 'aprovado',
  audit: 'auditoria',
  billing: 'faturamento',
  blocked: 'bloqueado',
  cancel: 'cancelar',
  cancelled: 'cancelado',
  close: 'fechar',
  confirmed: 'confirmado',
  created: 'criado',
  customer: 'cliente',
  damage: 'avaria',
  disabled: 'desativado',
  entry: 'entrada',
  expired: 'expirado',
  invoice: 'fatura',
  loss: 'perda',
  manual: 'manual',
  open: 'aberto',
  order: 'pedido',
  overdue: 'vencido',
  paid: 'pago',
  pending: 'pendente',
  read: 'consulta',
  return: 'devolução',
  revoked: 'revogado',
  sale: 'venda',
  settled: 'liquidado',
  status: 'status',
  suspended: 'suspenso',
  user: 'usuário',
  write: 'gerenciar',
};

function labelFromMap(map: Record<string, string>, code: string | undefined | null): string {
  if (code == null || code === '') return '—';
  const direct = map[code] ?? map[code.toLowerCase()] ?? map[code.toUpperCase()];
  if (direct) return direct;
  return humanizeCode(code);
}

export function humanizeCode(code: string): string {
  const parts = code.replace(/\./g, ' ').replace(/_/g, ' ').split(/\s+/).filter(Boolean);
  if (!parts.length) return code;
  return parts
    .map((word) => {
      const lower = word.toLowerCase();
      if (TOKEN_PT[lower]) return TOKEN_PT[lower];
      if (word === word.toUpperCase() && word.length > 1) {
        return word
          .toLowerCase()
          .split('_')
          .map((w) => TOKEN_PT[w] ?? w.charAt(0).toUpperCase() + w.slice(1))
          .join(' ');
      }
      return lower.charAt(0).toUpperCase() + lower.slice(1);
    })
    .join(' ');
}

const ORDER_STATUS: Record<string, string> = {
  confirmed: 'Confirmado',
  cancelled: 'Cancelado',
};

const INVOICE_STATUS: Record<string, string> = {
  open: 'Em aberto',
  paid: 'Paga',
  overdue: 'Vencida',
};

const CUSTOMER_STATUS: Record<string, string> = {
  pending: 'Pendente',
  approved: 'Aprovado',
  rejected: 'Rejeitado',
  blocked: 'Bloqueado',
};

const USER_STATUS: Record<string, string> = {
  active: 'Ativo',
  suspended: 'Suspenso',
  disabled: 'Desativado',
};

const STOCK_MOVEMENT: Record<string, string> = {
  entry: 'Entrada',
  initial_stock: 'Estoque inicial',
  sale: 'Venda',
  loss: 'Perda',
  damage: 'Avaria',
  adjustment: 'Ajuste',
  return: 'Devolução',
};

const PIX_RECONCILIATION: Record<string, string> = {
  CONCILIADO: 'Conciliado',
  PENDENTE: 'Pendente',
  VALOR_DIVERGENTE: 'Valor divergente',
  EXPIRADO: 'Expirado',
  EVENTO_DUPLICADO: 'Evento duplicado',
  PAGAMENTO_SEM_FATURA: 'Pagamento sem fatura',
};

const AGING_BUCKET: Record<string, string> = {
  pago: 'Pago',
  a_vencer: 'A vencer',
  '1_a_7': '1–7 dias em atraso',
  '8_a_30': '8–30 dias em atraso',
  mais_30: 'Mais de 30 dias em atraso',
};

const UTILIZATION_BAND: Record<string, string> = {
  ate_50: 'Até 50%',
  '51_a_80': '51–80%',
  '81_a_100': '81–100%',
  acima_100: 'Acima de 100%',
};

const EXCEPTION_EVENT: Record<string, string> = {
  order_cancelled: 'Pedido cancelado',
  invoice_adjustment: 'Ajuste de fatura',
};

const FORECAST_CONFIDENCE: Record<string, string> = {
  low: 'Baixa',
  medium: 'Média',
  high: 'Alta',
};

const FORECAST_METHOD: Record<string, string> = {
  weighted_avg_3m: 'Média ponderada (3 meses)',
};

const PAYMENT_CHARGE_STATUS: Record<string, string> = {
  pending: 'Pendente',
  active: 'Ativa',
  expired: 'Expirada',
  settled: 'Liquidada',
};

const STAFF_ROLE: Record<string, string> = {
  system_admin: 'Administrador do sistema',
  manager: 'Gerente',
  inventory_operator: 'Operador de estoque',
  finance_operator: 'Financeiro',
  customer: 'Cliente',
};

const AUDIT_ACTION: Record<string, string> = {
  'order.cancelled': 'Pedido cancelado',
  'billing.invoice_adjustment': 'Ajuste de fatura',
  'billing.close_manual': 'Fechamento manual de competência',
  'invitation.created': 'Convite criado',
  'invitation.revoked': 'Convite revogado',
  'invitation.accepted': 'Convite aceito',
  'user.role_assigned_from_customer': 'Papel atribuído a partir de cliente',
  'user.role_changed': 'Papel alterado',
  'user.status_changed': 'Status alterado',
  'user.sessions_revoked': 'Sessões revogadas',
};

const AUDIT_ENTITY: Record<string, string> = {
  order: 'Pedido',
  invoice: 'Fatura',
  billing_period: 'Competência',
  admin_invitation: 'Convite admin',
  customer: 'Cliente',
  admin_user: 'Usuário admin',
  stock_movement: 'Movimento de estoque',
};

const PERMISSION: Record<string, string> = {
  'products.read': 'Consultar produtos',
  'products.write': 'Gerenciar produtos',
  'inventory.read': 'Consultar estoque',
  'inventory.adjust': 'Ajustar estoque',
  'inventory.entry': 'Registrar entrada de estoque',
  'inventory.loss': 'Registrar perda e avaria',
  'customers.read': 'Consultar clientes',
  'customers.write': 'Gerenciar clientes',
  'customers.approve': 'Aprovar clientes',
  'customers.change_limit': 'Alterar limite',
  'orders.read': 'Consultar pedidos',
  'orders.cancel': 'Cancelar pedidos',
  'billing.read': 'Consultar faturamento',
  'billing.close': 'Fechar período',
  'payments.read': 'Consultar pagamentos',
  'reports.read': 'Consultar relatórios',
  'reports.dashboard.read': 'Relatório: visão geral',
  'reports.sales.read': 'Relatório: vendas e pedidos',
  'reports.inventory.read': 'Relatório: estoque e movimentações',
  'reports.receivables.read': 'Relatório: contas a receber',
  'reports.payments.read': 'Relatório: conciliação Pix',
  'reports.customers.read': 'Relatório: limites e exposição',
  'reports.exceptions.read': 'Relatório: exceções e ajustes',
  'reports.forecasting.read': 'Relatório: previsão de reposição',
  'settings.write': 'Alterar configurações',
  'audit.read': 'Consultar auditoria',
  'users.manage': 'Gerenciar usuários',
};

export function labelOrderStatus(status: string | undefined | null): string {
  return labelFromMap(ORDER_STATUS, status);
}

export function labelInvoiceStatus(status: string | undefined | null): string {
  return labelFromMap(INVOICE_STATUS, status);
}

export function labelCustomerStatus(status: string | undefined | null): string {
  return labelFromMap(CUSTOMER_STATUS, status);
}

export function labelUserStatus(status: string | undefined | null): string {
  return labelFromMap(USER_STATUS, status);
}

export function labelStockMovementType(type: string | undefined | null): string {
  return labelFromMap(STOCK_MOVEMENT, type);
}

export function labelPixReconciliationStatus(status: string | undefined | null): string {
  return labelFromMap(PIX_RECONCILIATION, status);
}

export function labelAgingBucket(bucket: string | undefined | null): string {
  return labelFromMap(AGING_BUCKET, bucket);
}

export function labelUtilizationBand(band: string | undefined | null): string {
  return labelFromMap(UTILIZATION_BAND, band);
}

export function labelExceptionEventType(type: string | undefined | null): string {
  if (type == null || type === '') return '—';
  if (EXCEPTION_EVENT[type]) return EXCEPTION_EVENT[type];
  const mov = STOCK_MOVEMENT[type] ?? STOCK_MOVEMENT[type.toLowerCase()];
  if (mov) return mov;
  return humanizeCode(type);
}

export function labelForecastConfidence(level: string | undefined | null): string {
  return labelFromMap(FORECAST_CONFIDENCE, level);
}

export function labelForecastMethod(method: string | undefined | null): string {
  return labelFromMap(FORECAST_METHOD, method);
}

export function labelPaymentChargeStatus(status: string | undefined | null): string {
  return labelFromMap(PAYMENT_CHARGE_STATUS, status);
}

export function labelStaffRoleCode(code: string | undefined | null): string {
  return labelFromMap(STAFF_ROLE, code);
}

export function labelAuditAction(action: string | undefined | null): string {
  return labelFromMap(AUDIT_ACTION, action);
}

export function labelAuditEntityType(entityType: string | undefined | null): string {
  return labelFromMap(AUDIT_ENTITY, entityType);
}

export function labelPermission(code: string | undefined | null): string {
  return labelFromMap(PERMISSION, code);
}
