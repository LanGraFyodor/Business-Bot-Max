import React,{lazy,Suspense} from 'react';
import {createRoot} from 'react-dom/client';
import {initData} from './lib/bridge';
const Guide=lazy(()=>import('./guide/GuideApp'));
const Documents=lazy(()=>import('./DocspaceEntry'));
const Launcher=lazy(()=>import('./guide/Launcher'));
const query=new URLSearchParams(location.search);
const hash=new URLSearchParams(location.hash.slice(1));
const start=window.WebApp?.initDataUnsafe?.start_param||new URLSearchParams(initData()).get('start_param')||query.get('WebAppStartParam')||hash.get('WebAppStartParam')||query.get('startapp')||'';
const path=location.pathname.replace(/\/$/,'');
const chosen=path==='/guide'?'guide':path==='/documents'?'docspace':start==='guide'?'guide':start==='docspace'||/^i_[a-f0-9]{24}$/.test(start)?'docspace':path==='/apps'?'apps':'docspace';
const Screen=chosen==='guide'?Guide:chosen==='apps'?Launcher:Documents;
createRoot(document.getElementById('root')!).render(<React.StrictMode><Suspense fallback={<p style={{fontFamily:'system-ui',padding:24}}>Загрузка приложения…</p>}><Screen/></Suspense></React.StrictMode>);

