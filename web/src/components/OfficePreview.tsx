import {useState} from 'react';
import type {FilePreview} from '../types';
import './OfficePreview.css';

function Grid({rows}:{rows:string[][]}){
 const width=Math.min(30,Math.max(1,...rows.map(row=>row.length)));
 return <div className="office-grid-scroll"><table className="office-grid"><thead><tr><th aria-label="Номер строки">№</th>{Array.from({length:width},(_,i)=><th key={i}>{String.fromCharCode(65+i)}</th>)}</tr></thead><tbody>{rows.map((row,i)=><tr key={i}><th scope="row">{i+1}</th>{Array.from({length:width},(_,j)=><td key={j}>{row[j]||''}</td>)}</tr>)}</tbody></table></div>
}
export default function OfficePreview({data,title}:{data:FilePreview;title:string}){
 const [sheet,setSheet]=useState(0);
 if(data.kind==='document')return <article className="office-document"><div className="office-kicker">ДОКУМЕНТ · ПРЕДПРОСМОТР</div>{data.blocks?.map((block,i)=>block.kind==='heading'?<h3 key={i}>{block.text}</h3>:block.kind==='table'?<div className="office-doc-table-wrap" key={i}><table><tbody>{block.rows?.map((row,r)=><tr key={r}>{row.map((cell,c)=><td key={c}>{cell}</td>)}</tr>)}</tbody></table></div>:<p key={i}>{block.text}</p>)}{data.truncated&&<p className="office-caution">Показана часть документа. Скачайте оригинал для полной проверки.</p>}</article>;
 const selected=data.sheets?.[sheet];
 return <div className="office-workbook"><header><span>ТАБЛИЦА · ПРЕДПРОСМОТР</span><strong>{title}</strong></header><div className="office-sheet-tabs" role="tablist" aria-label="Листы таблицы">{data.sheets?.map((item,i)=><button key={i} role="tab" aria-selected={i===sheet} className={i===sheet?'active':''} onClick={()=>setSheet(i)}>{item.name||`Лист ${i+1}`}</button>)}</div>{selected&&<Grid rows={selected.rows}/>}<p className="office-caution">Предпросмотр показывает значения без оформления и пересчёта формул. Для точной проверки скачайте оригинал.{data.truncated?' Показана часть таблицы.':''}</p></div>
}