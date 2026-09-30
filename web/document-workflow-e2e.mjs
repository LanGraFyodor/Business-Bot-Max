import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
const page=await browser.newPage({viewport:{width:390,height:844},acceptDownloads:true});
const errors=[];page.on('pageerror',e=>errors.push(e.message));
async function switchUser(id){
 await page.getByRole('button',{name:'Назад к документам'}).click();
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 await page.locator('.navigation.is-open .user-panel').click();
 await page.getByRole('combobox',{name:'Проверить от лица'}).selectOption(id);
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
}
async function downloadBytes(){
 const pending=page.waitForEvent('download');
 await page.getByRole('button',{name:'Скачать документ'}).click();
 const download=await pending;
 const stream=await download.createReadStream();
 const chunks=[];for await(const chunk of stream)chunks.push(chunk);
 return {name:download.suggestedFilename(),body:Buffer.concat(chunks).toString('utf8')};
}
try{
 await page.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('documents')));
 await page.goto('http://127.0.0.1:8082',{waitUntil:'domcontentloaded'});
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 await page.getByRole('button',{name:'Открыть команды и каналы'}).click();
 await page.locator('.navigation.is-open .user-panel').click();
 await page.getByRole('combobox',{name:'Проверить от лица'}).selectOption('2');
 await page.getByRole('heading',{name:'Все документы'}).waitFor();
 await page.getByRole('button',{name:'Загрузить документ',exact:true}).first().click();
 await page.locator('.modal input[type=file]').setInputFiles({name:'contract.txt',mimeType:'text/plain',buffer:Buffer.from('original contract')});
 await page.getByRole('button',{name:'Загрузить документ'}).last().click();
 await page.getByRole('dialog',{name:'contract'}).waitFor();
 await page.getByRole('button',{name:'На согласование'}).click();
 await page.getByText('На согласовании',{exact:true}).last().waitFor();
 await switchUser('1');
 await page.getByRole('button',{name:'contract',exact:false}).first().click();
 await page.getByRole('dialog',{name:'contract'}).waitFor();
 const original=await downloadBytes();
 if(original.name!=='contract.txt'||original.body!=='original contract')throw Error('original file download failed: '+JSON.stringify(original));
 const choose=page.waitForEvent('filechooser');
 await page.getByRole('button',{name:'Загрузить версию'}).click();
 const versionResponse=page.waitForResponse(r=>r.url().includes('/versions')&&r.request().method()==='POST');
 await (await choose).setFiles({name:'signed.txt',mimeType:'text/plain',buffer:Buffer.from('signed and stamped contract')});
 const response=await versionResponse;if(response.status()!==201)throw Error('version upload failed: '+await response.text());
 await page.waitForFunction(()=>document.querySelector('.preview-toolbar select')?.options.length===2);
 await page.getByRole('button',{name:'Согласовать'}).click();
 await page.getByText('Согласовано',{exact:true}).waitFor();
 const signed=await downloadBytes();
 if(signed.name!=='signed.txt'||signed.body!=='signed and stamped contract')throw Error('signed file download failed: '+JSON.stringify(signed));
 if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw Error('horizontal mobile overflow');
 if(errors.length)throw Error(errors.join('; '));
 const maxPage=await browser.newPage({viewport:{width:390,height:844}});
 await maxPage.route('https://st.max.ru/js/max-web-app.js',route=>route.abort());
 await maxPage.addInitScript(()=>{window.__calls=[];window.WebApp={initData:'mock',downloadFile:(url,name)=>{window.__calls.push({url,name,active:navigator.userActivation.isActive});return Promise.resolve()}}});
 await maxPage.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('documents')));
 await maxPage.goto('http://127.0.0.1:8082',{waitUntil:'domcontentloaded'});
 await maxPage.getByRole('listitem').first().getByRole('button').first().click();
 await maxPage.getByRole('button',{name:'Скачать документ'}).waitFor();
 await maxPage.waitForFunction(()=>!document.querySelector('button[aria-label="Скачать документ"]')?.disabled);
 await maxPage.getByRole('button',{name:'Скачать документ'}).click();
 const calls=await maxPage.evaluate(()=>window.__calls);
 if(calls.length!==1||!calls[0].active||!new URL(calls[0].url).pathname.startsWith('/api/download/'))throw Error('MAX Bridge download did not run in click handler: '+JSON.stringify(calls));
 const served=await maxPage.request.get(calls[0].url);
 if(served.status()!==200||!(await served.body()).length)throw Error('MAX Bridge URL did not serve a real file');
 await maxPage.close();
 console.log('Mobile: employee upload -> manager download -> stamped reupload -> approve -> signed download: PASS; MAX Bridge click + URL: PASS');
}finally{await browser.close()}