import {fileURLToPath} from 'node:url';
import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
try {
 const page=await browser.newPage({viewport:{width:1440,height:960}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://127.0.0.1:8080');await page.locator('.provider-card').first().waitFor();
 await page.screenshot({path:fileURLToPath(new URL('../artifacts/redesign-desktop.png',import.meta.url))});
 await page.locator('.provider-card.provider-1c').click();await page.getByRole('button',{name:'Загрузить отчёт',exact:true}).first().click();await page.locator('.dropzone').waitFor();
 await page.getByRole('button',{name:'Настроить обмен',exact:true}).click();await page.getByRole('button',{name:'Создать ключ',exact:true}).waitFor();
 await page.setViewportSize({width:390,height:844});await page.waitForTimeout(300);await page.screenshot({path:fileURLToPath(new URL('../artifacts/redesign-mobile.png',import.meta.url))});
 await page.locator('.product-mobile-nav').getByRole('button',{name:'Команды',exact:true}).click();await page.locator('.organisation-tree').waitFor();
 await page.locator('.product-mobile-nav').getByRole('button',{name:'Документы',exact:true}).click();await page.locator('.document-list').waitFor();
 if(errors.length)throw Error(errors.join('; '));console.log('Desktop/mobile: systems, 1C upload, connection, organisation, documents PASS');
} finally {await browser.close()}
