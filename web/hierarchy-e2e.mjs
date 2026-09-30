import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
try{
 const page=await browser.newPage({viewport:{width:390,height:844}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.addInitScript(()=>{localStorage.setItem('docspace-user',JSON.stringify('4'));localStorage.setItem('docspace-section',JSON.stringify('systems'))});
 await page.goto('http://127.0.0.1:8082');await page.locator('.provider-card.provider-1c').click();
 await page.getByRole('button',{name:'Демо-выгрузка',exact:true}).click();
 await page.getByRole('dialog',{name:/1С · новый счёт поставщика/}).waitFor();await page.getByText('Источник: 1С').waitFor();
 const managerPage=await browser.newPage({viewport:{width:390,height:844}});
 await managerPage.addInitScript(()=>{localStorage.setItem('docspace-user',JSON.stringify('5'));localStorage.setItem('docspace-team',JSON.stringify('demo-sales'));localStorage.setItem('docspace-section',JSON.stringify('teams'))});
 await managerPage.goto('http://127.0.0.1:8082');await managerPage.locator('.org-node-locked').filter({hasText:'Дизайн-студия'}).waitFor();
 if(errors.length)throw Error(errors.join('; '));console.log('Mobile UI: reports demo and protected organisation tree PASS');
}finally{await browser.close()}
