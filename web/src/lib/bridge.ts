import {useEffect} from 'react';
declare global { interface Window {WebApp?:{
 initData?:string; initDataUnsafe?:{start_param?:string};
 shareMaxContent?:(params:{text?:string;link?:string})=>Promise<unknown>;
 downloadFile?:(url:string,fileName:string)=>Promise<unknown>|void;
 BackButton?:{show:()=>void;hide:()=>void;onClick:(fn:()=>void)=>void;offClick:(fn:()=>void)=>void};
 enableClosingConfirmation?:()=>void;disableClosingConfirmation?:()=>void;
}}}
export function initData(){return window.WebApp?.initData||new URLSearchParams(location.hash.slice(1)).get('WebAppData')||''}
export function useMaxBack(enabled:boolean,onBack:()=>void){
 useEffect(()=>{const back=window.WebApp?.BackButton;if(!back)return;if(enabled){back.show();back.onClick(onBack)}else back.hide();return()=>{back.offClick(onBack);back.hide()}},[enabled,onBack]);
}
export function useViewport(){
 useEffect(()=>{const resize=()=>document.documentElement.style.setProperty('--app-height',`${window.visualViewport?.height||innerHeight}px`);resize();window.visualViewport?.addEventListener('resize',resize);window.addEventListener('resize',resize);return()=>{window.visualViewport?.removeEventListener('resize',resize);window.removeEventListener('resize',resize)}},[]);
}
export function stored<T>(key:string,fallback:T):T{try{return JSON.parse(localStorage.getItem(key)||'null')??fallback}catch{return fallback}}
export function remember(key:string,value:unknown){try{localStorage.setItem(key,JSON.stringify(value))}catch{/* Storage may be disabled in a WebView. */}}

export function extractInviteCode(input:string):string{
 const value=input.trim();
 const candidate=(()=>{try{const url=new URL(value);return url.searchParams.get('startapp')||url.searchParams.get('invite')||''}catch{return value}})();
 return /^(?:i_)?([a-f0-9]{24})$/i.exec(candidate)?.[1].toLowerCase()||'';
}
export function launchInviteCode():string{
 return extractInviteCode(window.WebApp?.initDataUnsafe?.start_param||new URLSearchParams(location.search).get('invite')||'');
}
