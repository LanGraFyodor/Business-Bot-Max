import {chromium} from '@playwright/test';
import assert from 'node:assert/strict';
import {fileURLToPath} from 'node:url';
import {createHash} from 'node:crypto';
import {readFile} from 'node:fs/promises';
const browser=await chromium.launch({channel:'msedge',headless:true});
try{
 const page=await browser.newPage({viewport:{width:1440,height:960}});const errors=[];const api=[];
 page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.url().includes('/api/'))api.push(r.url())});
 await page.goto('http://127.0.0.1:8080/guide');await page.getByRole('heading',{name:/Свое дело/}).waitFor();
 assert(!/\?{3}/.test(await page.locator('body').innerText()));
 await page.screenshot({path:fileURLToPath(new URL('../artifacts/guide-desktop.png',import.meta.url)),fullPage:true});
 await page.getByRole('textbox',{name:'Найти в справочнике'}).fill('работников');await page.getByRole('button',{name:'Искать',exact:true}).click();
 await page.locator('.guide-material-card').first().click();await page.getByRole('dialog').waitFor();
 const href=await page.getByRole('link',{name:'Скачать текст из базы'}).getAttribute('href');
 const download=page.waitForEvent('download');await page.getByRole('link',{name:'Скачать текст из базы'}).click();const file=await download;
 const bytes=await readFile(await file.path());const expected=await readFile(fileURLToPath(new URL('../web/public'+href,import.meta.url)));
 assert.equal(createHash('sha256').update(bytes).digest('hex'),createHash('sha256').update(expected).digest('hex'));
 await page.getByRole('button',{name:'Закрыть материал'}).click();
 await page.getByRole('navigation',{name:'Разделы справочника',exact:true}).getByRole('button',{name:'План запуска'}).click();
 await page.getByRole('checkbox').first().check();await page.reload();
 await page.getByRole('navigation',{name:'Разделы справочника',exact:true}).getByRole('button',{name:'План запуска'}).click();assert(await page.getByRole('checkbox').first().isChecked());
 await page.setViewportSize({width:390,height:844});
 const mobile=page.getByRole('navigation',{name:'Мобильная навигация справочника'});
 await mobile.getByRole('button',{name:'Примеры',exact:true}).click();await page.getByRole('combobox').selectOption({index:9});assert.equal(await page.locator('.guide-products article').count(),7);
 assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false);
 await mobile.getByRole('button',{name:'Обзор',exact:true}).click();await page.screenshot({path:fileURLToPath(new URL('../artifacts/guide-mobile.png',import.meta.url)),fullPage:true});
 assert.equal(api.length,0,'Guide must not request private Docspace API');assert.deepEqual(errors,[]);
 await page.goto('http://127.0.0.1:8080/apps?WebAppStartParam=guide');await page.getByRole('heading',{name:/Свое дело/}).waitFor();
 await page.goto('http://127.0.0.1:8080/apps');await page.getByRole('link',{name:/Бизнес-навигатор/}).click();await page.getByRole('heading',{name:/Свое дело/}).waitFor();
 await page.goto('http://127.0.0.1:8080/documents');await page.locator('.product-navigation').waitFor();
 console.log('PASS: guide search, source download bytes, checklist persistence, examples, mobile, launch parameter, separate Docspace');
}finally{await browser.close()}


