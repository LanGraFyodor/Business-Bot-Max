import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
try{
 const page=await browser.newPage({viewport:{width:390,height:844}});
 const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.addInitScript(()=>{localStorage.setItem('docspace-user',JSON.stringify('3'));localStorage.setItem('docspace-team',JSON.stringify('studio'))});
 await page.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('documents')));
 await page.goto('http://127.0.0.1:8082',{waitUntil:'domcontentloaded'});
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 const state=await page.evaluate(async()=>{const r=await fetch('/api/state',{headers:{'X-Demo-User':'3'}});return r.json()});
 const studioIds=new Set(state.channels.filter(c=>c.teamId==='studio').map(c=>c.id));
 const source=state.documents.find(d=>studioIds.has(d.channelId)&&d.status==='approved');
 if(!source)throw Error('No approved studio document');
 await page.getByRole('textbox',{name:'Поиск документов'}).fill(source.title);
 await page.getByRole('button',{name:source.title}).first().click();
 await page.getByRole('button',{name:'Передать в Продажи'}).click();
 await page.getByRole('dialog',{name:source.title}).waitFor();
 const forwarded=await page.evaluate(async()=>{const r=await fetch('/api/state',{headers:{'X-Demo-User':'3'}});return r.json()});
 if(!forwarded.documents.some(d=>d.escalatedFrom===source.id&&d.status==='review'))throw Error('forwarded request not present');
 if(errors.length)throw Error(errors.join('; '));
 console.log('Mobile UI: approved studio report escalates to Sales PASS');
}finally{await browser.close()}
