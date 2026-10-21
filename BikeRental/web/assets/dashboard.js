'use strict';
(() => {
 const root = document.querySelector('#dashboard');
 if (!root) return;
 const status = document.querySelector('#dashboard-status');
 const button = document.querySelector('#dashboard-refresh');
 const fields = [...root.querySelectorAll('[data-metric]')];
 const interval = 5000;
 let timer, active, pending = false, stopped = false;
 let lastSuccess = 'beim Öffnen der Seite';
 button.hidden = false;
 async function refresh() {
  clearTimeout(timer);
  if (document.hidden || stopped) return;
  if (active) {pending = true; return;}
  const controller = new AbortController();
  active = controller;
  root.setAttribute('aria-busy', 'true');
  const timeout = setTimeout(() => controller.abort('timeout'), 5000);
  try {
   const response = await fetch('/api/dashboard', {cache:'no-store',signal:controller.signal});
   if (!response.ok) throw new Error('Dashboard nicht erreichbar');
   const counts = await response.json();
   if (!fields.every(field => Number.isSafeInteger(counts[field.dataset.metric]) && counts[field.dataset.metric] >= 0) || counts.availableSkis + counts.rentedSkis !== counts.totalSkis) throw new Error('Ungültige Kennzahlen');
   if (document.hidden || stopped || controller.signal.aborted) return;
   fields.forEach(field => field.textContent = counts[field.dataset.metric]);
   lastSuccess = new Date().toLocaleTimeString('de-AT');
   status.textContent = `Aktualisiert um ${lastSuccess} · automatisch alle 5 Sekunden.`;
   status.className = '';
  } catch {
   if (!document.hidden && !stopped && controller.signal.reason !== 'hidden' && controller.signal.reason !== 'pagehide') {
    status.textContent = `Aktualisierung fehlgeschlagen. Letzter Stand: ${lastSuccess}. Die Zahlen sind möglicherweise veraltet. Erneuter Versuch automatisch oder über „Jetzt aktualisieren“.`;
    status.className = 'field-error';
   }
  } finally {
   clearTimeout(timeout);
   active = null;
   root.removeAttribute('aria-busy');
   if (!document.hidden && !stopped) timer = setTimeout(refresh, pending ? 0 : interval);
   pending = false;
  }
 }
 button.addEventListener('click', refresh);
 document.addEventListener('visibilitychange', () => {
  if (document.hidden) {clearTimeout(timer);pending = false;active?.abort('hidden');}
  else refresh();
 });
 window.addEventListener('focus', refresh);
 window.addEventListener('pagehide', () => {stopped = true;clearTimeout(timer);active?.abort('pagehide');});
 window.addEventListener('pageshow', () => {stopped = false;refresh();});
 refresh();
})();
