import {chromium} from '@playwright/test';
import {fileURLToPath} from 'node:url';
const browser=await chromium.launch({channel:'msedge',headless:true});
const page=await browser.newPage({viewport:{width:390,height:844},acceptDownloads:true});
const errors=[];page.on('pageerror',e=>errors.push(e.message));
async function open(title){await page.locator('.document-main').filter({hasText:title}).first().click();await page.getByRole('dialog',{name:title}).waitFor()}
async function close(){await page.getByRole('button',{name:'Назад к документам'}).click()}
async function download(){const pending=page.waitForEvent('download');await page.getByRole('button',{name:'Скачать документ'}).click();const result=await pending;const stream=await result.createReadStream();const chunks=[];for await(const chunk of stream)chunks.push(chunk);return {name:result.suggestedFilename(),bytes:Buffer.concat(chunks)}}
try{
 await page.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('documents')));
 await page.goto('http://127.0.0.1:8082',{waitUntil:'domcontentloaded'});
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 await open('Договор на разработку сайта');
 await page.locator('.office-document h3').getByText('Договор на разработку сайта').waitFor();
 await page.getByText('Проектирование',{exact:true}).waitFor();
 await page.screenshot({path:fileURLToPath(new URL('../artifacts/office-word-mobile.png',import.meta.url))});
 const word=await download();if(!word.name.endsWith('.docx')||word.bytes.subarray(0,2).toString()!=='PK')throw Error('Word original failed');
 await close();
 await open('Бюджет проекта');
 await page.locator('.office-workbook').waitFor();
 await page.getByRole('tab',{name:'План оплат'}).click();
 await page.getByText('350000',{exact:true}).waitFor();
 await page.screenshot({path:fileURLToPath(new URL('../artifacts/office-budget-mobile.png',import.meta.url))});
 await close();
 await open('Отчёт по продажам');
 await page.locator('.office-grid').getByText('Меридиан').waitFor();
 await close();
 await open('Счёт на оплату');
 await page.getByTitle('Предпросмотр PDF').waitFor();
 const pdf=await download();if(!pdf.name.endsWith('.pdf')||pdf.bytes.subarray(0,4).toString()!=='%PDF')throw Error('PDF original failed');
 await close();
 await page.getByRole('button',{name:'Загрузить документ',exact:true}).first().click();
 await page.locator('.modal input[type=file]').setInputFiles(fileURLToPath(new URL('../internal/app/demo_assets/contract.docx',import.meta.url)));
 await page.getByRole('button',{name:'Загрузить документ'}).last().click();
 await page.locator('.office-document h3').waitFor();
 if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw Error('mobile horizontal overflow');
 if(errors.length)throw Error('JS errors: '+errors.join('; '));
 console.log('Mobile PDF, DOCX, XLSX, CSV preview and original downloads; DOCX upload: PASS');
}finally{await browser.close()}