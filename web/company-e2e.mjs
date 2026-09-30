import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
try{
 const page=await browser.newPage({viewport:{width:390,height:844}});
 const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('documents')));
 await page.goto('http://127.0.0.1:8082',{waitUntil:'domcontentloaded'});
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 const teamButtons=page.locator('select[aria-label="Рабочая команда"] option');
 if(await teamButtons.count()<7)throw Error('company teams missing');
 await page.getByRole('combobox',{name:'Рабочая команда'}).selectOption('demo-fin');
 await page.getByRole('button',{name:'Закрыть каналы'}).click();
 await page.locator('.document-row').first().waitFor();
 if(await page.locator('.document-row').count()<25)throw Error('finance document set is too small');
 await page.getByRole('textbox',{name:'Поиск документов'}).fill('Счёт поставщика');
 if(await page.locator('.document-row').count()<5)throw Error('finance search failed');
 await page.getByRole('button',{name:'Очистить поиск'}).click();
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 await page.locator('.navigation.is-open .user-panel').click();
 const profiles=page.getByRole('combobox',{name:'Проверить от лица'});
 if(await profiles.locator('option').count()<30)throw Error('demo roles missing');
 await profiles.selectOption('4');
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 if(await page.locator('.document-row').count()<25)throw Error('finance manager cannot see documents');
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 await page.locator('.navigation.is-open .user-panel').click();
 await page.getByRole('combobox',{name:'Проверить от лица'}).selectOption('2');
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 if(await page.locator('select[aria-label="Рабочая команда"] option[value="demo-fin"]').count()!==0)throw Error('team leaked to Anna');
 if(errors.length)throw Error('browser errors: '+errors.join('; '));
 console.log('Company demo: teams, 25 finance documents, search, 30 profiles, access control PASS');
}finally{await browser.close()}

