'use strict';
(() => {
 if (!document.querySelector('#ausleihe')) return;
 const content=document.querySelector('#rental-content'),message=document.querySelector('#rental-message');
 const esc=value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
 const money=value=>new Intl.NumberFormat('de-AT').format(BigInt(value)/100n)+','+(BigInt(value)%100n).toString().padStart(2,'0')+' €';
 const extraOptions=[['helmet','Helm','helmetCents'],['protection','Schutzausrüstung','protectionCents'],['insurance','Versicherung','insuranceCents']];
 const emptyExtras=()=>({helmet:false,protection:false,insurance:false});
 function priceBreakdown(value,label='Gesamtpreis'){
  const selected=extraOptions.filter(([key])=>value.extras[key]);
  return `<dl class="price-breakdown"><div><dt>Fahrradpreis</dt><dd>${money(value.basePriceCents)}</dd></div>${selected.map(([key,name,price])=>`<div><dt>${name} · einmalig</dt><dd>${money(value.extraPrices[price])}</dd></div>`).join('')}<div><dt>Zusatzleistungen gesamt</dt><dd>${money(value.extrasPriceCents)}</dd></div><div class="price-total"><dt>${label}</dt><dd>${money(value.totalPriceCents)}</dd></div></dl>${selected.length?'':'<p class="no-extras">Keine Zusatzleistungen</p>'}`;
 }
 const date=value=>value?esc(new Date(value).toLocaleString('de-AT',{timeZone:'Europe/Vienna'})):'';
 const calendar=value=>esc(value.split('-').reverse().join('.'));
 const today=()=>new Intl.DateTimeFormat('en-CA',{timeZone:'Europe/Vienna',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date());
 document.querySelector('#ausleihe').addEventListener('error',event=>{if(event.target.matches('.bike-picture img')){event.target.hidden=true;event.target.nextElementSibling.hidden=false;}},true);
 const localMinute=(offset=0)=>{
  const parts=new Intl.DateTimeFormat('en-CA',{timeZone:'Europe/Vienna',year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit',hourCycle:'h23'}).formatToParts(new Date(Date.now()+offset));
  const get=type=>parts.find(p=>p.type===type).value;
  return `${get('year')}-${get('month')}-${get('day')}T${get('hour')}:${get('minute')}`;
 };
 const duration=r=>r.mode==='hour'?`${r.durationHours} ${r.durationHours===1?'Stunde':'Stunden'}`:`${r.durationDays} ${r.durationDays===1?'Tag':'Tage'}`;
 const period=r=>r.mode==='hour'?`${date(r.startsAt)} – ${date(r.endsAt)} (Wien)`:`${calendar(r.startsOn)} – ${calendar(r.endsOn)}`;
 const initialPerson=new URLSearchParams(location.search).get('person')||'';
 let generation=0;
 document.querySelector('.skip-link').addEventListener('click',event=>{event.preventDefault();document.querySelector('main').focus();});
 async function request(url,options) {
  const response=await fetch(url,{cache:'no-store',...options});const result=await response.json();
  if(!response.ok)throw {errors:result.errors,status:response.status};return result;
 }
 const post=(url,data)=>request(url,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
 function notice(text,action){
  message.textContent=text;message.className=text?'success':'';
  if(action){const link=document.createElement('a');link.href=action.href;link.textContent=action.text;link.className='notice-action';message.append(' ',link);}
 }
 function clearErrors(){content.querySelectorAll('#rental-error,[id^="rental-error-"]').forEach(el=>el.textContent='');content.querySelectorAll('[aria-invalid]').forEach(el=>el.removeAttribute('aria-invalid'));}
 function showErrors(error){
  clearErrors();
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
  return `<ul class="rental-list">${list.map(r=>`<li class="rental-card" id="rental-${esc(r.id)}" tabindex="-1"><p class="record-id">${esc(r.id)}</p><h3>${esc(r.bikeName)}</h3><p>Fahrrad-ID: ${esc(r.bikeID)}<br>Person: <a href="/personen#${encodeURIComponent(r.personID)}">${esc(r.personName)}</a> · ${esc(r.personID)}</p><p>${esc(r.bikeType)} · ${r.mode==='hour'?'Stundenmiete':'Tagesmiete'}<br>${period(r)} · ${duration(r)}</p>${priceBreakdown(r,'Vereinbarter Gesamtpreis')}<p>Gebucht: ${date(r.createdAt)}${r.rentedAt?`<br>Ausgegeben: ${date(r.rentedAt)}`:''}${r.returnedAt?`<br>Zurückgegeben: ${date(r.returnedAt)}`:''}${r.cancelledAt?`<br>Storniert: ${date(r.cancelledAt)}`:''}</p><span class="returned-badge ${r.overdue?'overdue':''}">${esc(stateNames[r.status])}${r.overdue?' · Überfällig':''}</span><div class="actions">${r.canIssue?`<button data-action="ausgabe" data-id="${esc(r.id)}">Fahrrad ausgeben</button>`:''}${r.status==='booked'||r.status==='expired'?`<button class="secondary" data-action="stornierung" data-id="${esc(r.id)}">Buchung stornieren</button>`:''}${r.status==='issued'?`<button data-action="rueckgabe" data-id="${esc(r.id)}">Rückgabe erfassen</button>`:''}</div></li>`).join('')}</ul>`;
 }
 function headingFocus(){
  let id='';try{if(location.hash.startsWith('#uebersicht/'))id=decodeURIComponent(location.hash.slice('#uebersicht/'.length));}catch{}
  const card=id?document.getElementById('rental-'+id):null;
  if(card){card.focus();return;}
  const h=content.querySelector('h2');if(h){h.tabIndex=-1;h.focus();}
 }
 async function render(saved={personID:initialPerson,bikeType:"Downhill",mode:"day",extras:emptyExtras(),startsOn:today(),endsOn:today()}){
  const own=++generation;content.innerHTML='<p>Ausleihen werden geladen …</p>';
  try{
   const [people,rentals,tariff]=await Promise.all([request('/api/personen'),request('/api/ausleihen'),request('/api/tarife')]);
   if(own!==generation)return;
   const overview=location.hash.split('/')[0]==='#uebersicht';
   content.innerHTML=`<div class="section-heading"><h2>${overview?'Buchungen, Ausgabe und Rückgabe':'Ausleihe buchen'}</h2><div class="actions compact"><button class="secondary" id="refresh-rentals">Aktualisieren</button>${overview?'<a class="button" href="#neu">Neue Ausleihe</a>':'<a class="button secondary" href="#uebersicht">Buchungen und Rückgaben</a>'}</div></div><p id="rental-error" class="field-error" role="alert" tabindex="-1"></p>${!overview?`<form id="rental-form" class="record-panel" novalidate><p>Zeiten gelten in Europe/Vienna. Tagesmiete zählt beide Randtage; Stundenmiete jede angefangene Stunde. Nach der Buchung die tatsächliche Ausgabe separat erfassen. Der bestätigte Preis bleibt bei abweichender Rückgabe unverändert.</p>${!people.length?'<p>Noch keine Personen vorhanden. <a href="/personen#neu">Person anlegen</a></p>':''}${!tariff.version?'<p class="field-error">Noch kein Tarif eingerichtet. <a href="/tarife">Tarife verwalten</a></p>':''}${field('personID','Person (Pflichtfeld)',`<select id="personID" aria-describedby="rental-error-personID"><option value="">Bitte Person wählen</option>${options(people)}</select>`)}${field('bikeType','Fahrradtyp (Pflichtfeld)',`<select id="bikeType" aria-describedby="rental-error-bikeType">${['Downhill','MTB','Rennrad'].map(kind=>`<option${saved.bikeType===kind?' selected':''}>${kind}</option>`).join('')}</select>`)}${field('mode','Mietart (Pflichtfeld)',`<select id="mode" aria-describedby="rental-error-mode"><option value="day"${saved.mode==='day'?' selected':''}>Tagesmiete</option><option value="hour"${saved.mode==='hour'?' selected':''}>Stundenmiete</option></select>`)}<p id="time-help">Stundenzeiten sind minutengenau. Bei Zeitumstellungen sind doppeldeutige/nicht existierende Uhrzeiten ausgeschlossen.</p><button id="use-now" type="button" class="secondary" ${saved.mode==='hour'?'':'hidden'}>Mietbeginn jetzt</button><div class="date-fields">${field('startsOn','Verleihbeginn (Pflichtfeld)',`<input type="${saved.mode==='hour'?'datetime-local':'date'}" id="startsOn" min="${saved.mode==='hour'?localMinute():today()}" value="${esc(saved.startsOn)}" aria-describedby="time-help rental-error-startsOn">`)}${field('endsOn','Geplante Rückgabe (Pflichtfeld)',`<input type="${saved.mode==='hour'?'datetime-local':'date'}" id="endsOn" min="${saved.mode==='hour'?localMinute():today()}" value="${esc(saved.endsOn)}" aria-describedby="time-help rental-error-endsOn">`)}</div><fieldset id="extras" class="extras-fieldset" aria-describedby="extras-help rental-error-extras"><legend>Zusatzleistungen (optional)</legend><p id="extras-help">Jede Pauschale gilt einmal pro Buchung, unabhängig von Dauer und Fahrradtyp. Ohne Auswahl buchen Sie ohne Extras.</p>${extraOptions.map(([key,name,price])=>`<label class="extra-option"><input type="checkbox" id="extra-${key}"${saved.extras?.[key]?' checked':''}><span><strong>${name}</strong><small id="extra-price-${key}">${tariff.version?money(tariff.extras[price])+' einmalig':'Preis nach Tarifeinrichtung'}</small></span></label>`).join('')}<p class="extras-insurance-note">Die Versicherungsoption erfasst Auswahl und Aufpreis. Im lokalen Prototyp erfolgt keine externe Versicherungsabwicklung.</p><small class="field-error" id="rental-error-extras"></small></fieldset><button type="button" class="secondary" id="check-period" ${!tariff.version?'disabled':''}>Verfügbarkeit und Preis prüfen</button><div id="price-preview" role="status">Bitte Verfügbarkeit und Gesamtpreis prüfen.</div><fieldset id="bikeID" class="bike-choice-group" disabled aria-describedby="rental-error-bikeID"><legend>Verfügbares Fahrrad im Zeitraum (Pflichtfeld)</legend><div id="bike-choices"><p>Bitte zuerst Zeitraum prüfen.</p></div><small class="field-error" id="rental-error-bikeID"></small></fieldset><p id="no-bikes" hidden>Für den gewählten Typ und Zeitraum sind keine Fahrräder verfügbar. Einen anderen Zeitraum wählen, <a href="#uebersicht">Rückgaben ansehen</a> oder <a href="/fahrraeder#neu">ein Fahrrad anlegen</a>.</p><div class="actions"><button type="submit" disabled>Zum angezeigten Preis buchen</button><a class="button secondary" href="#uebersicht">Abbrechen</a></div></form>`:''}${[['Gebuchte Ausleihen',['booked','expired']],['Ausgegebene Ausleihen',['issued']],['Abgeschlossene Ausleihen',['returned']],['Stornierte Buchungen',['cancelled']]].map(([title,statuses],i)=>`<section class="rental-section" aria-labelledby="rental-title-${i}"><h2 id="rental-title-${i}">${title}</h2>${rentalList(rentals.filter(r=>statuses.includes(r.status)))}</section>`).join('')}`;
   const form=content.querySelector('#rental-form');let quote=null,previewGeneration=0;
   const values=()=>form?{personID:document.querySelector('#personID').value,bikeType:document.querySelector('#bikeType').value,mode:document.querySelector('#mode').value,extras:Object.fromEntries(extraOptions.map(([key])=>[key,document.querySelector('#extra-'+key).checked])),startsOn:document.querySelector('#startsOn').value,endsOn:document.querySelector('#endsOn').value}:saved;
   if(form){
    const person=form.querySelector('#personID'),bike=form.querySelector('#bikeID'),choices=form.querySelector('#bike-choices'),submit=form.querySelector('[type="submit"]'),check=form.querySelector('#check-period'),preview=form.querySelector('#price-preview');
    person.value=people.some(p=>p.id===saved.personID)?saved.personID:'';
    if(saved.personID&&!person.value)showErrors({errors:{personID:'Die vorausgewählte Person wurde nicht gefunden. Bitte eine vorhandene Person auswählen.'}});
    function invalidate(){quote=null;previewGeneration++;choices.innerHTML='<p>Bitte zuerst Zeitraum prüfen.</p>';bike.disabled=true;submit.disabled=true;check.disabled=!tariff.version;form.querySelector('#no-bikes').hidden=true;preview.textContent='Auswahl geändert. Bitte Verfügbarkeit und Preis erneut prüfen.';}
    for(const id of ['startsOn','endsOn'])form.querySelector('#'+id).addEventListener('input',invalidate);
    form.querySelector('#bikeType').addEventListener('change',invalidate);
    for(const [key] of extraOptions)form.querySelector('#extra-'+key).addEventListener('change',invalidate);
    form.querySelector('#mode').addEventListener('change',()=>{
     const hour=form.querySelector('#mode').value==='hour';
     form.querySelector('#use-now').hidden=!hour;
     for(const [id,offset] of [['startsOn',0],['endsOn',3600000]]){
      const input=form.querySelector('#'+id);input.type=hour?'datetime-local':'date';input.min=hour?localMinute():today();input.value=hour?localMinute(offset):today();
     }
     invalidate();
    });
    form.querySelector('#use-now').addEventListener('click',()=>{form.querySelector('#startsOn').value=localMinute();form.querySelector('#endsOn').value=localMinute(3600000);invalidate();});
    async function loadPreview(){
     clearErrors();
     const attempt=++previewGeneration;quote=null;submit.disabled=true;bike.disabled=true;choices.innerHTML='<p>Fahrräder werden geladen …</p>';form.querySelector('#no-bikes').hidden=true;check.disabled=true;preview.textContent='Verfügbarkeit und Preis werden geprüft …';
     const selected=values();
     try{
      const [price,pairs]=await Promise.all([post('/api/ausleihen/preis',{bikeType:selected.bikeType,mode:selected.mode,startsOn:selected.startsOn,endsOn:selected.endsOn,extras:selected.extras}),request('/api/fahrraeder/verfuegbar?'+new URLSearchParams({bikeType:selected.bikeType,mode:selected.mode,startsOn:selected.startsOn,endsOn:selected.endsOn}))]);
      if(own!==generation||attempt!==previewGeneration)return;
      quote=price;form.querySelector('#no-bikes').hidden=!!pairs.length;choices.innerHTML=pairs.map(b=>`<label class="bike-choice"><input type="radio" name="bikeID" value="${esc(b.id)}"><span class="bike-choice-content"><span class="bike-picture"><img src="${esc(b.photoURL)}" alt="${esc(b.name)} – ${esc(b.type)}"><small hidden>Fahrradbild konnte nicht geladen werden.</small></span><strong>${esc(b.name)}</strong><span>${esc(b.type)} · ${esc(b.id)}</span></span></label>`).join('');bike.disabled=!pairs.length;
      for(const [key,name,fee] of extraOptions)form.querySelector('#extra-price-'+key).textContent=money(price.extraRates[fee])+' einmalig';
      preview.innerHTML=`<p>${esc(price.bikeType)} · ${duration(price)}${pairs.length?'':'. Keine Fahrräder im Zeitraum verfügbar.'}</p>${priceBreakdown(price)}`;
      submit.disabled=true;
      choices.querySelectorAll('input').forEach(radio=>radio.addEventListener('change',()=>{submit.disabled=!quote||!people.length;}));
     }catch(error){if(own!==generation||attempt!==previewGeneration)return;preview.textContent='Der Zeitraum konnte nicht geprüft werden.';showErrors(error);}
     finally{if(own===generation&&attempt===previewGeneration)check.disabled=false;}
    }
    check.addEventListener('click',loadPreview);
    form.addEventListener('submit',async event=>{
     event.preventDefault();if(!quote)return;
     const current=values();submit.disabled=true;clearErrors();
     try{
      const booked=await post('/api/ausleihen',{...current,bikeID:form.querySelector('[name="bikeID"]:checked')?.value||'',tariffVersion:quote.tariffVersion});
      if(own!==generation)return;notice('Buchung gespeichert. Die tatsächliche Ausgabe bitte separat erfassen.',{href:'#uebersicht/'+encodeURIComponent(booked.id),text:'Zur Ausgabe und Buchungsübersicht'});await render({...current,extras:emptyExtras()});message.focus();
     }catch(error){
      if(own!==generation)return;
      if(error.status===409){await loadPreview();if(own!==generation)return;}
      else submit.disabled=!quote||!form.querySelector('[name="bikeID"]:checked');
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
 render().then(()=>{if(location.hash.startsWith('#uebersicht/'))headingFocus();});
})();
