import {chromium} from '@playwright/test';
const browser=await chromium.launch({channel:'msedge',headless:true});
const teamName='Роли QA '+Date.now();
try{
 const boss=await browser.newPage({viewport:{width:1280,height:900}});
 await boss.addInitScript(()=>localStorage.setItem('docspace-section',JSON.stringify('teams')));
 await boss.goto('http://127.0.0.1:8082');
 await boss.locator('.organisation-page').getByRole('button',{name:'Добавить команду',exact:true}).click();
 await boss.getByPlaceholder('Например, Команда запуска').fill(teamName);await boss.locator('form').getByRole('button',{name:'Создать команду',exact:true}).click();
 await boss.getByRole('button',{name:'Приглашения и доступ',exact:true}).click();
 await boss.getByRole('combobox',{name:'Роль приглашения'}).selectOption('manager');await boss.getByRole('textbox',{name:'MAX ID адресата'}).fill('2');await boss.getByRole('button',{name:'Создать приглашение',exact:true}).click();
 const managerLink=await boss.getByRole('textbox',{name:'Ссылка приглашения'}).inputValue();
 const anna=await browser.newPage();await anna.addInitScript(()=>localStorage.setItem('docspace-user',JSON.stringify('2')));await anna.goto(managerLink);await anna.getByRole('button',{name:'Присоединиться',exact:true}).click();
 await anna.locator('.organisation-team-heading h3').filter({hasText:teamName}).waitFor();await anna.getByRole('button',{name:'Приглашения и доступ',exact:true}).waitFor();
 const employee=await browser.newPage();await employee.addInitScript(()=>localStorage.setItem('docspace-user',JSON.stringify('3')));await employee.goto(managerLink);await employee.getByRole('button',{name:'Присоединиться',exact:true}).click();await employee.getByText('Приглашение недействительно').waitFor();
 await boss.getByRole('combobox',{name:'Роль приглашения'}).selectOption('employee');await boss.getByRole('textbox',{name:'MAX ID адресата'}).fill('3');await boss.getByRole('button',{name:'Создать приглашение',exact:true}).click();
 await boss.waitForFunction(old=>document.querySelector('input[aria-label="Ссылка приглашения"]')?.value!==old,managerLink);
 const employeeLink=await boss.getByRole('textbox',{name:'Ссылка приглашения'}).inputValue();await employee.goto(employeeLink);await employee.getByRole('button',{name:'Присоединиться',exact:true}).click();await employee.locator('.organisation-team-heading h3').filter({hasText:teamName}).waitFor();
 if(await employee.getByRole('button',{name:'Приглашения и доступ',exact:true}).count())throw Error('employee has invitation controls');
 await boss.locator('.organisation-detail .module-tabs').getByRole('button',{name:/Участники/}).click();await boss.getByRole('button',{name:'Исключить: Михаил Волков',exact:true}).click();await boss.getByRole('button',{name:'Исключить',exact:true}).click();
 await employee.locator('.organisation-team-heading h3').filter({hasText:teamName}).waitFor({state:'detached'});
 await boss.setViewportSize({width:390,height:844});await boss.waitForTimeout(250);if(await boss.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw Error('mobile overflow');
 console.log('Organisation: MAX-bound invitations, distinct roles, one-use links, removal, mobile PASS');
}finally{await browser.close()}
