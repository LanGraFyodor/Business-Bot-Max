import {CheckCheck,ChevronDown,FileCheck2,Files,Hash,Network,Plus,Settings2,Workflow,X} from 'lucide-react';
import type {Team,Channel,User,Doc} from '../types';
import {initials} from '../lib/format';
import {providers} from '../lib/providers';
export type Section='documents'|'systems'|'teams';
type Props={teams:Team[];team:Team|undefined;channels:Channel[];docs:Doc[];view:string;user:User;open:boolean;demo:boolean;section:Section;source:string;onSection:(s:Section)=>void;onSource:(s:string)=>void;selectTeam:(id:string)=>void;navigate:(id:string)=>void;close:()=>void;newTeam:()=>void;newChannel:()=>void;members:()=>void;profile:()=>void};
export default function Navigation(p:Props){
 const manager=p.team?.members.find(m=>m.id===p.user.id)?.role==='manager';
 return <><button className={`drawer-scrim ${p.open?'visible':''}`} aria-label="Закрыть список каналов" onClick={p.close} tabIndex={p.open?0:-1}/><aside className={`navigation product-navigation ${p.open?'is-open':''}`}>
 <div className="product-brand"><span className="product-logo"><FileCheck2 size={25}/></span><div><strong>docspace<span>.</span></strong><small>Рабочее пространство в MAX</small></div></div>
 <nav className="product-sections" aria-label="Разделы пространства">
 <button className={p.section==='documents'?'active':''} onClick={()=>p.onSection('documents')}><Files size={20}/><span>Документы<small>Файлы и обсуждения</small></span></button>
 <button className={p.section==='systems'?'active':''} onClick={()=>p.onSection('systems')}><Workflow size={20}/><span>Отчёты из систем<small>1С, CRM и другие источники</small></span><span className="section-dot"/></button>
 <button className={p.section==='teams'?'active':''} onClick={()=>p.onSection('teams')}><Network size={20}/><span>Команды и роли<small>Структура и доступ</small></span></button>
 </nav>
 <div className="product-nav-scroll">
 {p.section==='systems'?<><div className="sidebar-label">ИСТОЧНИКИ ОТЧЁТОВ</div><nav className="source-navigation">{providers.map(s=><button key={s.id} className={p.source===s.id?'active':''} onClick={()=>p.onSource(s.id)}><span className={`source-mark source-${s.id}`}>{s.short}</span><span>{s.name}</span></button>)}</nav><p className="navigation-note">Выгрузки из рабочих систем проходят обсуждение и согласование вместе с командой.</p></>:<>
 <div className="sidebar-label">РАБОЧАЯ КОМАНДА<button aria-label="Добавить команду" title="Создать или присоединиться" onClick={p.newTeam}><Plus size={17}/></button></div>
 <label className="team-switch"><span className={`team-icon ${p.team?.color||'violet'}`}>{p.team?initials(p.team.name):'+'}</span><select aria-label="Рабочая команда" value={p.team?.id||''} onChange={e=>p.selectTeam(e.target.value)}>{!p.team&&<option value="">Нет команды</option>}{p.teams.map(t=><option key={t.id} value={t.id}>{t.name}</option>)}</select><ChevronDown size={14}/></label>
 {p.section==='documents'?<><nav className="quick-nav"><button className={`nav-item ${p.view==='all'?'active':''}`} onClick={()=>p.navigate('all')}><Files size={18}/>Все документы</button><button className={`nav-item ${p.view==='review'?'active':''}`} onClick={()=>p.navigate('review')}><CheckCheck size={18}/>На согласовании<em>{p.docs.filter(d=>d.status==='review').length}</em></button></nav><div className="sidebar-label">КАНАЛЫ{manager&&<button aria-label="Создать канал" onClick={p.newChannel}><Plus size={17}/></button>}</div><nav>{p.channels.map(c=><button key={c.id} className={`nav-item ${p.view===c.id?'active':''}`} onClick={()=>p.navigate(c.id)}><Hash size={18}/><span>{c.name}</span><small>{p.docs.filter(d=>d.channelId===c.id).length||''}</small></button>)}</nav></>:<p className="navigation-note">Один аккаунт — разные роли. Управляйте своей командой и передавайте отчёты руководству.</p>}
 </>}
 </div>
 <button className="user-panel" aria-label="Настройки профиля" onClick={p.profile}><span className="avatar">{initials(p.user.name)}</span><span><strong>{p.user.name}</strong><small>{p.demo?'Демонстрационный аккаунт':'Аккаунт MAX'}</small></span><Settings2 size={18}/></button>
 <button className="drawer-close icon-button" aria-label="Закрыть каналы" onClick={p.close}><X size={21}/></button>
 </aside></>;
}
