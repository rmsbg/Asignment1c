'use strict';
(() => {
 const root = document.querySelector('#verwaltung');
 if (!root) return;
 const people = root.dataset.kind === '/personen';
 const api = '/api' + root.dataset.kind;
 const content = document.querySelector('#content');
 const message = document.querySelector('#message');
 const esc = value => String(value ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
 const types = ['Personalausweis', 'Reisepass', 'Führerschein', 'Sonstiger Ausweis'];
 const label = people ? 'Person' : 'Fahrrad';
 document.querySelector('.skip-link').addEventListener('click', event => {event.preventDefault();document.querySelector('main').focus();});
 let generation = 0;
 let savedNotice = false;
 async function request(url, options) {
  const response = await fetch(url, {cache:'no-store', ...options});
  const result = await response.json();
  if (!response.ok) throw result;
  return result;
 }
 function notice(text) { message.textContent = text; message.className = text ? 'success' : ''; }
 const field = (name, title, value = '', type = 'text') => `<div class="field"><label for="${name}">${title}</label><input id="${name}" name="${name}" type="${type}" value="${esc(value)}" aria-describedby="error-${name}"><small class="field-error" id="error-${name}"></small></div>`;
 function showErrors(error) {
  content.querySelectorAll('.field-error').forEach(e => e.textContent = '');
  content.querySelectorAll('[aria-invalid]').forEach(e => e.removeAttribute('aria-invalid'));
  const errors = error.errors || {form:'Die Verbindung ist fehlgeschlagen. Bitte erneut versuchen.'};
  const summary = document.querySelector('#form-error');
  summary.textContent = Object.values(errors).join(' ');
  for (const [key, text] of Object.entries(errors)) {
   const target = document.getElementById('error-' + key);
   if (target) target.textContent = text;
   document.getElementById(key)?.setAttribute('aria-invalid', 'true');
  }
  summary.focus();
 }
 function bikePicture(p) {
  return `<figure class="bike-picture"><img src="${esc(p.photoURL)}" alt="${esc(p.name)} – ${esc(p.type)}" loading="lazy"><figcaption hidden>Fahrradbild konnte nicht geladen werden.</figcaption></figure>`;
 }
 root.addEventListener('error',event=>{
  if(event.target.matches('.bike-picture img')){event.target.hidden=true;event.target.nextElementSibling.hidden=false;}
 },true);
 function bikeFields(p) {
  return `<div class="field"><label for="type">Fahrradtyp (Pflichtfeld)</label><select id="type" name="type" aria-describedby="error-type"><option value="">Bitte Typ wählen</option>${['Downhill','MTB','Rennrad'].map(t=>`<option${p.type===t?' selected':''}>${t}</option>`).join('')}</select><small class="field-error" id="error-type"></small></div>${p.id?`${bikePicture(p)}<div class="field"><label for="image">Fahrradbild</label><select id="image" name="image" aria-describedby="error-image"><option value="keep">Vorhandenes Bild beibehalten</option><option value="replace">Neues Bild hochladen</option></select><small class="field-error" id="error-image"></small></div>`:'<input type="hidden" name="image" value="replace">'}<div id="bike-photo-fields" class="field"><label for="photo">Fahrradbild (Pflichtfeld)</label><input type="file" id="photo" name="photo" accept="image/jpeg,image/png" aria-describedby="photo-help error-photo"><small id="photo-help">JPEG oder PNG · höchstens 5 MiB und 20 Millionen Pixel</small><small class="field-error" id="error-photo"></small></div>`;
 }
 function form(p = {}) {
  const editing = !!p.id;
  content.innerHTML = `<div class="section-heading"><h2>${label} ${editing ? 'bearbeiten' : 'anlegen'}</h2></div><form novalidate class="record-panel" id="record-form"><p id="form-error" role="alert" tabindex="-1"></p>${editing ? `<p class="record-id">ID: ${esc(p.id)} · unveränderlich</p>` : '<p>Die ID wird automatisch vergeben.</p>'}${field('name',people ? 'Name (Pflichtfeld)' : 'Bezeichnung (Pflichtfeld)',p.name)}${people ? `<div class="field"><label for="proof">Ausweisnachweis (Pflichtfeld)</label><select name="proof" id="proof" aria-describedby="error-proof">${editing ? '<option value="keep">Vorhandenen Nachweis beibehalten</option>' : ''}<option value="manual">Ausweisnummer manuell erfassen</option><option value="photo">Foto hochladen</option></select><small class="field-error" id="error-proof"></small></div><div id="manual-fields"><div class="field"><label for="documentType">Ausweisart (Pflichtfeld)</label><select id="documentType" name="documentType" aria-describedby="error-documentType">${types.map(t=>`<option>${t}</option>`).join('')}</select><small class="field-error" id="error-documentType"></small></div><div id="description-field">${field('documentDescription','Bezeichnung des sonstigen Ausweises (Pflichtfeld)')}</div>${field('documentNumber','Ausweisnummer (Pflichtfeld)')}</div><div id="photo-fields" class="field"><label for="photo">Ausweisfoto (Pflichtfeld)</label><input type="file" id="photo" name="photo" accept="image/jpeg,image/png" aria-describedby="photo-help error-photo"><small id="photo-help">JPEG oder PNG · höchstens 5 MiB und 20 Millionen Pixel</small><small class="field-error" id="error-photo"></small></div>` : bikeFields(p)}<div class="actions"><button type="submit">Speichern</button><a class="button secondary" href="${editing ? '#'+encodeURIComponent(p.id) : '#'}">Abbrechen</a></div></form>`;
  if (people) {
   const update = () => {
    const mode = document.querySelector('#proof').value;
    for (const [id, enabled] of [['manual-fields',mode==='manual'],['photo-fields',mode==='photo']]) {
     const section=document.getElementById(id); section.hidden=!enabled;
     section.querySelectorAll('input,select').forEach(e=>e.disabled=!enabled);
    }
    const other=document.querySelector('#documentType').value==='Sonstiger Ausweis';
    document.querySelector('#description-field').hidden=!other;
    document.querySelector('#documentDescription').disabled=mode!=='manual'||!other;
   };
   document.querySelector('#proof').addEventListener('change',update);
   document.querySelector('#documentType').addEventListener('change',update);
   update();
  }
  if (!people && editing) {
   const updateImage=()=>{
    const replace=document.querySelector('#image').value==='replace';
    document.querySelector('#bike-photo-fields').hidden=!replace;
    document.querySelector('#photo').disabled=!replace;
   };
   document.querySelector('#image').addEventListener('change',updateImage);
   updateImage();
  }
  document.querySelector('form').addEventListener('submit',async event=>{
   event.preventDefault();
   const form=event.currentTarget;
   const formGeneration=generation;
   const button=form.querySelector('button');
   const data=new FormData(form);
   let options={method:editing?'PUT':'POST'};
   options.body=data;
   button.disabled=true;
   try {
    const saved=await request(api+(editing?'/'+encodeURIComponent(p.id):''), options);
    if(formGeneration!==generation)return;
    savedNotice = true;
    location.hash=encodeURIComponent(saved.id);
   } catch(error) {if(formGeneration!==generation)return;showErrors(error); button.disabled=false;}
  });
 }
 async function render() {
  const own=++generation;
  const route=location.hash.slice(1).split('/');
  let id;
  try {id=decodeURIComponent(route[0]||'');} catch {id='ungueltige-id';}
  if(id==='inhalt')id='';
  content.innerHTML='<p>Daten werden geladen …</p>';
  try {
   if (id==='neu') {form();return;}
   if (id) {
    const p=await request(api+'/'+encodeURIComponent(id));if(own!==generation)return;
    if(route[1]==='bearbeiten'){form(p);return;}
    content.innerHTML=`<div class="section-heading"><h2>${label} im Detail</h2><a href="#">Zur Übersicht</a></div><article class="record-panel"><p class="record-id">ID: ${esc(p.id)}</p><h3>${esc(p.name)}</h3>${people ? `<h4>Ausweisnachweis</h4>${p.photoURL ? `<img class="identity-photo" src="${esc(p.photoURL)}" alt="Hinterlegtes Ausweisfoto">` : `<dl><dt>Ausweisart</dt><dd>${esc(p.documentType)}</dd>${p.documentDescription ? `<dt>Bezeichnung</dt><dd>${esc(p.documentDescription)}</dd>` : ''}<dt>Ausweisnummer</dt><dd>${esc(p.documentNumber)}</dd></dl>`}` : `<p class="bike-type">${esc(p.type)}</p>${bikePicture(p)}`}<div class="actions"><a class="button" href="#${encodeURIComponent(p.id)}/bearbeiten">Bearbeiten</a>${people ? `<a class="button secondary" href="/ausleihe?person=${encodeURIComponent(p.id)}">Fahrrad ausleihen</a>` : ''}</div></article>`;
   } else {
    const list=await request(api);if(own!==generation)return;
    content.innerHTML=`<div class="section-heading"><h2>${people?'Registrierte Personen':'Fahrräder im Inventar'}</h2><a class="button" href="#neu">${label} anlegen</a></div><div class="record-panel">${list.length ? `<ul class="record-list">${list.map(p=>`<li><a${!people?' class="inventory-bike"':''} href="#${encodeURIComponent(p.id)}"><span class="record-id">${esc(p.id)}</span><strong>${esc(p.name)}${!people?`<small class="bike-type">${esc(p.type)}</small>`:''}</strong>${!people?bikePicture(p):''}<span aria-hidden="true">→</span></a></li>`).join('')}</ul>` : `<p>Noch keine ${people?'Personen':'Fahrräder'} vorhanden. Lege den ersten Eintrag an.</p>`}</div>`;
   }
  } catch(error) {if(own!==generation)return;content.innerHTML='<p id="form-error" role="alert" tabindex="-1"></p><a href="#">Zur Übersicht</a>';showErrors(error);}
 }
 window.addEventListener('hashchange',()=>{
  notice(savedNotice ? `${label} gespeichert.` : '');savedNotice=false;
  const expected=generation+1;
  render().then(()=>{if(generation!==expected)return;const heading=content.querySelector('h2');if(heading){heading.tabIndex=-1;heading.focus();}});
 });
 render();
})();
