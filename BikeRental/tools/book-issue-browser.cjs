'use strict';
// Run against a fresh local demo server. Creates synthetic bookings and changes a demo fee.
const {chromium,expect}=require(process.env.PLAYWRIGHT_MODULE||'playwright/test');
const assert=require('node:assert/strict'),path=require('node:path');
(async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.CHROME_EXECUTABLE_PATH?{executablePath:process.env.CHROME_EXECUTABLE_PATH}:{})});
 try{
  const page=await browser.newPage({viewport:{width:1280,height:1000}}),errors=[],base='http://127.0.0.1:8080';
  page.on('pageerror',e=>errors.push(e.message));
  const day=date=>new Intl.DateTimeFormat('en-CA',{timeZone:'Europe/Vienna',year:'numeric',month:'2-digit',day:'2-digit'}).format(date);
  const json=async url=>(await page.request.get(base+url)).json();
  async function setup(mode){
   await page.goto(base+'/ausleihe?person=demo-person-01');await expect(page.locator('#personID')).toHaveValue('demo-person-01');
   await page.locator('#bikeType').selectOption('MTB');await page.locator('#mode').selectOption(mode);
   await expect(page.locator('#book-rental')).toHaveText('Buchen und ausgeben');await expect(page.locator('#book-only')).toBeVisible();
  }
  async function preview(){await page.locator('#check-period').click();await page.locator('[name=bikeID][value="demo-bike-02"]').check();await expect(page.locator('#book-rental')).toBeEnabled();}
  for(const mode of ['day','hour'])for(const issue of [true,false]){
   await setup(mode);await page.locator('#extra-helmet').check();await preview();
   if(issue){
    if(mode==='hour')await page.setViewportSize({width:390,height:844});
    await page.locator('#book-rental').scrollIntoViewIfNeeded();assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    await page.screenshot({path:path.resolve(__dirname,'../../specs/screenshots/07-buchen-ausgeben-'+(mode==='hour'?'mobil':'desktop')+'.png'),fullPage:true});
   }
   let release;const held=new Promise(resolve=>release=resolve);
   await page.route('**/api/ausleihen',async route=>{if(route.request().method()==='POST')await held;await route.continue();});
   const response=page.waitForResponse(r=>r.url()===base+'/api/ausleihen'&&r.request().method()==='POST');
   await page.locator(issue?'#book-rental':'#book-only').focus();await page.keyboard.press('Enter');
   await expect(page.locator('#book-rental')).toBeDisabled();await expect(page.locator('#book-only')).toBeDisabled();release();
   const result=await response;assert.equal(result.status(),201);const rental=await result.json();
   assert.equal(rental.status,issue?'issued':'booked');assert.equal(!!rental.rentedAt,issue);assert.equal(rental.extras.helmet,true);
   if(mode==='hour')assert.equal(new Date(rental.endsAt)-new Date(rental.startsAt),3600000);
   await expect(page.locator('#rental-message')).toContainText(issue?'Buchung und Ausgabe gespeichert.':'Buchung gespeichert.');
   await page.unroute('**/api/ausleihen');assert.equal((await json('/api/dashboard')).rentedBikes,issue?1:0);
   if(issue){
    await page.getByRole('link',{name:'Zur Buchungsübersicht',exact:true}).click();await expect(page.locator('#rental-'+rental.id)).toBeFocused();
    await page.locator('#rental-'+rental.id).getByRole('button',{name:'Rückgabe erfassen',exact:true}).click();await expect(page.locator('#rental-message')).toContainText('Rückgabe gespeichert');
   }else assert.equal((await page.request.put(base+'/api/ausleihen/'+rental.id+'/stornierung')).status(),200);
   await page.setViewportSize({width:1280,height:1000});
  }
  await setup('day');const future=day(new Date(Date.now()+86400000));
  await page.locator('#startsOn').fill(future);await page.locator('#endsOn').fill(future);
  await expect(page.locator('#book-rental')).toHaveText('Buchen');await expect(page.locator('#book-only')).toBeHidden();
  await page.locator('#startsOn').fill(day(new Date()));await expect(page.locator('#book-only')).toBeVisible();
  await page.locator('#mode').selectOption('hour');await page.locator('#startMode').selectOption('scheduled');
  await expect(page.locator('#book-rental')).toHaveText('Buchen');await expect(page.locator('#book-only')).toBeHidden();
  await page.locator('#startMode').selectOption('now');await expect(page.locator('#book-only')).toBeVisible();
  await preview();const before=await json('/api/ausleihen'),tariff=await json('/api/tarife');tariff.extras.helmetCents++;
  assert.equal((await page.request.put(base+'/api/tarife',{data:tariff})).status(),200);
  const conflict=page.waitForResponse(r=>r.url()===base+'/api/ausleihen'&&r.request().method()==='POST');await page.locator('#book-rental').click();assert.equal((await conflict).status(),409);
  await expect(page.locator('#rental-error')).toContainText('Tarif wurde geändert');await expect(page.locator('#book-rental')).toBeDisabled();await expect(page.locator('#book-only')).toBeDisabled();
  assert.deepEqual(await json('/api/ausleihen'),before);await page.locator('[name=bikeID][value="demo-bike-02"]').check();await expect(page.locator('#book-only')).toBeEnabled();
  assert.deepEqual(errors,[]);console.log('PASS: day/hour × issue/book, keyboard, busy buttons, mobile, future modes, tariff conflict');
 }finally{await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
