const BASE = import.meta.env.VITE_API_URL || '';

export function getToken(): string | null {
  return localStorage.getItem('admin_token') || localStorage.getItem('token');
}

export function setPanelToken(t: string) {
  localStorage.setItem('token', t);
}

export function setAdminToken(t: string) {
  localStorage.setItem('admin_token', t);
}

export function clearTokens() {
  localStorage.removeItem('token');
  localStorage.removeItem('admin_token');
}

async function api<T = unknown>(path: string, opts?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = { ...(opts?.headers as Record<string, string>) };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  if (!headers['Content-Type'] && opts?.body) headers['Content-Type'] = 'application/json';

  const res = await fetch(BASE + path, { ...opts, headers });
  if (res.status === 403) {
    clearTokens();
    const dest = path.startsWith('/api/admin/') ? '/admin/login' : '/login';
    window.location.href = dest;
    return undefined as T;
  }
  if (!res.ok) {
    const text = await res.text();
    let msg: string;
    try { msg = JSON.parse(text).error || text; } catch { msg = text; }
    throw new Error(msg);
  }
  const ct = res.headers.get('content-type') || '';
  if (ct.includes('application/json')) return res.json();
  return undefined as T;
}

// Panel auth
export function panelLogin(email: string, password: string) {
  return api<{ token: string; tenant?: any }>('/api/panel/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export function panelRegister(email: string, password: string, code: string, name: string) {
  return api<{ token: string }>('/api/panel/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, invitation_code: code, name }),
  });
}

// Panel projects
export function listPanelProjects() {
  return api<any[]>('/api/panel/projects');
}

export function getPanelProject(slug: string) {
  return api<any[]>('/api/panel/projects?slug=' + encodeURIComponent(slug));
}

export function createPanelProject(data: Record<string, any>) {
  return api<any>('/api/panel/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updatePanelProject(id: number, data: Record<string, any>) {
  return api<any>('/api/panel/projects/' + id, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deletePanelProject(id: number) {
  return api<void>('/api/panel/projects/' + id, { method: 'DELETE' });
}

// Panel metrics
export function getPanelProjectMetrics(id: number, days = 30) {
  return api<any>('/api/panel/projects/' + id + '/metrics?days=' + days);
}

export function getPanelProjectSummary(id: number, days = 30) {
  return api<any>('/api/panel/projects/' + id + '/summary?days=' + days);
}

export function getPanelTenantMetrics(days = 7) {
  return api<any>('/api/panel/metrics?days=' + days);
}

// Panel account
export function getPanelAccount() {
  return api<any>('/api/panel/account');
}

export function updatePanelAccount(data: Record<string, any>) {
  return api<any>('/api/panel/account', {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// Admin auth
export function adminLogin(username: string, password: string) {
  return api<{ token: string }>('/api/admin/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

// Admin stats
export function getAdminStats() {
  return api<any>('/api/admin/stats');
}

// Admin tenants
export function listAdminTenants() {
  return api<any[]>('/api/admin/tenants');
}

export function getAdminTenant(id: number) {
  return api<any>('/api/admin/tenants/' + id);
}

// Admin projects
export function listAdminProjects() {
  return api<any[]>('/api/admin/projects');
}

export function createAdminProject(data: Record<string, any>) {
  return api<any>('/api/admin/projects', { method: 'POST', body: JSON.stringify(data) });
}

// Admin invitations
export function listAdminInvitations() {
  return api<any[]>('/api/admin/invitations');
}

export function createAdminInvitation() {
  return api<any>('/api/admin/invitations', { method: 'POST', body: '{}' });
}

// Admin users
export function listAdminUsers() {
  return api<any[]>('/api/admin/users');
}

// Admin metrics
export function getAdminGlobalMetrics(days = 30) {
  return api<any>('/api/admin/metrics?days=' + days);
}
