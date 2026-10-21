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
  const own=++generation;content.innerHTML='<p>Tarif wird geladen …</p>';
  try{
   const tariff=await request();if(own!==generation)return;
   content.innerHTML=`<div class="section-heading"><h2>Gemeinsame Tarifstaffel</h2><button class="secondary" id="reload-tariff">Tarif neu laden</button></div><p id="tariff-error" class="field-error" role="alert" tabindex="-1"></p><form class="record-panel" id="tariff-form" novalidate><p>${tariff.version?'Aktueller Tarif · Version '+tariff.version:'Noch kein Tarif eingerichtet.'} Alle Preise gelten pro Fahrrad in EUR. Bestehende Buchungspreise bleiben unverändert.</p><p>Gesamtpreise für 1 bis N Tage; längere Zeiträume: letzter Gesamtpreis plus Preis je Zusatztag.</p><div id="tariff-rows"></div><div class="actions"><button type="button" class="secondary" id="add-day">Weiteren Tagespreis hinzufügen</button><button type="button" class="secondary" id="remove-day">Letzten Tagespreis entfernen</button></div><div class="field"><label for="extraDayCents">Preis je Zusatztag in EUR (Pflichtfeld)</label><input id="extraDayCents" inputmode="decimal" value="${euro(tariff.extraDayCents)}"></div><div class="actions"><button type="submit">Tarif speichern</button><a class="button secondary" href="/ausleihe">Zur Ausleihe</a></div></form>`;
   const rows=content.querySelector('#tariff-rows'),form=content.querySelector('form');
   const add=value=>{const day=rows.children.length+1;rows.insertAdjacentHTML('beforeend',`<div class="field"><label for="price-${day}">Gesamtpreis für ${day} ${day===1?'Tag':'Tage'} in EUR (Pflichtfeld)</label><input id="price-${day}" inputmode="decimal" value="${esc(value)}"></div>`);};
   (tariff.pricesCents.length?tariff.pricesCents:[0]).forEach(value=>add(euro(value)));
   const remove=content.querySelector('#remove-day');remove.disabled=rows.children.length===1;
   content.querySelector('#add-day').addEventListener('click',()=>{add('');remove.disabled=false;rows.lastElementChild.querySelector('input').focus();});
   remove.addEventListener('click',()=>{if(rows.children.length>1)rows.lastElementChild.remove();remove.disabled=rows.children.length===1;rows.lastElementChild.querySelector('input').focus();});
   content.querySelector('#reload-tariff').addEventListener('click',()=>{message.textContent='';message.className='';load();});
   form.addEventListener('submit',async event=>{
    event.preventDefault();const button=form.querySelector('[type="submit"]');button.disabled=true;
    try{
     const data={version:tariff.version,pricesCents:[...rows.querySelectorAll('input')].map(input=>cents(input.value)),extraDayCents:cents(content.querySelector('#extraDayCents').value)};
     await request({method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
     if(own!==generation)return;message.textContent='Tarif gespeichert. Neue Buchungen verwenden diese Preise.';message.className='success';await load();message.focus();
    }catch(err){if(own!==generation)return;error(err);button.disabled=false;}
   });
  }catch(err){if(own!==generation)return;content.innerHTML='<p id="tariff-error" class="field-error" role="alert" tabindex="-1"></p><button id="retry-tariff">Erneut laden</button>';error(err);content.querySelector('button').addEventListener('click',load);}
 }
 load();
})();
