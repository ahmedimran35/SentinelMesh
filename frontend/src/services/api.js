const API_BASE = '/api';

// API key from settings or env
let API_KEY = '';

export function setApiKey(key) {
  API_KEY = key;
  try { sessionStorage.setItem('sm_api_key', key); } catch {}
}

export function loadApiKey() {
  try { API_KEY = sessionStorage.getItem('sm_api_key') || ''; } catch {}
  return API_KEY;
}

function headers() {
  const h = { 'Content-Type': 'application/json' };
  if (API_KEY) h['X-API-Key'] = API_KEY;
  return h;
}

export async function startInvestigation(target, type = 'domain') {
  const res = await fetch(`${API_BASE}/investigate`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ target, type }),
  });
  return res.json();
}

export async function listInvestigations() {
  const res = await fetch(`${API_BASE}/investigations`, { headers: headers() });
  return res.json();
}

export async function getInvestigation(id) {
  const res = await fetch(`${API_BASE}/investigations/${id}`, { headers: headers() });
  return res.json();
}

export async function deleteInvestigation(id) {
  const res = await fetch(`${API_BASE}/investigations/${id}`, { method: 'DELETE', headers: headers() });
  return res.json();
}

export async function getFindings(investigationId) {
  const res = await fetch(`${API_BASE}/investigations/${investigationId}/findings`, { headers: headers() });
  return res.json();
}

export async function searchFindings(query, severity, agent) {
  const params = new URLSearchParams();
  if (query) params.set('query', query);
  if (severity) params.set('severity', severity);
  if (agent) params.set('agent', agent);
  const res = await fetch(`${API_BASE}/findings/search?${params}`, { headers: headers() });
  return res.json();
}

export async function getStats() {
  const res = await fetch(`${API_BASE}/stats`, { headers: headers() });
  return res.json();
}

export async function getAlerts(unacknowledged = false) {
  const params = unacknowledged ? '?unacknowledged=true' : '';
  const res = await fetch(`${API_BASE}/alerts${params}`, { headers: headers() });
  return res.json();
}

export async function acknowledgeAlert(id) {
  const res = await fetch(`${API_BASE}/alerts/${id}/ack`, { method: 'PUT', headers: headers() });
  return res.json();
}

export async function getRules(investigationId) {
  const res = await fetch(`${API_BASE}/investigations/${investigationId}/rules`, { headers: headers() });
  return res.json();
}

export async function addMonitor(target, type = 'domain', scanInterval = '24h') {
  const res = await fetch(`${API_BASE}/monitors`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ target, type, scan_interval: scanInterval }),
  });
  return res.json();
}

export async function listMonitors() {
  const res = await fetch(`${API_BASE}/monitors`, { headers: headers() });
  return res.json();
}

export async function removeMonitor(id) {
  const res = await fetch(`${API_BASE}/monitors/${id}`, { method: 'DELETE', headers: headers() });
  return res.json();
}
