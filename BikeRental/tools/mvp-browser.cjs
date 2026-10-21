'use strict';

// Optional acceptance checks; the application itself needs only Go.
const {chromium, expect} = require(process.env.PLAYWRIGHT_MODULE || 'playwright/test');
const {execFileSync, spawn} = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const assert = require('node:assert/strict');

const project = path.resolve(__dirname, '..');
const base = 'http://127.0.0.1:8080';
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'bikerental-mvp-'));
const binary = path.join(temporary, process.platform === 'win32' ? 'bikerental.exe' : 'bikerental');
const screenshots = path.resolve(project, '../specs/screenshots');
const rates = {
  Downhill: {hourly: 12, days: [40, 70, 95], extra: 25},
  MTB: {hourly: 8, days: [25, 45, 60], extra: 15},
  Rennrad: {hourly: 10, days: [35, 60, 80], extra: 20},
};
let browser, server;

async function start(demo) {
  server = spawn(binary, demo ? ['-demo'] : [], {cwd: temporary, stdio: ['ignore', 'ignore', 'pipe']});
  server.on('error', error => { server.startError = error; });
  let log = '';
  server.stderr.on('data', data => { log += data; });
  for (let attempt = 0; attempt < 100; attempt++) {
    if (server.startError) throw server.startError;
    if (server.exitCode !== null) throw new Error('Testserver vorzeitig beendet: ' + log);
    try { if ((await fetch(base + '/api/dashboard')).ok) return; } catch {}
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  throw new Error('Testserver nicht erreichbar: ' + log);
}
async function stop() {
  if (!server) return;
  const current = server;
  server = null;
  if (current.exitCode === null) {
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => { current.kill('SIGKILL'); reject(new Error('Server-Shutdown fehlgeschlagen')); }, 10000);
      current.once('exit', () => { clearTimeout(timeout); resolve(); });
      current.kill('SIGINT');
    });
  }
  assert.equal(current.exitCode, 0, 'Server muss sauber beenden');
}
async function json(page, endpoint) {
  const response = await page.request.get(base + endpoint);
  assert.equal(response.status(), 200);
  return response.json();
}
async function initial(page, demo) {
  assert.deepEqual(await json(page, '/api/dashboard'), {
    availableBikes: demo ? 3 : 0, rentedBikes: 0,
    totalBikes: demo ? 3 : 0, registeredPeople: demo ? 1 : 0,
  });
  const tariff = await json(page, '/api/tarife');
  assert.equal(tariff.version, demo ? 1 : 0);
  assert.deepEqual(tariff.extras, demo
    ? {helmetCents: 500, protectionCents: 1000, insuranceCents: 800}
    : {helmetCents: 0, protectionCents: 0, insuranceCents: 0});
  if (demo) {
    for (const [kind, rate] of Object.entries(rates)) {
      assert.deepEqual(tariff.types[kind], {
        hourlyCents: rate.hourly * 100,
        pricesCents: rate.days.map(price => price * 100),
        extraDayCents: rate.extra * 100,
      });
    }
  } else assert.deepEqual(tariff.types, {});
  assert.deepEqual(await json(page, '/api/ausleihen'), []);
}
async function setupTariffs(page) {
  await page.goto(base + '/tarife');
  for (const [kind, rate] of Object.entries(rates)) {
    await page.locator('#' + kind + '-hourly').fill(String(rate.hourly));
    await page.locator('#' + kind + '-price-1').fill(String(rate.days[0]));
    for (let day = 2; day <= 3; day++) {
      await page.getByRole('button', {name: kind + ': Weiteren Tagespreis hinzufügen', exact: true}).click();
      await page.locator('#' + kind + '-price-' + day).fill(String(rate.days[day - 1]));
    }
    await page.locator('#' + kind + '-extra').fill(String(rate.extra));
  }
  for (const [key, value] of Object.entries({helmetCents: 5, protectionCents: 10, insuranceCents: 8})) {
    await page.locator('#extras-' + key).fill(String(value));
  }
  await page.getByRole('button', {name: 'Tarife speichern', exact: true}).click();
  await expect(page.locator('#tariff-message')).toContainText('Tarife gespeichert');
}
async function createPerson(page, proof) {
  await page.goto(base + '/personen#neu');
  const name = 'Synthetische MVP-Person ' + proof;
  await page.locator('#name').fill(name);
  await page.locator('#proof').selectOption(proof);
  if (proof === 'manual') {
    await page.locator('#documentType').selectOption('Reisepass');
    await page.locator('#documentNumber').fill('SYNTHETISCH-MVP-05');
  } else {
    await page.locator('#photo').setInputFiles(path.join(project, 'testdata/ausweis-demo.png'));
  }
  await page.getByRole('button', {name: 'Speichern', exact: true}).click();
  await expect(page.getByRole('heading', {name, exact: true})).toBeVisible();
  if (proof === 'photo') {
    await expect(page.locator('.identity-photo')).toHaveJSProperty('complete', true);
    assert(await page.locator('.identity-photo').evaluate(img => img.naturalWidth > 0));
  }
  else await expect(page.locator('#content')).toContainText('SYNTHETISCH-MVP-05');
  const id = decodeURIComponent(new URL(page.url()).hash.slice(1));
  await page.getByRole('link', {name: 'Fahrrad ausleihen', exact: true}).click();
  await expect(page.locator('#personID')).toHaveValue(id);
  return id;
}
async function createBike(page, kind) {
  await page.goto(base + '/fahrraeder#neu');
  const name = 'MVP-Prüfrad ' + kind;
  await page.locator('#name').fill(name);
  await page.locator('#type').selectOption(kind);
  await page.locator('#photo').setInputFiles(path.join(project, 'web/assets/demo/' + kind + '.png'));
  await page.getByRole('button', {name: 'Speichern', exact: true}).click();
  await expect(page.getByRole('heading', {name, exact: true})).toBeVisible();
  return decodeURIComponent(new URL(page.url()).hash.slice(1));
}
async function workflow(page, demo) {
  await initial(page, demo);
  await page.goto(base + '/');
  await expect(page.locator('h1')).toHaveText('Dashboard');
  await expect(page.locator('#dashboard-status')).toContainText('(Wien)');
  // The binary runs from a new temporary directory with no adjacent web assets.
  for (const asset of ['/assets/styles.css', '/assets/dashboard.js', '/assets/ausleihe.js', '/assets/verwaltung.js', '/assets/tarife.js', '/assets/demo/MTB.png']) {
    assert.equal((await page.request.get(base + asset)).status(), 200, asset);
  }
  if (!demo) {
    await expect(page.locator('#dashboard-empty-inventory')).toBeVisible();
    await page.goto(base + '/ausleihe');
    await expect(page.locator('#check-period')).toBeDisabled();
    await setupTariffs(page);
  }
  const people = [await createPerson(page, 'manual'), await createPerson(page, 'photo')];
  const bikes = {};
  for (const kind of Object.keys(rates)) bikes[kind] = await createBike(page, kind);
  const expected = [];
  let sequence = 0;
  for (const kind of Object.keys(rates)) {
    for (const mode of ['day', 'hour']) {
      for (const withExtras of [false, true]) {
        const personID = people[sequence++ % people.length];
        await page.goto(base + '/ausleihe?person=' + personID);
        await expect(page.locator('#personID')).toHaveValue(personID);
        await page.locator('#bikeType').selectOption(kind);
        await page.locator('#mode').selectOption(mode);
        if (mode === 'hour') {
          await expect(page.locator('#startMode')).toHaveValue('now');
          await page.locator('#rentalHours').fill('1');
        }
        for (const key of ['helmet', 'protection', 'insurance']) {
          await expect(page.locator('#extra-' + key)).not.toBeChecked();
          if (withExtras) { await page.locator('#extra-' + key).focus(); await page.keyboard.press('Space'); }
        }
        await page.getByRole('button', {name: 'Verfügbarkeit und Preis prüfen', exact: true}).click();
        const choice = page.locator('[name=bikeID][value="' + bikes[kind] + '"]');
        await expect(choice).toBeVisible();
        const picture = choice.locator('..').locator('img');
        await expect(picture).toHaveJSProperty('complete', true);
        assert(await picture.evaluate(img => img.naturalWidth > 0), 'Fahrradbild nicht dekodiert');
        const basePrice = mode === 'hour' ? rates[kind].hourly : rates[kind].days[0];
        const total = basePrice + (withExtras ? 23 : 0);
        const money = total.toFixed(2).replace('.', ',') + ' €';
        await expect(page.locator('#price-preview .price-total dd')).toHaveText(money);
        if (demo && kind === 'Rennrad' && mode === 'day' && !withExtras) {
          await page.locator('#bikeID').screenshot({path: path.join(screenshots, '05-mvp-fahrradauswahl.png')});
        }
        await choice.focus(); await page.keyboard.press('Space');
        await page.getByRole('button', {name: 'Zum angezeigten Preis buchen', exact: true}).click();
        const link = page.getByRole('link', {name: 'Zur Ausgabe und Buchungsübersicht', exact: true});
        await expect(link).toBeVisible();
        const rentalID = decodeURIComponent((await link.getAttribute('href')).split('/')[1]);
        await link.focus(); await page.keyboard.press('Enter');
        const card = page.locator('#rental-' + rentalID);
        await expect(card).toBeFocused();
        assert.equal((await json(page, '/api/dashboard')).rentedBikes, 0);
        await card.getByRole('button', {name: 'Fahrrad ausgeben', exact: true}).click();
        await expect(page.locator('#rental-message')).toContainText('Ausgabe gespeichert');
        assert.equal((await json(page, '/api/dashboard')).rentedBikes, 1);
        await card.getByRole('button', {name: 'Rückgabe erfassen', exact: true}).click();
        await expect(page.locator('#rental-message')).toContainText('Rückgabe gespeichert');
        await expect(card).toContainText('Zurückgegeben');
        await expect(card.locator('.price-total dd')).toHaveText(money);
        expected.push({id: rentalID, personID, bikeID: bikes[kind], bikeType: kind, mode, totalPriceCents: total * 100, withExtras});
        if (demo && kind === 'MTB' && mode === 'hour' && withExtras) {
          await card.screenshot({path: path.join(screenshots, '05-mvp-preisverlauf.png')});
          await page.setViewportSize({width: 390, height: 844});
          assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
          await card.screenshot({path: path.join(screenshots, '05-mvp-mobil.png')});
          await page.setViewportSize({width: 1440, height: 1000});
        }
      }
    }
  }
  const history = await json(page, '/api/ausleihen');
  assert.equal(history.length, 12);
  for (const want of expected) {
    const got = history.find(record => record.id === want.id);
    for (const key of ['personID', 'bikeID', 'bikeType', 'mode', 'totalPriceCents']) assert.equal(got[key], want[key]);
    assert.equal(got.status, 'returned');
    assert.deepEqual(got.extras, {helmet: want.withExtras, protection: want.withExtras, insurance: want.withExtras});
  }
  const summaries = await json(page, '/api/personen');
  for (const person of summaries) assert.deepEqual(Object.keys(person).sort(), ['id', 'name']);
  assert(!JSON.stringify(history).includes('SYNTHETISCH-MVP-05'), 'Ausweisnummer im Verlauf');
  assert.deepEqual(await json(page, '/api/dashboard'), {
    availableBikes: demo ? 6 : 3, rentedBikes: 0, totalBikes: demo ? 6 : 3, registeredPeople: demo ? 3 : 2,
  });
  const tariff = await json(page, '/api/tarife');
  tariff.extras.helmetCents = 999;
  assert.equal((await page.request.put(base + '/api/tarife', {data: tariff})).status(), 200);
  assert.deepEqual(await json(page, '/api/ausleihen'), history, 'Tarifänderung verändert Verlauf');
  return {people, bikes, rentalID: expected[0].id};
}


function localInput(value) {
  const parts = new Intl.DateTimeFormat('en-CA', {timeZone: 'Europe/Vienna', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23'}).formatToParts(value);
  const get = key => parts.find(part => part.type === key).value;
  return `${get('year')}-${get('month')}-${get('day')}T${get('hour')}:${get('minute')}`;
}
async function immediateRegression(page, demo, saved) {
  // The main workflow deliberately changed the helmet fee to test history.
  const tariff = await json(page, '/api/tarife');
  tariff.extras.helmetCents = 500;
  assert.equal((await page.request.put(base + '/api/tarife', {data: tariff})).status(), 200);
  await page.goto(base + '/ausleihe?person=' + saved.people[0]);
  await page.locator('#bikeType').selectOption('MTB');
  await page.locator('#mode').selectOption('hour');
  await expect(page.locator('#startMode')).toHaveValue('now');
  await expect(page.locator('#fixed-dates')).toBeHidden();
  await page.locator('#rentalHours').fill('0');
  await page.getByRole('button', {name: 'Verfügbarkeit und Preis prüfen', exact: true}).click();
  await expect(page.locator('#rental-error')).toContainText('positive ganze Stundenzahl');
  await page.locator('#rentalHours').fill('2');
  await page.locator('#extra-helmet').check();
  const quoteResponse = page.waitForResponse(response => response.url() === base + '/api/ausleihen/preis' && response.request().method() === 'POST');
  await page.getByRole('button', {name: 'Verfügbarkeit und Preis prüfen', exact: true}).click();
  const quote = await (await quoteResponse).json();
  await expect(page.locator('#price-preview')).toContainText('2 Stunden ab Buchungsabschluss');
  await expect(page.locator('#rental-error')).toBeEmpty();
  await expect(page.locator('#price-preview .price-total dd')).toHaveText('21,00 €');
  await page.locator('[name=bikeID][value="' + saved.bikes.MTB + '"]').check();
  const delay = demo ? 0 : Number(process.env.MVP_CONFIRM_DELAY_MS || 0);
  assert(Number.isInteger(delay) && delay >= 0 && delay <= 300000, 'MVP_CONFIRM_DELAY_MS: 0 bis 300000');
  for (let remaining = delay; remaining > 0; remaining -= 30000) {
    console.log('Sofortbeginn: Wartezeit vor Abschluss, noch ' + remaining + ' ms');
    await page.waitForTimeout(Math.min(30000, remaining));
  }
  const bookingResponse = page.waitForResponse(response => response.url() === base + '/api/ausleihen' && response.request().method() === 'POST');
  await page.getByRole('button', {name: 'Zum angezeigten Preis buchen', exact: true}).click();
  const response = await bookingResponse;
  assert.equal(response.status(), 201);
  const rental = await response.json();
  assert(new Date(rental.startsAt) - new Date(quote.startsAt) >= delay, 'Beginn wurde aus alter Vorschau übernommen');
  assert.equal(rental.startsAt, rental.createdAt);
  assert.equal(new Date(rental.endsAt) - new Date(rental.startsAt), 2 * 3600000);
  assert.equal(rental.totalPriceCents, 2100);
  await expect(page.locator('#rental-message')).toContainText('Mietzeit läuft von');
  await page.getByRole('link', {name: 'Zur Ausgabe und Buchungsübersicht', exact: true}).click();
  const card = page.locator('#rental-' + rental.id);
  await expect(card).toBeFocused();
  await expect(card.getByRole('button', {name: 'Fahrrad ausgeben', exact: true})).toBeVisible();
  if (demo) await card.screenshot({path: path.join(screenshots, '06-sofortbeginn.png')});
  await card.getByRole('button', {name: 'Fahrrad ausgeben', exact: true}).click();
  await expect(page.locator('#rental-message')).toContainText('Ausgabe gespeichert');
  await card.getByRole('button', {name: 'Rückgabe erfassen', exact: true}).click();
  await expect(page.locator('#rental-message')).toContainText('Rückgabe gespeichert');

  // Fixed future dates keep their explicit issue restriction.
  await page.goto(base + '/ausleihe?person=' + saved.people[0]);
  await page.locator('#bikeType').selectOption('MTB');
  await page.locator('#mode').selectOption('hour');
  await page.locator('#startMode').selectOption('scheduled');
  await expect(page.locator('#rentalHours')).toBeDisabled();
  await expect(page.locator('#fixed-dates')).toBeVisible();
  const future = new Date(Date.now() + 10 * 60000);
  await page.locator('#startsOn').fill(localInput(new Date(Date.now() - 5 * 60000)));
  await page.locator('#endsOn').fill(localInput(new Date(future.getTime() + 3600000)));
  await page.getByRole('button', {name: 'Verfügbarkeit und Preis prüfen', exact: true}).click();
  await expect(page.locator('#rental-error')).toContainText('nicht vor');
  await page.locator('#startsOn').fill(localInput(future));
  await page.getByRole('button', {name: 'Verfügbarkeit und Preis prüfen', exact: true}).click();
  await page.locator('[name=bikeID][value="' + saved.bikes.MTB + '"]').check();
  await expect(page.locator('#rental-error')).toBeEmpty();
  await page.getByRole('button', {name: 'Zum angezeigten Preis buchen', exact: true}).click();
  await page.getByRole('link', {name: 'Zur Ausgabe und Buchungsübersicht', exact: true}).click();
  const fixed = (await json(page, '/api/ausleihen')).find(record => record.startMode === 'scheduled' && record.mode === 'hour' && record.status === 'booked');
  assert(fixed);
  const futureCard = page.locator('#rental-' + fixed.id);
  await expect(futureCard.getByRole('button', {name: 'Fahrrad ausgeben', exact: true})).toHaveCount(0);
  assert.equal((await page.request.put(base + '/api/ausleihen/' + fixed.id + '/ausgabe')).status(), 409);
  await futureCard.getByRole('button', {name: 'Buchung stornieren', exact: true}).click();
  await expect(page.locator('#rental-message')).toContainText('Buchung storniert');
  await page.goto(base + '/ausleihe?person=' + saved.people[0]);
  await page.locator('#mode').selectOption('hour');
  await page.setViewportSize({width: 390, height: 844});
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
  if (demo) await page.locator('#immediate-duration').screenshot({path: path.join(screenshots, '06-sofortbeginn-mobil.png')});
  await page.setViewportSize({width: 1440, height: 1000});
  console.log((demo ? 'demo' : 'normal') + ': Sofortbeginn nach Wartezeit, volle Dauer, sofortige Ausgabe, feste Reservierung und mobile Eingabe: PASS');
}

(async () => {
  try { await fetch(base + '/api/dashboard'); throw new Error('Port 8080 belegt; vorhandenen Server zuerst beenden.'); }
  catch (error) { if (error.message.startsWith('Port 8080')) throw error; }
  fs.mkdirSync(screenshots, {recursive: true});
  execFileSync(process.env.GO_BINARY || 'go', ['build', '-o', binary, './cmd/server'], {cwd: project, stdio: 'inherit'});
  browser = await chromium.launch({headless: true, ...(process.env.CHROME_EXECUTABLE_PATH ? {executablePath: process.env.CHROME_EXECUTABLE_PATH} : {})});
  for (const demo of [false, true]) {
    const context = await browser.newContext({viewport: {width: 1440, height: 1000}, timezoneId: 'America/New_York'});
    const page = await context.newPage(), errors = [];
    page.on('pageerror', error => errors.push(error.message));
    try {
      await start(demo);
      const saved = await workflow(page, demo);
      await immediateRegression(page, demo, saved);
      await stop(); await start(demo);
      await initial(page, demo);
      for (const endpoint of ['/api/personen/' + saved.people[1] + '/foto', '/api/fahrraeder/' + saved.bikes.MTB + '/foto', '/api/ausleihen/' + saved.rentalID]) {
        assert.equal((await page.request.get(base + endpoint)).status(), 404, 'Neustart behält Daten: ' + endpoint);
      }
      await page.goto(base + '/');
      if (demo) await expect(page.locator('#dashboard-empty-inventory')).toBeHidden();
      else await expect(page.locator('#dashboard-empty-inventory')).toBeVisible();
      if (demo) {
        const bikes = await json(page, '/api/fahrraeder');
        for (const bike of bikes) assert.equal((await page.request.get(base + bike.photoURL)).status(), 200);
        await page.screenshot({path: path.join(screenshots, '05-mvp-neustart.png'), fullPage: true});
      }
      assert.deepEqual(errors, [], 'Unbehandelte JavaScript-Fehler');
      console.log((demo ? 'demo' : 'normal') + ': PASS – 12 Verleihabläufe, beide Ausweisnachweise, eingebettete Assets, Tarifhistorie und Prozessneustart');
    } finally { await context.close(); await stop(); }
  }
})().catch(error => { console.error(error); process.exitCode = 1; }).finally(async () => {
  try { await stop(); } finally {
    if (browser) await browser.close();
    fs.rmSync(temporary, {recursive: true, force: true});
  }
});
