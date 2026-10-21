'use strict';
(() => {
 if (!document.querySelector('#ausleihe')) return;
 const content=document.querySelector('#rental-content'),message=document.querySelector('#rental-message');
 const esc=value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
 const money=value=>new Intl.NumberFormat('de-AT').format(BigInt(value)/100n)+','+(BigInt(value)%100n).toString().padStart(2,'0')+' €';
 const date=value=>value?esc(new Date(value).toLocaleString('de-AT')):'';
 const calendar=value=>esc(value.split('-').reverse().join('.'));
 const today=()=>new Intl.DateTimeFormat('en-CA',{timeZone:'Europe/Vienna',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date());
 const initialPerson=new URLSearchParams(location.search).get('person')||'';
 let generation=0;
 document.querySelector('.skip-link').addEventListener('click',event=>{event.preventDefault();document.querySelector('main').focus();});
 async function request(url,options) {
  const response=await fetch(url,{cache:'no-store',...options});const result=await response.json();
  if(!response.ok)throw {errors:result.errors,status:response.status};return result;
 }
 const post=(url,data)=>request(url,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
 function notice(text){message.textContent=text;message.className=text?'success':'';}
 function showErrors(error){
  content.querySelectorAll('.field-error').forEach(el=>el.textContent='');
  content.querySelectorAll('[aria-invalid]').forEach(el=>el.removeAttribute('aria-invalid'));
  const errors=error.errors||{form:'Die Verbindung ist fehlgeschlagen. Bitte erneut versuchen.'};
  const summary=content.querySelector('#rental-error');summary.textContent=Object.values(errors).join(' ');
  for(const [key,text]of Object.entries(errors)){
   const field=document.getElementById('rental-error-'+key);if(field)field.textContent=text;
   document.getElementById(key)?.setAttribute('aria-invalid','true');
  }
  summary.focus();
 }
 const field=(id,label,input)=>`<div class="field"><label for="${id}">${label}</label>${input}<small class="field-error" id="rental-error-${id}"></small></div>`;
 const options=list=>list.map(v=>`<option value="${esc(v.id)}">${esc(v.name)} · ${esc(v.id)}</option>`).join('');
 const stateNames={booked:'Gebucht',expired:'Abgelaufen · nicht ausgegeben',issued:'Ausgegeben',returned:'Zurückgegeben',cancelled:'Storniert'};
 function rentalList(list){
  if(!list.length)return '<p>Noch keine Einträge.</p>';
  return `<ul class="rental-list">${list.map(r=>`<li class="rental-card"><p class="record-id">${esc(r.id)}</p><h3>${esc(r.skiName)}</h3><p>Ski-ID: ${esc(r.skiID)}<br>Person: <a href="/personen#${encodeURIComponent(r.personID)}">${esc(r.personName)}</a> · ${esc(r.personID)}</p><p>${calendar(r.startsOn)} – ${calendar(r.endsOn)} · ${r.durationDays} ${r.durationDays===1?'Tag':'Tage'}<br><strong>Vereinbarter Preis: ${money(r.totalPriceCents)}</strong></p><p>Gebucht: ${date(r.createdAt)}${r.rentedAt?`<br>Ausgegeben: ${date(r.rentedAt)}`:''}${r.returnedAt?`<br>Zurückgegeben: ${date(r.returnedAt)}`:''}${r.cancelledAt?`<br>Storniert: ${date(r.cancelledAt)}`:''}</p><span class="returned-badge ${r.overdue?'overdue':''}">${esc(stateNames[r.status])}${r.overdue?' · Überfällig':''}</span><div class="actions">${r.status==='booked'&&r.startsOn<=today()?`<button data-action="ausgabe" data-id="${esc(r.id)}">Ski ausgeben</button>`:''}${r.status==='booked'||r.status==='expired'?`<button class="secondary" data-action="stornierung" data-id="${esc(r.id)}">Buchung stornieren</button>`:''}${r.status==='issued'?`<button data-action="rueckgabe" data-id="${esc(r.id)}">Rückgabe erfassen</button>`:''}</div></li>`).join('')}</ul>`;
 }
 function headingFocus(){const h=content.querySelector('h2');if(h){h.tabIndex=-1;h.focus();}}
 async function render(saved={personID:initialPerson,startsOn:today(),endsOn:today()}){
  const own=++generation;content.innerHTML='<p>Ausleihen werden geladen …</p>';
  try{
   const [people,rentals,tariff]=await Promise.all([request('/api/personen'),request('/api/ausleihen'),request('/api/tarife')]);
   if(own!==generation)return;
   const overview=location.hash==='#uebersicht';
   content.innerHTML=`<div class="section-heading"><h2>${overview?'Buchungen, Ausgabe und Rückgabe':'Ausleihe buchen'}</h2><div class="actions compact"><button class="secondary" id="refresh-rentals">Aktualisieren</button>${overview?'<a class="button" href="#neu">Neue Ausleihe</a>':''}</div></div><p id="rental-error" class="field-error" role="alert" tabindex="-1"></p>${!overview?`<form id="rental-form" class="record-panel" novalidate><p>Beginn und geplante Rückgabe zählen beide als Verleihtag. Nach der Buchung die tatsächliche Ausgabe separat erfassen.</p>${!people.length?'<p>Noch keine Personen vorhanden. <a href="/personen#neu">Person anlegen</a></p>':''}${!tariff.version?'<p class="field-error">Noch kein Tarif eingerichtet. <a href="/tarife">Tarife verwalten</a></p>':''}${field('personID','Person (Pflichtfeld)',`<select id="personID" aria-describedby="rental-error-personID"><option value="">Bitte Person wählen</option>${options(people)}</select>`)}<div class="date-fields">${field('startsOn','Verleihbeginn (Pflichtfeld)',`<input type="date" id="startsOn" min="${today()}" value="${esc(saved.startsOn)}" aria-describedby="rental-error-startsOn">`)}${field('endsOn','Geplante Rückgabe (Pflichtfeld)',`<input type="date" id="endsOn" min="${today()}" value="${esc(saved.endsOn)}" aria-describedby="rental-error-endsOn">`)}</div><button type="button" class="secondary" id="check-period" ${!tariff.version?'disabled':''}>Verfügbarkeit und Preis prüfen</button><p id="price-preview" role="status">Bitte den Zeitraum prüfen.</p>${field('skiID','Verfügbares Ski-Paar im Zeitraum (Pflichtfeld)','<select id="skiID" disabled aria-describedby="rental-error-skiID"><option value="">Bitte zuerst Zeitraum prüfen</option></select>')}<p id="no-skis" hidden>Keine Paare im Zeitraum verfügbar. Einen anderen Zeitraum wählen, eine Rückgabe erfassen oder <a href="/ski#neu">ein Ski-Paar anlegen</a>.</p><div class="actions"><button type="submit" disabled>Zum angezeigten Preis buchen</button><a class="button secondary" href="#uebersicht">Abbrechen</a></div></form>`:''}${[['Gebuchte Ausleihen',['booked','expired']],['Ausgegebene Ausleihen',['issued']],['Abgeschlossene Ausleihen',['returned']],['Stornierte Buchungen',['cancelled']]].map(([title,statuses],i)=>`<section class="rental-section" aria-labelledby="rental-title-${i}"><h2 id="rental-title-${i}">${title}</h2>${rentalList(rentals.filter(r=>statuses.includes(r.status)))}</section>`).join('')}`;
   const form=content.querySelector('#rental-form');let quote=null,previewGeneration=0;
   const values=()=>form?{personID:document.querySelector('#personID').value,startsOn:document.querySelector('#startsOn').value,endsOn:document.querySelector('#endsOn').value}:saved;
   if(form){
    const person=form.querySelector('#personID'),ski=form.querySelector('#skiID'),submit=form.querySelector('[type="submit"]'),check=form.querySelector('#check-period'),preview=form.querySelector('#price-preview');
    person.value=people.some(p=>p.id===saved.personID)?saved.personID:'';
    if(saved.personID&&!person.value)showErrors({errors:{personID:'Die vorausgewählte Person wurde nicht gefunden. Bitte eine vorhandene Person auswählen.'}});
    function invalidate(){quote=null;previewGeneration++;ski.innerHTML='<option value="">Bitte zuerst Zeitraum prüfen</option>';ski.disabled=true;submit.disabled=true;check.disabled=!tariff.version;form.querySelector('#no-skis').hidden=true;preview.textContent='Zeitraum geändert. Bitte Verfügbarkeit und Preis erneut prüfen.';}
    for(const id of ['startsOn','endsOn'])form.querySelector('#'+id).addEventListener('input',invalidate);
    async function loadPreview(){
     const attempt=++previewGeneration;quote=null;submit.disabled=true;ski.disabled=true;check.disabled=true;preview.textContent='Verfügbarkeit und Preis werden geprüft …';
     const selected=values();
     try{
      const [price,pairs]=await Promise.all([post('/api/ausleihen/preis',{startsOn:selected.startsOn,endsOn:selected.endsOn}),request('/api/ski/verfuegbar?'+new URLSearchParams({startsOn:selected.startsOn,endsOn:selected.endsOn}))]);
      if(own!==generation||attempt!==previewGeneration)return;
      quote=price;form.querySelector('#no-skis').hidden=!!pairs.length;ski.innerHTML='<option value="">Bitte Ski-Paar wählen</option>'+options(pairs);ski.disabled=!pairs.length;
      preview.textContent=`${price.durationDays} ${price.durationDays===1?'Tag':'Tage'} · Gesamtpreis ${money(price.totalPriceCents)}${pairs.length?'':'. Keine Ski-Paare im Zeitraum verfügbar.'}`;
      submit.disabled=!pairs.length||!people.length;
     }catch(error){if(own!==generation||attempt!==previewGeneration)return;preview.textContent='Der Zeitraum konnte nicht geprüft werden.';showErrors(error);}
     finally{if(own===generation&&attempt===previewGeneration)check.disabled=false;}
    }
    check.addEventListener('click',loadPreview);
    form.addEventListener('submit',async event=>{
     event.preventDefault();if(!quote)return;
     const current=values();submit.disabled=true;
     try{
      await post('/api/ausleihen',{...current,skiID:ski.value,tariffVersion:quote.tariffVersion});
      if(own!==generation)return;notice('Buchung gespeichert. Die tatsächliche Ausgabe bitte separat erfassen.');await render(current);message.focus();
     }catch(error){
      if(own!==generation)return;
      if(error.status===409){await loadPreview();if(own!==generation)return;}
      else submit.disabled=!quote;
      showErrors(error);
     }
    });
   }
   content.querySelector('#refresh-rentals').addEventListener('click',()=>{notice('');render(values()).then(headingFocus);});
   content.querySelectorAll('[data-action]').forEach(button=>button.addEventListener('click',async()=>{
    button.disabled=true;
    try{
     await request('/api/ausleihen/'+encodeURIComponent(button.dataset.id)+'/'+button.dataset.action,{method:'PUT'});
     if(own!==generation)return;
     notice({ausgabe:'Ausgabe gespeichert.',rueckgabe:'Rückgabe gespeichert.',stornierung:'Buchung storniert.'}[button.dataset.action]);
     await render(values());message.focus();
    }catch(error){if(own!==generation)return;showErrors(error);button.disabled=false;}
   }));
  }catch(error){
   if(own!==generation)return;
   content.innerHTML='<p id="rental-error" class="field-error" role="alert" tabindex="-1"></p><button id="retry-rentals">Erneut laden</button>';
   showErrors(error);content.querySelector('#retry-rentals').addEventListener('click',()=>render(saved));
  }
 }
 window.addEventListener('hashchange',()=>{notice('');render().then(headingFocus);});
 render();
})();
