// konbu Portal
(function(){
'use strict';
const API='/api/v1',H={'Content-Type':'application/json','X-Forwarded-User':'test@example.com'};
async function api(m,p,b){const o={method:m,headers:H};if(b)o.body=JSON.stringify(b);const r=await fetch(API+p,o);if(r.status===204)return null;const d=await r.json();if(!r.ok)throw new Error(d.error?.message||'error');return d}
function el(t,a,...c){const e=document.createElement(t);if(a)Object.entries(a).forEach(([k,v])=>{if(k.startsWith('on'))e.addEventListener(k.slice(2).toLowerCase(),v);else if(k==='cls')e.className=v;else e.setAttribute(k,v)});c.flat().forEach(x=>{if(x!=null)e.appendChild(typeof x==='string'?document.createTextNode(x):x)});return e}
function $(i){return document.getElementById(i)}

// === Date helpers ===
function rel(iso){if(!iso)return'';const d=new Date(iso),n=new Date(),x=n-d;if(x<6e4)return'now';if(x<36e5)return Math.floor(x/6e4)+'m';if(x<864e5)return Math.floor(x/36e5)+'h';if(x<6048e5)return Math.floor(x/864e5)+'d';return d.toLocaleDateString('ja-JP',{month:'short',day:'numeric'})}
function hm(iso){return iso?new Date(iso).toLocaleTimeString('ja-JP',{hour:'2-digit',minute:'2-digit'}):''}
function md(iso){return iso?new Date(iso).toLocaleDateString('ja-JP',{month:'short',day:'numeric'}):''}

// Header date
$('header-date').textContent=new Date().toLocaleDateString('ja-JP',{year:'numeric',month:'long',day:'numeric',weekday:'short'});
$('header-user').textContent='U';

// === THEME SWITCHER ===
$('theme-switcher').addEventListener('click',e=>{
  const b=e.target.closest('.theme-btn');if(!b)return;
  document.body.className=b.dataset.theme;
  document.querySelectorAll('.theme-btn').forEach(x=>x.classList.toggle('active',x===b));
  localStorage.setItem('konbu-theme',b.dataset.theme);
});
const saved=localStorage.getItem('konbu-theme');
if(saved){document.body.className=saved;document.querySelectorAll('.theme-btn').forEach(x=>x.classList.toggle('active',x.dataset.theme===saved))}

// === MINI CALENDAR ===
let cY,cM,evDates=new Set();
function initCal(){const n=new Date();cY=n.getFullYear();cM=n.getMonth();drawCal()}
function drawCal(){
  const c=$('mini-cal'),dw=['Mo','Tu','We','Th','Fr','Sa','Su'];
  const f=new Date(cY,cM,1),L=new Date(cY,cM+1,0),s=(f.getDay()+6)%7,td=new Date();
  c.innerHTML='';
  c.appendChild(el('div',{cls:'cal-hdr'},
    el('span',{cls:'cal-hdr-t'},`${cY}/${String(cM+1).padStart(2,'0')}`),
    el('div',{cls:'cal-nav'},
      el('button',{onClick:()=>{cM--;if(cM<0){cM=11;cY--}drawCal()}},'\u25C0'),
      el('button',{onClick:()=>{const n=new Date();cY=n.getFullYear();cM=n.getMonth();drawCal()}},'\u25CF'),
      el('button',{onClick:()=>{cM++;if(cM>11){cM=0;cY++}drawCal()}},'\u25B6'),
    )
  ));
  const g=el('div',{cls:'cal-g'});
  dw.forEach(d=>g.appendChild(el('div',{cls:'cal-dw'},d)));
  const pL=new Date(cY,cM,0).getDate();
  for(let i=s-1;i>=0;i--)g.appendChild(el('div',{cls:'cal-d om'},String(pL-i)));
  for(let d=1;d<=L.getDate();d++){
    const isT=d===td.getDate()&&cM===td.getMonth()&&cY===td.getFullYear();
    const dk=`${cY}-${String(cM+1).padStart(2,'0')}-${String(d).padStart(2,'0')}`;
    const hasE=evDates.has(dk);
    g.appendChild(el('div',{cls:'cal-d'+(isT?' today':'')+(hasE?' ev':'')},String(d)));
  }
  const tot=s+L.getDate(),rem=(7-tot%7)%7;
  for(let d=1;d<=rem;d++)g.appendChild(el('div',{cls:'cal-d om'},String(d)));
  c.appendChild(g);
}

// === EVENTS ===
async function loadEv(){
  const r=await api('GET','/events?limit=20');const items=r.data||[];
  evDates.clear();
  items.forEach(e=>{const d=new Date(e.start_at);evDates.add(`${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`)});
  drawCal();
  const c=$('event-list');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'ev-empty'},'No events'));return}
  const list=el('div',{cls:'ev-list'});
  items.forEach(ev=>list.appendChild(el('div',{cls:'ev-item',onClick:()=>showEvDetail(ev.id)},
    el('span',{cls:'ev-time'},ev.all_day?'All day':hm(ev.start_at)),
    el('div',null,el('div',{cls:'ev-title'},ev.title),el('div',{cls:'ev-date'},md(ev.start_at)))
  )));
  c.appendChild(list);
}

// === TODOS ===
let tFilter='open';
async function loadTd(){
  const r=await api('GET','/todos?limit=100');let items=r.data||[];
  $('todo-count').textContent=items.filter(t=>t.status==='open').length;
  if(tFilter==='open')items=items.filter(t=>t.status==='open');
  else if(tFilter==='done')items=items.filter(t=>t.status==='done');
  const c=$('todo-list');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'empty'},tFilter==='open'?'All done!':'No items'));return}
  const list=el('div',{cls:'td-list'});
  items.forEach(t=>{
    const done=t.status==='done',over=t.due_date&&!done&&new Date(t.due_date+'T23:59:59')<new Date();
    list.appendChild(el('div',{cls:'td-item',onClick:()=>showTdDetail(t.id)},
      el('span',{cls:'td-ck'+(done?' done':''),onClick:e=>{e.stopPropagation();api('PATCH','/todos/'+t.id+(done?'/reopen':'/done')).then(loadTd)}}),
      el('span',{cls:'td-txt'+(done?' done':'')},t.title),
      t.due_date?el('span',{cls:'td-due'+(over?' over':'')},t.due_date):null,
      ...((t.tags||[]).map(x=>el('span',{cls:'tag'},x.name))),
    ));
  });
  c.appendChild(list);
}
$('todo-chips').addEventListener('click',e=>{const b=e.target.closest('.chip');if(!b||!b.dataset.f)return;tFilter=b.dataset.f;document.querySelectorAll('#todo-chips .chip').forEach(x=>x.classList.toggle('on',x.dataset.f===tFilter));loadTd()});

// === MEMOS ===
async function loadMm(){
  const r=await api('GET','/memos?limit=20');const items=r.data||[];
  $('memo-count').textContent=r.total||items.length;
  const c=$('memo-list');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'empty'},'No memos'));return}
  const list=el('div',{cls:'ml'});
  items.forEach(m=>{
    const tc=m.type==='table'?'type-tbl':'type-md';
    list.appendChild(el('div',{cls:'mc',onClick:()=>showMmDetail(m.id)},
      el('div',{cls:'mc-t'},el('span',{cls:'type-b '+tc},m.type==='table'?'TBL':'MD'),m.title||'Untitled'),
      m.content?el('div',{cls:'mc-p'},m.content):null,
      el('div',{cls:'mc-m'},...(m.tags||[]).map(t=>el('span',{cls:'tag'},t.name)),el('span',null,rel(m.updated_at))),
    ));
  });
  c.appendChild(list);
}

// === TOOLS ===
async function loadTl(){
  const r=await api('GET','/tools');const items=r.data||[];
  const c=$('tool-grid');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'empty'},'No tools'));return}
  items.forEach((t,i)=>c.appendChild(el('a',{cls:'tc',href:t.url,target:'_blank',rel:'noopener'},
    el('div',{cls:'tc-i tc-i-'+i%6},(t.icon||t.name||'?')[0].toUpperCase()),
    el('div',{cls:'tc-n'},t.name)
  )));
}

// === TAGS ===
async function loadTg(){
  const r=await api('GET','/tags');const items=r.data||[];
  const c=$('tag-list');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'empty'},'No tags'));return}
  const cloud=el('div',{cls:'tcloud'});
  items.forEach(t=>cloud.appendChild(el('span',{cls:'tag'},t.name)));
  c.appendChild(cloud);
}

// === QUICK INPUT ===
let qType='memo';
document.querySelector('.w-quick .chips')?.addEventListener('click',e=>{
  const b=e.target.closest('.chip');if(!b||!b.dataset.q)return;
  qType=b.dataset.q;
  document.querySelectorAll('.w-quick .chips .chip').forEach(x=>x.classList.toggle('on',x.dataset.q===qType));
  $('quick-in').placeholder=qType==='memo'?'Type and Enter to add memo...':'Type and Enter to add todo...';
});
$('quick-in').addEventListener('keydown',async e=>{
  if(e.key!=='Enter')return;const v=e.target.value.trim();if(!v)return;e.target.value='';
  if(qType==='memo'){await api('POST','/memos',{title:v,type:'markdown',content:v});loadMm()}
  else{await api('POST','/todos',{title:v,description:''});loadTd()}
});

// === CREATE ACTIONS ===
document.addEventListener('click',e=>{const b=e.target.closest('[data-action]');if(!b)return;({
  'new-memo':newMm,'new-todo':newTd,'new-event':newEv,'new-tool':newTl
})[b.dataset.action]?.()});

// === FORM HELPERS ===
function fg(l,t,id,ph,v){const g=el('div',{cls:'fg'},el('label',null,l));if(t==='textarea'){const a=el('textarea',{id,placeholder:ph||''});if(v)a.textContent=v;g.appendChild(a)}else{const i=el('input',{type:t,id,placeholder:ph||''});if(v!==undefined)i.value=v;g.appendChild(i)}return g}
function fa(fn){return el('div',{cls:'fa'},el('button',{cls:'btn btn-g',onClick:closeM},'Cancel'),el('button',{cls:'btn btn-p',onClick:fn},'Create'))}
function pt(s){return(s||'').split(',').map(x=>x.trim()).filter(Boolean)}

// === MEMO MODALS ===
function newMm(){showM('New Memo',el('div',null,fg('Title','text','nm-t','Title'),fg('Content','textarea','nm-c','Write...'),fg('Tags','text','nm-tg','tag1, tag2'),fa(async()=>{await api('POST','/memos',{title:$('nm-t').value,type:'markdown',content:$('nm-c').value,tags:pt($('nm-tg').value)});closeM();loadMm()})));$('nm-t').focus()}
async function showMmDetail(id){
  const r=await api('GET','/memos/'+id),m=r.data;
  showM('Edit Memo',el('div',null,fg('Title','text','em-t','',m.title||''),fg('Content','textarea','em-c','',m.content||''),fg('Tags','text','em-tg','',(m.tags||[]).map(t=>t.name).join(', ')),
    el('div',{cls:'fa'},
      el('button',{cls:'btn btn-d',onClick:()=>{if(confirm('Delete?'))api('DELETE','/memos/'+id).then(()=>{closeM();loadMm()})}},'Delete'),
      el('div',{style:'flex:1'}),
      el('button',{cls:'btn btn-g',onClick:closeM},'Cancel'),
      el('button',{cls:'btn btn-p',onClick:async()=>{await api('PUT','/memos/'+id,{title:$('em-t').value,content:$('em-c').value,tags:pt($('em-tg').value)});closeM();loadMm()}},'Save'),
    )
  ))
}

// === TODO MODALS ===
function newTd(){showM('New ToDo',el('div',null,fg('Title','text','nt-t','What needs to be done?'),fg('Description','textarea','nt-d',''),el('div',{cls:'fr'},fg('Due','date','nt-du'),fg('Tags','text','nt-tg','tag1, tag2')),fa(async()=>{await api('POST','/todos',{title:$('nt-t').value,description:$('nt-d').value,due_date:$('nt-du').value||null,tags:pt($('nt-tg').value)});closeM();loadTd()})));$('nt-t').focus()}
async function showTdDetail(id){
  const r=await api('GET','/todos/'+id),t=r.data;
  showM('Edit ToDo',el('div',null,fg('Title','text','et-t','',t.title),fg('Description','textarea','et-d','',t.description||''),el('div',{cls:'fr'},fg('Due','date','et-du','',t.due_date||''),fg('Tags','text','et-tg','',(t.tags||[]).map(x=>x.name).join(', '))),
    el('div',{cls:'fa'},
      el('button',{cls:'btn btn-d',onClick:()=>{if(confirm('Delete?'))api('DELETE','/todos/'+id).then(()=>{closeM();loadTd()})}},'Delete'),
      el('div',{style:'flex:1'}),
      el('button',{cls:'btn btn-g',onClick:closeM},'Cancel'),
      el('button',{cls:'btn btn-p',onClick:async()=>{await api('PUT','/todos/'+id,{title:$('et-t').value,description:$('et-d').value,status:t.status,due_date:$('et-du').value||null,tags:pt($('et-tg').value)});closeM();loadTd()}},'Save'),
    )
  ))
}

// === EVENT MODALS ===
function newEv(){showM('New Event',el('div',null,fg('Title','text','ne-t','Event'),fg('Description','textarea','ne-d',''),el('div',{cls:'fr'},fg('Start','datetime-local','ne-s'),fg('End','datetime-local','ne-e')),fg('Tags','text','ne-tg','tag1, tag2'),fa(async()=>{await api('POST','/events',{title:$('ne-t').value,description:$('ne-d').value,start_at:new Date($('ne-s').value).toISOString(),end_at:$('ne-e').value?new Date($('ne-e').value).toISOString():null,all_day:false,tags:pt($('ne-tg').value)});closeM();loadEv()})));$('ne-t').focus()}
async function showEvDetail(id){
  const r=await api('GET','/events/'+id),ev=r.data;
  const sL=ev.start_at?new Date(ev.start_at).toISOString().slice(0,16):'',eL=ev.end_at?new Date(ev.end_at).toISOString().slice(0,16):'';
  showM('Edit Event',el('div',null,fg('Title','text','ee-t','',ev.title),fg('Description','textarea','ee-d','',ev.description||''),el('div',{cls:'fr'},fg('Start','datetime-local','ee-s','',sL),fg('End','datetime-local','ee-e','',eL)),fg('Tags','text','ee-tg','',(ev.tags||[]).map(t=>t.name).join(', ')),
    el('div',{cls:'fa'},
      el('button',{cls:'btn btn-d',onClick:()=>{if(confirm('Delete?'))api('DELETE','/events/'+id).then(()=>{closeM();loadEv()})}},'Delete'),
      el('div',{style:'flex:1'}),
      el('button',{cls:'btn btn-g',onClick:closeM},'Cancel'),
      el('button',{cls:'btn btn-p',onClick:async()=>{await api('PUT','/events/'+id,{title:$('ee-t').value,description:$('ee-d').value,start_at:new Date($('ee-s').value).toISOString(),end_at:$('ee-e').value?new Date($('ee-e').value).toISOString():null,all_day:ev.all_day,tags:pt($('ee-tg').value)});closeM();loadEv()}},'Save'),
    )
  ))
}

// === TOOL MODAL ===
function newTl(){showM('New Tool',el('div',null,fg('Name','text','ntl-n','Name'),fg('URL','url','ntl-u','https://...'),fg('Icon','text','ntl-i','Emoji or letter'),fa(async()=>{await api('POST','/tools',{name:$('ntl-n').value,url:$('ntl-u').value,icon:$('ntl-i').value});closeM();loadTl()})));$('ntl-n').focus()}

// === MODAL ===
function showM(t,b){$('modal-title').textContent=t;const mb=$('modal-body');mb.innerHTML='';mb.appendChild(b);$('modal-bg').classList.remove('hidden')}
function closeM(){$('modal-bg').classList.add('hidden')}
$('modal-x').addEventListener('click',closeM);
$('modal-bg').addEventListener('click',e=>{if(e.target===$('modal-bg'))closeM()});
document.addEventListener('keydown',e=>{if(e.key==='Escape')closeM();if(e.key==='/'&&!['INPUT','TEXTAREA'].includes(document.activeElement.tagName)){e.preventDefault();$('global-search').focus()}});

// === INIT ===
Promise.all([loadEv(),loadTd(),loadMm(),loadTl(),loadTg()]).then(initCal);
})();
