import {chromium} from '@playwright/test';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
import fs from 'node:fs/promises';
import path from 'node:path';
const base=process.env.DOCSPACE_E2E_URL||'http://127.0.0.1:8082';
const root=fileURLToPath(new URL('../',import.meta.url));
const browser=await chromium.launch({channel:'msedge',headless:true});
const api=async(route,options={})=>{const response=await fetch(base+'/api/'+route,{...options,headers:{'X-Demo-User':'4','Content-Type':'application/json',...options.headers}});if(!response.ok)throw Error(await response.text());return response.json()};
let key;
try{
 key=await api('teams/demo-fin/integrations',{method:'POST',body:JSON.stringify({name:'End-to-end accounting client',provider:'1c'})});
 const dir=await fs.mkdtemp(path.join(root,'tmp','report-client-'));const filename=path.join(dir,'cashflow.csv');const bytes=Buffer.from('\ufeffDepartment;Plan;Actual\nSales;125000;120000\n');await fs.writeFile(filename,bytes);
 function client(args){const r=spawnSync('python',[path.join(root,'scripts','send-report.py'),'--base-url',base,...args],{encoding:'utf8',env:{...process.env,PYTHONIOENCODING:'utf-8',DOCSPACE_INTEGRATION_TOKEN:key.token}});if(r.status!==0)throw Error(r.stderr);return JSON.parse(r.stdout)}
 const channels=client(['--channels']);if(!channels.length||channels.some(c=>c.teamId!=='demo-fin'))throw Error('key channels leaked');
 const title='API Cashflow '+Date.now();const external='cashflow-'+Date.now();const args=['--file',filename,'--channel','demo-fin-cashflow','--title',title,'--external-id',external,'--review'];
 const imported=client(args);if(imported.status!=='review'||imported.importMethod!=='api')throw Error('bad real API import');if(client(args).id!==imported.id)throw Error('duplicate import');
 const page=await browser.newPage({viewport:{width:390,height:844},acceptDownloads:true});const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.addInitScript(()=>{localStorage.setItem('docspace-user',JSON.stringify('4'));localStorage.setItem('docspace-section',JSON.stringify('systems'))});
 await page.goto(base);await page.locator('.provider-card.provider-1c').click();await page.getByRole('combobox',{name:'Команда отчётов'}).selectOption('demo-fin');await page.getByRole('textbox',{name:'Поиск отчётов'}).fill(title);await page.locator('.report-item').filter({hasText:title}).click();
 await page.locator('.office-grid').getByText('125000',{exact:true}).waitFor();const pending=page.waitForEvent('download');await page.getByRole('button',{name:'Скачать документ'}).click();const download=await pending;const stream=await download.createReadStream();const chunks=[];for await(const chunk of stream)chunks.push(chunk);if(!Buffer.concat(chunks).equals(bytes))throw Error('download bytes changed');
 await page.getByRole('button',{name:'Согласовать',exact:true}).click();await page.locator('.detail-shell .approved-note').waitFor();if(client(['--status',external]).status!=='approved')throw Error('client did not receive approval');
 await page.getByRole('button',{name:'Назад к документам'}).click();await page.getByRole('button',{name:'Загрузить отчёт',exact:true}).click();await page.locator('.import-layout input[type=file]').setInputFiles({name:'manual-crm.csv',mimeType:'text/csv',buffer:bytes});await page.locator('.import-layout form').getByRole('button',{name:'Загрузить отчёт',exact:true}).click();await page.getByRole('dialog',{name:'manual-crm'}).waitFor();
 const state=await api('state');const manual=state.documents.find(d=>d.title==='manual-crm');if(!manual||manual.sourceSystem!=='1c'||manual.importMethod!=='file'||manual.status!=='review')throw Error('manual report metadata lost');
 if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw Error('mobile overflow');if(errors.length)throw Error(errors.join('; '));
 const script=await fetch(base+'/integration/send-report.py');if(!script.ok||!(await script.text()).includes('DOCSPACE_INTEGRATION_TOKEN'))throw Error('sender download unavailable');
 await api('teams/demo-fin/integrations/'+key.id,{method:'DELETE'});const rejected=spawnSync('python',[path.join(root,'scripts','send-report.py'),'--base-url',base,'--status',external],{encoding:'utf8',env:{...process.env,DOCSPACE_INTEGRATION_TOKEN:key.token}});if(rejected.status===0||!rejected.stderr.includes('401'))throw Error('revocation not enforced');key=null;
 console.log('Real HTTP sender -> idempotent CSV import -> mobile preview/download -> approval -> client status; manual import; revoked key PASS');
}finally{if(key)await api('teams/demo-fin/integrations/'+key.id,{method:'DELETE'}).catch(()=>{});await browser.close()}
