import {useCallback,useEffect,useRef,useState} from 'react';
import type {Data} from '../types';
import {initData,remember,stored} from '../lib/bridge';
export function useWorkspace(){
 const [data,setData]=useState<Data|null>(null),[error,setError]=useState(''),[busy,setBusy]=useState(false),[toast,setToast]=useState(''),[demoUser,setDemoUser]=useState(()=>String(stored('docspace-user','1')));
 const sequence=useRef(0);
 const headers=useCallback(()=>({'X-Demo-User':demoUser,'X-Max-Init-Data':initData()}),[demoUser]);
 const api=useCallback(async(path:string,options:RequestInit={})=>{const res=await fetch('/api/'+path,{...options,headers:{...headers(),...options.headers}});if(!res.ok){const body=await res.json().catch(()=>({error:'Сервер недоступен'}));throw Error(body.error||'Не удалось выполнить действие')}return res.json()},[headers]);
 const refresh=useCallback(async()=>{const seq=++sequence.current;try{const next=await api('state');if(seq===sequence.current){setData(next);setError('')}}catch(e){if(seq===sequence.current)setError((e as Error).message)}},[api]);
 useEffect(()=>{setData(null);refresh();const timer=setInterval(refresh,6000);return()=>{sequence.current++;clearInterval(timer)}},[refresh]);
 useEffect(()=>{if(!toast)return;const timer=setTimeout(()=>setToast(''),4200);return()=>clearTimeout(timer)},[toast]);
 async function mutate<T>(operation:()=>Promise<T>,message:string):Promise<T|undefined>{setBusy(true);try{const result=await operation();await refresh();setToast(message);return result}catch(e){setToast((e as Error).message);return undefined}finally{setBusy(false)}}
 function switchUser(id:string){sequence.current++;setData(null);setDemoUser(id);remember('docspace-user',id)}
 return {data,error,busy,toast,setToast,headers,api,refresh,mutate,demoUser,switchUser};
}
