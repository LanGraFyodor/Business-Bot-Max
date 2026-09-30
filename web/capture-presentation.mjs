import {chromium} from '@playwright/test';
import {mkdir} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';

const base = process.env.PRESENTATION_BASE_URL || 'http://127.0.0.1:8080';
const folder = fileURLToPath(new URL('../presentation/screenshots/', import.meta.url));
await mkdir(folder, {recursive: true});
const browser = await chromium.launch({channel: 'msedge', headless: true});
const saved = [];
const failures = [];
async function capture(page, name) {
  await page.evaluate(() => document.fonts.ready);
  const path = fileURLToPath(new URL('../presentation/screenshots/' + name, import.meta.url));
  await page.screenshot({path, animations: 'disabled'});
  saved.push(name);
}
async function openPage(viewport, start) {
  const page = await browser.newPage({viewport, deviceScaleFactor: viewport.width < 600 ? 2 : 1});
  page.on('pageerror', error => failures.push(error.message));
  if (start) await page.addInitScript(start);
  return page;
}
try {
  const launcher = await openPage({width: 390, height: 844});
  await launcher.goto(base + '/apps');
  await launcher.getByRole('link', {name: /Бизнес-навигатор/}).waitFor();
  await capture(launcher, '01-one-bot-two-apps-mobile.png');

  const docs = await openPage({width: 390, height: 844}, () => {
    localStorage.setItem('docspace-section', JSON.stringify('documents'));
    localStorage.setItem('docspace-team', JSON.stringify('orbita'));
  });
  await docs.goto(base + '/documents');
  await docs.getByRole('heading', {name: 'Все документы'}).waitFor();
  await docs.locator('.document-row').first().waitFor();
  await capture(docs, '02-documents-mobile.png');
  await docs.locator('.product-mobile-nav').getByRole('button', {name: /Проверка/}).click();
  await docs.getByRole('heading', {name: 'На согласовании'}).waitFor();
  await capture(docs, '03-review-queue-mobile.png');
  await docs.locator('.document-row').filter({has: docs.locator('.file-icon small', {hasText: 'DOCX'})}).first().click();
  await docs.getByRole('dialog').waitFor();
  await docs.locator('.office-document').waitFor();
  await capture(docs, '04-document-decision-mobile.png');

  const systems = await openPage({width: 1440, height: 900}, () => {
    localStorage.setItem('docspace-section', JSON.stringify('systems'));
  });
  await systems.goto(base + '/documents');
  await systems.locator('.provider-card.provider-1c').waitFor();
  await capture(systems, '05-integrations-desktop.png');
  await systems.locator('.provider-card.provider-1c').click();
  await systems.getByRole('heading', {name: /1С/}).first().waitFor();
  await capture(systems, '06-1c-report-desktop.png');

  const teams = await openPage({width: 1440, height: 900}, () => {
    localStorage.setItem('docspace-section', JSON.stringify('teams'));
  });
  await teams.goto(base + '/documents');
  await teams.locator('.organisation-tree').waitFor();
  await capture(teams, '07-team-hierarchy-desktop.png');

  const guide = await openPage({width: 390, height: 844});
  await guide.goto(base + '/guide');
  await guide.getByRole('heading', {name: /Свое дело/}).waitFor();
  await capture(guide, '08-business-guide-mobile.png');
  const nav = guide.getByRole('navigation', {name: 'Мобильная навигация справочника'});
  await nav.getByRole('button', {name: 'План запуска'}).click();
  await guide.getByRole('heading', {name: 'План запуска своего дела'}).waitFor();
  await capture(guide, '09-business-plan-mobile.png');
  await nav.getByRole('button', {name: 'Справочник'}).click();
  await guide.getByRole('tab', {name: 'Законы'}).click();
  await guide.locator('.guide-material-card').first().click();
  await guide.getByRole('dialog').waitFor();
  await capture(guide, '10-source-evidence-mobile.png');

  const guideDesktop = await openPage({width: 1440, height: 900});
  await guideDesktop.goto(base + '/guide');
  await guideDesktop.getByRole('heading', {name: /Свое дело/}).waitFor();
  await capture(guideDesktop, '11-business-guide-desktop.png');

  if (failures.length) throw new Error('Browser errors: ' + failures.join('; '));
  console.log('Captured ' + saved.length + ' screenshots in presentation/screenshots');
  console.log(saved.join('\n'));
} finally {
  await browser.close();
}
