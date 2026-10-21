'use strict';
(() => {
 if(!document.querySelector('#tarife'))return;
 const content=document.querySelector('#tariff-content'),message=document.querySelector('#tariff-message');
 const esc=v=>String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
 let generation=0;
 document.querySelector('.skip-link').addEventListener('click',e=>{e.preventDefault();document.querySelector('main').focus();});
 async function request(options){const response=await fetch('/api/tarife',{cache:'no-store',...options});const data=await response.json();if(!response.ok)throw data;return data;}
 function error(err){const el=content.querySelector('#tariff-error');el.textContent=Object.values(err.errors||{form:'Die Verbindung ist fehlgeschlagen. Bitte erneut versuchen.'}).join(' ');el.focus();}
 function euro(cents){return (BigInt(cents)/100n).toString()+','+(BigInt(cents)%100n).toString().padStart(2,'0');}
 function cents(value){
  const text=value.trim();if(!/^\d+(?:[,.]\d{1,2})?$/.test(text))throw {errors:{form:'Preise bitte als nichtnegative Eurobeträge mit höchstens zwei Nachkommastellen eingeben.'}};
  const [whole,fraction='']=text.replace(',','.').split('.');const result=BigInt(whole)*100n+BigInt(fraction.padEnd(2,'0'));
  if(result>BigInt(Number.MAX_SAFE_INTEGER))throw {errors:{form:'Der eingegebene Preis ist zu groß.'}};return Number(result);
 }
 async function load(){
  const own=++generation;content.innerHTML='<p>Tarife werden geladen …</p>';
  try{
   const tariff=await request();if(own!==generation)return;
   const kinds=['Downhill','MTB','Rennrad'];
   const extraOptions=[['helmetCents','Helm'],['protectionCents','Schutzausrüstung'],['insuranceCents','Versicherung']];
   content.innerHTML=`<div class="section-heading"><h2>Tarife je Fahrradtyp</h2><button class="secondary" id="reload-tariff">Tarife neu laden</button></div><p id="tariff-error" class="field-error" role="alert" tabindex="-1"></p><form class="record-panel" id="tariff-form" novalidate><p>${tariff.version?'Aktuelle Tarife · Version '+tariff.version:'Noch keine Tarife eingerichtet.'} Alle Preise in EUR. Bestehende Buchungspreise bleiben unverändert.</p><p>Stundenmiete: jede angefangene Stunde. Tagesmiete: Gesamtpreise für 1 bis N Kalendertage, danach Preis je Zusatztag. Alle drei Typen gemeinsam speichern.</p>${kinds.map(kind=>{
    const r=tariff.types[kind];
    return `<fieldset class="tariff-type" data-kind="${kind}"><legend>${kind}</legend><div class="field"><label for="${kind}-hourly">${kind}: Preis je Stunde in EUR (Pflichtfeld)</label><input id="${kind}-hourly" data-error-key="${kind}.hourlyCents" inputmode="decimal" value="${r?euro(r.hourlyCents):''}"></div><div class="tariff-rows"></div><div class="actions"><button type="button" class="secondary add-day">${kind}: Weiteren Tagespreis hinzufügen</button><button type="button" class="secondary remove-day">${kind}: Letzten Tagespreis entfernen</button></div><div class="field"><label for="${kind}-extra">${kind}: Preis je Zusatztag in EUR (Pflichtfeld)</label><input id="${kind}-extra" data-error-key="${kind}.extraDayCents" inputmode="decimal" value="${r?euro(r.extraDayCents):''}"></div></fieldset>`;
   }).join('')}<fieldset class="extra-tariffs"><legend>Zusatzleistungen · Pauschalen pro Buchung</legend><p>Einmalig pro gewählter Option, für alle Fahrradtypen und Mietarten. Auch ausdrücklich kostenlose Optionen (0 EUR) sind möglich.</p>${extraOptions.map(([key,name])=>`<div class="field"><label for="extras-${key}">${name}: Pauschale in EUR (Pflichtfeld)</label><input id="extras-${key}" data-error-key="extras.${key}" inputmode="decimal" value="${tariff.version?euro(tariff.extras[key]):''}"></div>`).join('')}</fieldset><div class="actions"><button type="submit">Tarife speichern</button><a class="button secondary" href="/ausleihe">Zur Ausleihe</a></div></form>`;
   for(const section of content.querySelectorAll('.tariff-type')){
    const kind=section.dataset.kind,rows=section.querySelector('.tariff-rows'),remove=section.querySelector('.remove-day');
    const add=value=>{const day=rows.children.length+1;rows.insertAdjacentHTML('beforeend',`<div class="field"><label for="${kind}-price-${day}">${kind}: Gesamtpreis für ${day} ${day===1?'Tag':'Tage'} in EUR (Pflichtfeld)</label><input id="${kind}-price-${day}" data-error-key="${kind}.pricesCents" inputmode="decimal" value="${esc(value)}"></div>`);};
    const rate=tariff.types[kind];
    if(rate)rate.pricesCents.forEach(value=>add(euro(value)));else add('');
    remove.disabled=rows.children.length===1;
    section.querySelector('.add-day').addEventListener('click',()=>{add('');remove.disabled=false;rows.lastElementChild.querySelector('input').focus();});
    remove.addEventListener('click',()=>{if(rows.children.length>1)rows.lastElementChild.remove();remove.disabled=rows.children.length===1;rows.lastElementChild.querySelector('input').focus();});
   }
   content.querySelector('#reload-tariff').addEventListener('click',()=>{message.textContent='';message.className='';load();});
   content.querySelector('form').addEventListener('submit',async event=>{
    event.preventDefault();const button=event.currentTarget.querySelector('[type="submit"]');button.disabled=true;
    content.querySelectorAll('[aria-invalid]').forEach(el=>el.removeAttribute('aria-invalid'));
    try{
     const types={};
     for(const section of content.querySelectorAll('.tariff-type')){
      const kind=section.dataset.kind;
      types[kind]={hourlyCents:cents(section.querySelector('#'+kind+'-hourly').value),pricesCents:[...section.querySelectorAll('.tariff-rows input')].map(el=>cents(el.value)),extraDayCents:cents(section.querySelector('#'+kind+'-extra').value)};
     }
     await request({method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({version:tariff.version,types,extras:Object.fromEntries(extraOptions.map(([key])=>[key,cents(content.querySelector('#extras-'+key).value)]))})});
     if(own!==generation)return;message.textContent='Tarife gespeichert. Neue Buchungen verwenden diese Preise.';message.className='success';await load();message.focus();
    }catch(err){if(own!==generation)return;
     for(const el of content.querySelectorAll('[data-error-key]'))if(err.errors?.[el.dataset.errorKey])el.setAttribute('aria-invalid','true');
     error(err);button.disabled=false;
    }
   });
  }catch(err){if(own!==generation)return;content.innerHTML='<p id="tariff-error" class="field-error" role="alert" tabindex="-1"></p><button id="retry-tariff">Erneut laden</button>';error(err);content.querySelector('button').addEventListener('click',load);}
 }
 load();
})();
