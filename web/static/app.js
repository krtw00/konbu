// konbu
(function(){
'use strict';
const API='/api/v1';
async function api(m,p,b){const h={'Content-Type':'application/json'};const o={method:m,headers:h};if(b)o.body=JSON.stringify(b);const r=await fetch(API+p,o);if(r.status===204)return null;const d=await r.json();if(!r.ok)throw new Error(d.error?.message||'error');return d}
function el(t,a,...c){const e=document.createElement(t);if(a)Object.entries(a).forEach(([k,v])=>{if(k.startsWith('on'))e.addEventListener(k.slice(2).toLowerCase(),v);else if(k==='cls')e.className=v;else e.setAttribute(k,v)});c.flat().forEach(x=>{if(x!=null)e.appendChild(typeof x==='string'?document.createTextNode(x):x)});return e}
function $(i){return document.getElementById(i)}

// Date helpers
function rel(iso){if(!iso)return'';const d=new Date(iso),n=new Date(),x=n-d;if(x<6e4)return'now';if(x<36e5)return Math.floor(x/6e4)+'m ago';if(x<864e5)return Math.floor(x/36e5)+'h ago';if(x<6048e5)return Math.floor(x/864e5)+'d ago';return d.toLocaleDateString('ja-JP',{month:'short',day:'numeric'})}
function hm(iso){return iso?new Date(iso).toLocaleTimeString('ja-JP',{hour:'2-digit',minute:'2-digit'}):''}
function md(iso){return iso?new Date(iso).toLocaleDateString('ja-JP',{month:'short',day:'numeric'}):''}

// === PAGE NAVIGATION ===
let currentPage='home';
function navigateTo(page){
  // Auto-save memo editor if leaving
  if(currentPage==='memo-edit'&&page!=='memo-edit'&&memoEditId){clearTimeout(memoSaveTimer);saveMemo();memoEditId=null}
  currentPage=page;
  document.querySelectorAll('.page').forEach(p=>p.classList.toggle('active',p.id==='page-'+page));
  document.querySelectorAll('.sb-item[data-page]').forEach(b=>b.classList.toggle('active',b.dataset.page===page));
  document.querySelectorAll('.bn-item[data-page]').forEach(b=>b.classList.toggle('active',b.dataset.page===page));
  loaders[page]?.();
}
$('sidebar').addEventListener('click',e=>{const b=e.target.closest('[data-page]');if(b)navigateTo(b.dataset.page)});
$('bottom-nav').addEventListener('click',e=>{const b=e.target.closest('[data-page]');if(b)navigateTo(b.dataset.page)});

// === HOME ===
$('home-date').textContent=new Date().toLocaleDateString('ja-JP',{year:'numeric',month:'long',day:'numeric',weekday:'short'});

async function loadHome(){
  // Events today
  const evR=await api('GET','/events?limit=10');
  const evs=(evR.data||[]).filter(e=>{
    const d=new Date(e.start_at).toDateString(),t=new Date().toDateString();return d===t;
  });
  const ec=$('home-events');ec.innerHTML='';
  if(!evs.length)ec.appendChild(el('div',{cls:'empty'},'No events today'));
  else evs.forEach(e=>ec.appendChild(el('div',{cls:'home-ev'},
    el('span',{cls:'home-ev-time'},e.all_day?'All day':hm(e.start_at)),
    el('span',{cls:'home-ev-title'},e.title)
  )));

  // Todos open
  const tdR=await api('GET','/todos?limit=100');
  const tds=(tdR.data||[]).filter(t=>t.status==='open');
  $('home-todo-count').textContent=tds.length;
  const tc=$('home-todos');tc.innerHTML='';
  if(!tds.length)tc.appendChild(el('div',{cls:'empty'},'All done!'));
  else tds.slice(0,8).forEach(t=>{
    const df=dueFmt(t.due_date);
    tc.appendChild(el('div',{cls:'home-td'},
      el('span',{cls:'home-td-ck',onClick:e=>{e.stopPropagation();api('PATCH','/todos/'+t.id+'/done').then(loadHome)}}),
      el('span',{cls:'home-td-txt'},t.title),
      df?el('span',{cls:'home-td-due'+(df.cls?' '+df.cls:'')},df.text):null
    ));
  });

  // Recent memos
  const mmR=await api('GET','/memos?limit=6');
  const mms=mmR.data||[];
  const mc=$('home-memos');mc.innerHTML='';
  if(!mms.length)mc.appendChild(el('div',{cls:'empty'},'No memos yet'));
  else{const wrap=el('div',{cls:'home-mm'});mms.forEach(m=>wrap.appendChild(
    el('div',{cls:'home-mc',onClick:()=>{navigateTo('memos');showMmDetail(m.id)}},
      el('div',{cls:'home-mc-t'},m.title||'Untitled'),
      m.content?el('div',{cls:'home-mc-p'},m.content):null
    )
  ));mc.appendChild(wrap)}
}

// === MEMOS ===
let memoCtx=null;
function closeMemoCtx(){if(memoCtx){memoCtx.remove();memoCtx=null}}
function showMemoCtx(e,m){
  e.stopPropagation();
  closeMemoCtx();
  const btn=e.currentTarget;
  const rect=btn.getBoundingClientRect();
  const tags=(m.tags||[]).map(t=>t.name);

  // Tag toggle submenu
  function tagItems(){
    return allTagsCache.map(name=>{
      const has=tags.includes(name);
      return el('button',{onClick:async ev=>{
        ev.stopPropagation();
        let newTags;
        if(has)newTags=tags.filter(n=>n!==name);
        else{tags.push(name);newTags=[...tags]}
        await api('PUT','/memos/'+m.id,{title:m.title,content:m.content,tags:newTags});
        closeMemoCtx();loadMemos();
      }},(has?'✓ ':'')+name);
    });
  }

  const menu=el('div',{cls:'memo-ctx'},
    el('button',{onClick:()=>{closeMemoCtx();showMmDetail(m.id)}},'Edit'),
    el('div',{cls:'memo-ctx-sep'}),
    ...tagItems(),
    allTagsCache.length?el('div',{cls:'memo-ctx-sep'}):null,
    el('button',{cls:'danger',onClick:async()=>{if(confirm('Delete "'+( m.title||'Untitled')+'"?')){await api('DELETE','/memos/'+m.id);closeMemoCtx();loadMemos()}}},
      'Delete')
  );
  menu.style.position='fixed';
  if(e.type==='contextmenu'){
    menu.style.top=e.clientY+'px';
    menu.style.left=Math.min(e.clientX,window.innerWidth-180)+'px';
  }else{
    menu.style.top=rect.bottom+4+'px';
    menu.style.left=Math.min(rect.left,window.innerWidth-180)+'px';
  }
  document.body.appendChild(menu);
  memoCtx=menu;
  document.addEventListener('click',closeMemoCtx,{once:true});
}

let memoTagFilter=null,allMemos=[];
async function loadMemos(){
  if(!allTagsCache.length)await fetchAllTags();
  const r=await api('GET','/memos?limit=100');allMemos=r.data||[];
  buildMemoTagTree(allMemos);
  renderMemoList();
}
function buildMemoTagTree(memos){
  const tree=$('memo-tags-tree');tree.innerHTML='';
  // Count tags
  const counts={};
  memos.forEach(m=>(m.tags||[]).forEach(t=>{counts[t.name]=(counts[t.name]||0)+1}));
  if(Object.keys(counts).length===0)return;
  // "All" item
  const allItem=el('div',{cls:'mt-item mt-all'+(memoTagFilter===null?' active':''),onClick:()=>{memoTagFilter=null;loadMemos()}});
  allItem.innerHTML=`All <span class="mt-count">${memos.length}</span>`;
  tree.appendChild(allItem);
  // Sort by count desc
  const sorted=Object.entries(counts).sort((a,b)=>b[1]-a[1]);
  sorted.forEach(([name,cnt])=>{
    const item=el('div',{cls:'mt-item'+(memoTagFilter===name?' active':''),onClick:()=>{memoTagFilter=name;renderMemoList();buildMemoTagTree(allMemos)}});
    item.innerHTML=`${name} <span class="mt-count">${cnt}</span>`;
    tree.appendChild(item);
  });
}
function renderMemoList(){
  const c=$('memo-list');c.innerHTML='';
  let items=allMemos;
  if(memoTagFilter){
    items=items.filter(m=>(m.tags||[]).some(t=>t.name===memoTagFilter));
  }
  if(!items.length){c.appendChild(el('div',{cls:'empty'},'No memos'));return}
  items.forEach(m=>{
    const tc=m.type==='table'?'memo-type-tbl':'memo-type-md';
    const menuBtn=el('button',{cls:'memo-menu-btn',onClick:e=>showMemoCtx(e,m)},'⋯');
    const row=el('div',{cls:'memo-item',onClick:()=>showMmDetail(m.id)},
      el('span',{cls:'memo-type '+tc},m.type==='table'?'TBL':'MD'),
      el('div',{cls:'memo-body'},
        el('div',{cls:'memo-title'},m.title||'Untitled'),
        m.content?el('div',{cls:'memo-preview'},m.content):null
      ),
      el('div',{cls:'memo-meta'},
        ...((m.tags||[]).map(t=>el('span',{cls:'tag'},t.name))),
        el('span',null,rel(m.updated_at))
      ),
      menuBtn
    );
    row.addEventListener('contextmenu',e=>{e.preventDefault();showMemoCtx(e,m)});
    c.appendChild(row);
  });
}

// === TODOS ===
let tFilter='open';
function ckSvg(done){
  const s=document.createElementNS('http://www.w3.org/2000/svg','svg');
  s.setAttribute('width','20');s.setAttribute('height','20');s.setAttribute('viewBox','0 0 20 20');
  s.classList.add('todo-ck');if(done)s.classList.add('done');
  const c=document.createElementNS('http://www.w3.org/2000/svg','circle');
  c.setAttribute('cx','10');c.setAttribute('cy','10');c.setAttribute('r','8.5');
  const p=document.createElementNS('http://www.w3.org/2000/svg','polyline');
  p.classList.add('ck-mark');p.setAttribute('points','6,10 9,13 14,7');
  s.appendChild(c);s.appendChild(p);return s;
}
function dueFmt(d){
  if(!d)return null;
  const td=new Date(),du=new Date(d+'T00:00:00');
  const diff=Math.floor((du-new Date(td.getFullYear(),td.getMonth(),td.getDate()))/(864e5));
  if(diff<0)return{text:Math.abs(diff)+'d overdue',cls:'over'};
  if(diff===0)return{text:'Today',cls:'today'};
  if(diff===1)return{text:'Tomorrow',cls:''};
  if(diff<=7)return{text:diff+'d',cls:''};
  return{text:du.toLocaleDateString('ja-JP',{month:'short',day:'numeric'}),cls:''};
}
const calIcon='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/></svg>';
function todoRow(t){
  const done=t.status==='done';
  const ck=ckSvg(done);
  ck.addEventListener('click',e=>{
    e.stopPropagation();
    if(!done){
      const row=ck.closest('.todo-item');
      ck.classList.add('done');
      row.querySelector('.todo-title').classList.add('done');
      setTimeout(()=>{row.classList.add('completing');
        setTimeout(()=>{api('PATCH','/todos/'+t.id+'/done').then(loadTodos)},400);
      },300);
    } else {api('PATCH','/todos/'+t.id+'/reopen').then(loadTodos)}
  });
  const tags=(t.tags||[]).map((x,i)=>el('span',{cls:'todo-tag todo-tag-'+i%5},x.name));
  const df=dueFmt(t.due_date);
  let dueEl=null;
  if(df){const s=el('span',{cls:'todo-due'+(df.cls?' '+df.cls:'')});s.innerHTML=calIcon+' ';s.appendChild(document.createTextNode(df.text));dueEl=s}
  const meta=(tags.length||dueEl)?el('div',{cls:'todo-meta'},...tags,dueEl):null;
  const row=el('div',{cls:'todo-item'+(openTodoId===t.id?' selected':''),'data-id':t.id,onClick:()=>showTdDetail(t.id),draggable:'true'},
    ck,
    el('div',{cls:'todo-body'},
      el('div',{cls:'todo-title'+(done?' done':'')},t.title),
      t.description?el('div',{cls:'todo-desc'},t.description):null,
      meta
    )
  );
  row.addEventListener('dragstart',e=>{e.dataTransfer.setData('text/plain',t.id);row.classList.add('dragging')});
  row.addEventListener('dragend',()=>row.classList.remove('dragging'));
  return row;
}
async function loadTodos(){
  if(!$('todo-inp'))initTodoAdd();
  const r=await api('GET','/todos?limit=100');let items=r.data||[];
  const c=$('todo-list');c.innerHTML='';
  if(tFilter==='all'){
    const open=items.filter(t=>t.status==='open'),done=items.filter(t=>t.status==='done');
    if(!items.length){c.appendChild(el('div',{cls:'todo-empty'},el('div',{cls:'todo-empty-icon'},'✓'),'No items yet'));return}
    open.forEach(t=>c.appendChild(todoRow(t)));
    if(done.length){c.appendChild(el('div',{cls:'todo-divider'},'Done ('+done.length+')'));done.forEach(t=>c.appendChild(todoRow(t)))}
  } else {
    items=items.filter(t=>tFilter==='done'?t.status==='done':t.status==='open');
    if(!items.length){c.appendChild(el('div',{cls:'todo-empty'},el('div',{cls:'todo-empty-icon'},tFilter==='open'?'✓':'—'),tFilter==='open'?'All done!':'No completed items'));return}
    items.forEach(t=>c.appendChild(todoRow(t)));
  }
  // D&D reorder
  c.addEventListener('dragover',e=>{e.preventDefault();const after=getDragAfter(c,e.clientY);const drag=c.querySelector('.dragging');if(!drag)return;if(after)c.insertBefore(drag,after);else c.appendChild(drag)});
}
function getDragAfter(c,y){
  const els=[...c.querySelectorAll('.todo-item:not(.dragging)')];
  let closest=null,off=Number.NEGATIVE_INFINITY;
  els.forEach(e=>{const box=e.getBoundingClientRect();const o=y-box.top-box.height/2;if(o<0&&o>off){off=o;closest=e}});
  return closest;
}
$('todo-tabs').addEventListener('click',e=>{const b=e.target.closest('.tab');if(!b)return;tFilter=b.dataset.f;document.querySelectorAll('#todo-tabs .tab').forEach(x=>x.classList.toggle('active',x.dataset.f===tFilter));loadTodos()});

// === CALENDAR ===
let cY,cM,calEvents=[],selectedDay=null;
const MAX_CELL_EVENTS=3;
function initCal(){const n=new Date();cY=n.getFullYear();cM=n.getMonth();drawCal()}
function dateKey(y,m,d){return `${y}-${String(m+1).padStart(2,'0')}-${String(d).padStart(2,'0')}`}
function eventsForDay(dk){return calEvents.filter(e=>{const d=new Date(e.start_at);return dateKey(d.getFullYear(),d.getMonth(),d.getDate())===dk})}
function drawCal(){
  const c=$('cal-grid'),dw=['Mo','Tu','We','Th','Fr','Sa','Su'];
  const f=new Date(cY,cM,1),L=new Date(cY,cM+1,0),s=(f.getDay()+6)%7,td=new Date();
  c.innerHTML='';
  c.appendChild(el('div',{cls:'cal-hdr'},
    el('span',{cls:'cal-hdr-t'},`${cY}/${String(cM+1).padStart(2,'0')}`),
    el('div',{cls:'cal-nav'},
      el('button',{onClick:()=>{cM--;if(cM<0){cM=11;cY--}loadCalendar()}},'\u25C0'),
      el('button',{cls:'cal-today-btn',onClick:()=>{const n=new Date();cY=n.getFullYear();cM=n.getMonth();loadCalendar()}},'Today'),
      el('button',{onClick:()=>{cM++;if(cM>11){cM=0;cY++}loadCalendar()}},'\u25B6'),
    )
  ));
  const g=el('div',{cls:'cal-g'});
  dw.forEach(d=>g.appendChild(el('div',{cls:'cal-dw'},d)));
  const pL=new Date(cY,cM,0).getDate();
  for(let i=s-1;i>=0;i--){
    const cell=el('div',{cls:'cal-cell om'},el('span',{cls:'cal-cell-day'},String(pL-i)));
    g.appendChild(cell);
  }
  for(let d=1;d<=L.getDate();d++){
    const isT=d===td.getDate()&&cM===td.getMonth()&&cY===td.getFullYear();
    const dk=dateKey(cY,cM,d);
    const dayEvs=eventsForDay(dk);
    const cell=el('div',{cls:'cal-cell'+(isT?' today':''),onClick:()=>showDayDetail(cY,cM,d)});
    cell.appendChild(el('span',{cls:'cal-cell-day'},String(d)));
    const evsC=el('div',{cls:'cal-cell-evs'});
    dayEvs.slice(0,MAX_CELL_EVENTS).forEach((ev,i)=>{
      const chip=el('div',{cls:'cal-ev cal-ev-'+i%4+(ev.all_day?' cal-ev-allday':''),onClick:e=>{e.stopPropagation();showEvDetail(ev.id)}},
        ev.all_day?ev.title:hm(ev.start_at)+' '+ev.title);
      evsC.appendChild(chip);
    });
    if(dayEvs.length>MAX_CELL_EVENTS)evsC.appendChild(el('div',{cls:'cal-more'},`+${dayEvs.length-MAX_CELL_EVENTS} more`));
    cell.appendChild(evsC);
    g.appendChild(cell);
  }
  const tot=s+L.getDate(),rem=(7-tot%7)%7;
  for(let d=1;d<=rem;d++){
    const cell=el('div',{cls:'cal-cell om'},el('span',{cls:'cal-cell-day'},String(d)));
    g.appendChild(cell);
  }
  c.appendChild(g);
  if(selectedDay)showDayDetail(selectedDay[0],selectedDay[1],selectedDay[2]);
}
function closeCalDetail(){
  selectedDay=null;
  $('cal-detail').classList.add('hidden');
  $('cal-detail').innerHTML='';
  document.querySelectorAll('.cal-cell.selected').forEach(e=>e.classList.remove('selected'));
}
function showDayDetail(y,m,d){
  selectedDay=[y,m,d];
  document.querySelectorAll('.cal-cell.selected').forEach(e=>e.classList.remove('selected'));
  const dk=dateKey(y,m,d);
  // highlight selected cell
  document.querySelectorAll('.cal-cell').forEach(c=>{
    const dayEl=c.querySelector('.cal-cell-day');
    if(dayEl&&dayEl.textContent===String(d)&&!c.classList.contains('om'))c.classList.add('selected');
  });

  const dayEvs=eventsForDay(dk);
  const panel=$('cal-detail');
  panel.classList.remove('hidden');
  panel.innerHTML='';

  const label=new Date(y,m,d).toLocaleDateString('ja-JP',{month:'long',day:'numeric',weekday:'short'});
  panel.appendChild(el('div',{cls:'cd-head'},
    el('span',{cls:'cd-date'},label),
    el('button',{cls:'cd-close',onClick:closeCalDetail},'×')
  ));

  // Event list
  const evList=el('div',{cls:'cd-events'});
  if(!dayEvs.length)evList.appendChild(el('div',{cls:'cd-empty'},'No events'));
  else dayEvs.forEach(ev=>{
    evList.appendChild(el('div',{cls:'cd-ev',onClick:()=>showEvInPanel(ev.id)},
      el('div',{cls:'cd-ev-time'},ev.all_day?'All day':hm(ev.start_at)+(ev.end_at?' – '+hm(ev.end_at):'')),
      el('div',{cls:'cd-ev-title'},ev.title),
      ev.description?el('div',{cls:'cd-ev-desc'},ev.description):null
    ));
  });
  panel.appendChild(evList);

  // New event form
  panel.appendChild(el('div',{cls:'cd-divider'},'New Event'));
  const defStart=dk+'T09:00';
  const defEnd=dk+'T10:00';
  const form=el('div',{cls:'cd-form'},
    el('input',{type:'text',id:'cd-ne-t',placeholder:'Event title'}),
    el('div',{cls:'cd-form-row'},
      el('input',{type:'datetime-local',id:'cd-ne-s',value:defStart}),
      el('input',{type:'datetime-local',id:'cd-ne-e',value:defEnd})
    ),
    el('textarea',{id:'cd-ne-d',placeholder:'Description (optional)'}),
    el('button',{cls:'btn-primary',style:'align-self:flex-end',onClick:async()=>{
      const t=$('cd-ne-t').value.trim();if(!t)return;
      await api('POST','/events',{title:t,description:$('cd-ne-d').value,start_at:new Date($('cd-ne-s').value).toISOString(),end_at:$('cd-ne-e').value?new Date($('cd-ne-e').value).toISOString():null,all_day:false,tags:[]});
      loadCalendar();
    }},'Add')
  );
  panel.appendChild(form);
  $('cd-ne-t')?.focus();
}

async function showEvInPanel(id){
  const r=await api('GET','/events/'+id),ev=r.data;
  const panel=$('cal-detail');
  panel.innerHTML='';
  const sL=ev.start_at?new Date(ev.start_at).toISOString().slice(0,16):'';
  const eL=ev.end_at?new Date(ev.end_at).toISOString().slice(0,16):'';

  const titleInp=el('input',{cls:'td-title-input',value:ev.title});
  const descInp=el('textarea',{cls:'td-desc',placeholder:'Description...'});
  descInp.textContent=ev.description||'';
  const startInp=el('input',{cls:'td-date-input',type:'datetime-local',value:sL});
  const endInp=el('input',{cls:'td-date-input',type:'datetime-local',value:eL});

  async function saveEv(){
    await api('PUT','/events/'+id,{title:titleInp.value,description:descInp.value,start_at:new Date(startInp.value).toISOString(),end_at:endInp.value?new Date(endInp.value).toISOString():null,all_day:ev.all_day,tags:(ev.tags||[]).map(t=>t.name)});
    loadCalendar();
  }

  panel.appendChild(el('div',{cls:'cd-head'},
    el('button',{cls:'btn-ghost',style:'font-size:.75rem',onClick:()=>{if(selectedDay)showDayDetail(...selectedDay)}},'← Back'),
    el('button',{cls:'cd-close',onClick:closeCalDetail},'×')
  ));
  panel.appendChild(titleInp);
  panel.appendChild(el('div',{cls:'cd-edit-field'},el('span',{cls:'cd-edit-label'},'Start'),startInp));
  panel.appendChild(el('div',{cls:'cd-edit-field'},el('span',{cls:'cd-edit-label'},'End'),endInp));
  panel.appendChild(el('div',{cls:'cd-edit-field'},el('span',{cls:'cd-edit-label'},'Description'),descInp));
  panel.appendChild(el('div',{cls:'td-actions'},
    el('button',{cls:'btn-danger',onClick:async()=>{if(confirm('Delete?')){await api('DELETE','/events/'+id);loadCalendar();if(selectedDay)showDayDetail(...selectedDay)}}},'Delete'),
    el('div',{style:'flex:1'}),
    el('button',{cls:'btn-primary',onClick:async()=>{await saveEv();if(selectedDay)showDayDetail(...selectedDay)}},'Save')
  ));
  titleInp.focus();
}
async function loadCalendar(){
  const r=await api('GET','/events?limit=100&sort=start_at:asc');
  calEvents=r.data||[];
  drawCal();
}

// === TOOLS ===
async function loadTools(){
  const r=await api('GET','/tools');const items=r.data||[];
  const c=$('tool-grid');c.innerHTML='';
  if(!items.length){c.appendChild(el('div',{cls:'empty'},'No tools yet'));return}
  items.forEach((t,i)=>{
    const letterIcon=()=>el('div',{cls:'tool-icon tool-icon-'+i%6},(t.name||'?')[0].toUpperCase());
    let icon;
    if(t.icon){
      // Use server-fetched favicon (data URI)
      const img=el('img',{cls:'tool-favicon'});
      img.alt=t.name;
      img.src=t.icon;
      img.onerror=()=>img.replaceWith(letterIcon());
      icon=el('div',{cls:'tool-icon-wrap'},img);
    }else{
      icon=letterIcon();
    }
    const delBtn=el('button',{cls:'tool-del',title:'Delete'});
    delBtn.textContent='×';
    delBtn.onclick=async(e)=>{
      e.preventDefault();e.stopPropagation();
      if(!confirm(`Delete "${t.name}"?`))return;
      await api('DELETE','/tools/'+t.id);loadTools();
    };
    c.appendChild(el('a',{cls:'tool-card',href:t.url,target:'_blank',rel:'noopener'},icon,el('div',{cls:'tool-name'},t.name),delBtn));
  });
}

const loaders={home:loadHome,memos:loadMemos,todos:loadTodos,calendar:loadCalendar,tools:loadTools};

// === CREATE ACTIONS ===
document.addEventListener('click',e=>{const b=e.target.closest('[data-action]');if(!b)return;({
  'new-memo':newMm,'new-todo':newTd,'new-event':newEv,'new-tool':newTl
})[b.dataset.action]?.()});

// === FORM HELPERS ===
function fg(l,t,id,ph,v){const g=el('div',{cls:'fg'},el('label',null,l));if(t==='textarea'){const a=el('textarea',{id,placeholder:ph||''});if(v)a.textContent=v;g.appendChild(a)}else{const i=el('input',{type:t,id,placeholder:ph||''});if(v!==undefined)i.value=v;g.appendChild(i)}return g}
function fa(...btns){return el('div',{cls:'fa'},...btns)}
function pt(s){return(s||'').split(',').map(x=>x.trim()).filter(Boolean)}

// === MEMO EDITOR (fullscreen) ===
let memoEditId=null,memoEditTags=[],memoSaveTimer=null,cmView=null,previewOpen=false;

function getMemoContent(){
  if(cmView)return cmView.state.doc.toString();
  return $('me-ta')?.value||'';
}

function openMemoEditor(id,data){
  memoEditId=id;
  memoEditTags=(data.tags||[]).map(t=>typeof t==='string'?t:t.name);
  document.body.classList.add('memo-edit-active');

  // Switch to editor page
  currentPage='memo-edit';
  document.querySelectorAll('.page').forEach(p=>p.classList.remove('active'));
  $('page-memo-edit').classList.add('active');
  document.querySelectorAll('.sb-item[data-page]').forEach(b=>b.classList.toggle('active',b.dataset.page==='memos'));

  $('me-title').value=data.title||'';
  $('me-status').textContent='';

  // Destroy previous CM instance
  if(cmView){cmView.destroy();cmView=null}

  const edArea=$('me-editor');
  edArea.innerHTML='';

  function onDocChange(){
    $('me-status').textContent='Editing...';
    clearTimeout(memoSaveTimer);
    memoSaveTimer=setTimeout(()=>saveMemo(),1500);
    updatePreview();
  }
  function updatePreview(){
    const pv=$('me-preview');
    if(!pv||pv.classList.contains('hidden'))return;
    const md=getMemoContent();
    pv.innerHTML=window.marked?marked.parse(md):md.replace(/</g,'&lt;').replace(/\n/g,'<br>');
  }

  // Wait for CodeMirror module to load (async ESM), then init editor
  function initEditor(){
    cmView=window.createCM(edArea,data.content||'',onDocChange);
    edArea.addEventListener('keydown',e=>{
      if((e.ctrlKey||e.metaKey)&&e.key==='s'){e.preventDefault();saveMemo()}
    });
    cmView.focus();
  }
  if(window.createCM){
    initEditor();
  }else{
    edArea.textContent='Loading editor...';
    let cmLoaded=false;
    const wt=setInterval(()=>{
      if(window.createCM){clearInterval(wt);cmLoaded=true;edArea.textContent='';initEditor()}
    },100);
    setTimeout(()=>{
      if(!cmLoaded&&!window.createCM){
        clearInterval(wt);edArea.textContent='';
        const ta=el('textarea',{id:'me-ta',placeholder:'Write in Markdown...',spellcheck:'false'});
        ta.textContent=data.content||'';edArea.appendChild(ta);
        ta.addEventListener('input',onDocChange);ta.focus();
      }
    },8000);
  }

  $('me-title').addEventListener('input',()=>{
    clearTimeout(memoSaveTimer);
    memoSaveTimer=setTimeout(()=>saveMemo(),1500);
  });

  renderMemoTags();

  // Preview toggle
  const pvBtn=$('me-preview-btn');
  const pv=$('me-preview');
  pv.classList.add('hidden');
  previewOpen=false;
  pvBtn.onclick=()=>{
    previewOpen=!previewOpen;
    pv.classList.toggle('hidden',!previewOpen);
    pvBtn.textContent=previewOpen?'Edit only':'Preview';
    if(previewOpen)updatePreview();
  };
}

function renderMemoTags(){
  const bar=$('me-tag-bar');bar.innerHTML='';bar.style.position='relative';
  memoEditTags.forEach(name=>{
    bar.appendChild(el('span',{cls:'td-tag-chip'},name,el('button',{onClick:()=>{memoEditTags=memoEditTags.filter(n=>n!==name);renderMemoTags();saveMemo()}},'×')));
  });
  const addBtn=el('button',{cls:'td-tag-add',onClick:e=>{
    e.stopPropagation();
    const existing=bar.querySelector('.td-tag-dropdown');
    if(existing){existing.remove();return}
    const dd=el('div',{cls:'td-tag-dropdown'});
    const available=allTagsCache.filter(n=>!memoEditTags.includes(n));
    if(!available.length)dd.appendChild(el('div',{style:'padding:8px;color:var(--text3);font-size:.75rem'},'No more tags'));
    else available.forEach(name=>{
      dd.appendChild(el('button',{onClick:()=>{memoEditTags.push(name);renderMemoTags();saveMemo();dd.remove()}},name));
    });
    bar.appendChild(dd);
    document.addEventListener('click',()=>dd.remove(),{once:true});
  }},'+');
  bar.appendChild(addBtn);
  $('me-tag-btn').textContent=memoEditTags.length?`Tags (${memoEditTags.length})`:'Tags';
  $('me-tag-btn').classList.toggle('has-val',memoEditTags.length>0);
}

async function saveMemo(){
  if(!memoEditId)return;
  const title=$('me-title').value;
  const content=getMemoContent();
  try{
    await api('PUT','/memos/'+memoEditId,{title,content,tags:memoEditTags});
    $('me-status').textContent='Saved';
    setTimeout(()=>{if($('me-status').textContent==='Saved')$('me-status').textContent=''},2000);
  }catch(err){$('me-status').textContent='Error saving'}
}

function closeMemoEditor(){
  clearTimeout(memoSaveTimer);
  saveMemo();
  if(cmView){cmView.destroy();cmView=null}
  memoEditId=null;
  document.body.classList.remove('memo-edit-active');
  navigateTo('memos');
}

async function newMm(){
  const r=await api('POST','/memos',{title:'',type:'markdown',content:'',tags:[]});
  openMemoEditor(r.data.id,r.data);
  $('me-title').focus();
}

async function showMmDetail(id){
  const r=await api('GET','/memos/'+id);
  openMemoEditor(id,r.data);
}

// Back button & delete
$('me-back').addEventListener('click',closeMemoEditor);
$('me-delete').addEventListener('click',async()=>{
  if(!memoEditId||!confirm('Delete this memo?'))return;
  await api('DELETE','/memos/'+memoEditId);
  memoEditId=null;navigateTo('memos');
});
$('me-tag-btn').addEventListener('click',()=>{
  const bar=$('me-tag-bar');
  bar.querySelector('.td-tag-add')?.click();
});

// === TODO INLINE ADD ===
let todoAddDue='',todoAddTagNames=[],allTagsCache=[];
async function fetchAllTags(){try{const r=await api('GET','/tags');allTagsCache=(r.data||[]).map(t=>t.name)}catch{allTagsCache=[]}}

function dateDelta(n){const d=new Date();d.setDate(d.getDate()+n);return d.toISOString().slice(0,10)}
function nextMonday(){const d=new Date(),day=d.getDay(),add=day===0?1:(8-day);d.setDate(d.getDate()+add);return d.toISOString().slice(0,10)}

const DATE_SHORTCUTS={'!today':()=>dateDelta(0),'!tomorrow':()=>dateDelta(1),'!nextweek':()=>nextMonday()};

function initTodoAdd(){
  const c=$('todo-add');c.innerHTML='';
  todoAddDue='';todoAddTagNames=[];
  fetchAllTags();

  const inp=el('input',{type:'text',placeholder:'Add a task... Enter to create, !today !tomorrow for date, #tag',id:'todo-inp'});
  const wrapper=el('div',{cls:'todo-add',style:'position:relative'});

  // --- Date button + popover ---
  const dueBtn=el('button',{id:'todo-due-btn'},'Due');
  let datePop=null;
  function closeDatePop(){if(datePop){datePop.remove();datePop=null}}
  function updateDueBtn(){dueBtn.classList.toggle('has-val',!!todoAddDue);dueBtn.textContent=todoAddDue||'Due'}
  function setDue(v){todoAddDue=v;updateDueBtn();closeDatePop();inp.focus()}
  dueBtn.addEventListener('click',e=>{
    e.stopPropagation();closeTagPop();
    if(datePop){closeDatePop();return}
    const hiddenDate=el('input',{type:'date'});
    hiddenDate.addEventListener('change',()=>setDue(hiddenDate.value));
    datePop=el('div',{cls:'todo-date-pop'},
      el('button',{onClick:()=>setDue(dateDelta(0))},'Today'),
      el('button',{onClick:()=>setDue(dateDelta(1))},'Tomorrow'),
      el('button',{onClick:()=>setDue(nextMonday())},'Next Week'),
      el('button',{onClick:()=>{try{hiddenDate.showPicker()}catch{hiddenDate.click()}}},'Pick...'),
      todoAddDue?el('button',{cls:'dp-clear',onClick:()=>setDue('')},'Clear'):null,
      hiddenDate
    );
    wrapper.appendChild(datePop);
  });

  // --- Tag button + selector ---
  const tagBtn=el('button',{id:'todo-tag-btn'},'Tags');
  let tagPop=null;
  function closeTagPop(){if(tagPop){tagPop.remove();tagPop=null}}
  function updateTagBtn(){tagBtn.classList.toggle('has-val',todoAddTagNames.length>0);tagBtn.textContent=todoAddTagNames.length?todoAddTagNames.join(', '):'Tags'}
  function renderTagPop(){
    closeTagPop();
    if(!allTagsCache.length){tagPop=el('div',{cls:'todo-tag-pop'},el('div',{cls:'todo-tag-empty'},'No tags yet'));wrapper.appendChild(tagPop);return}
    const list=el('div',{cls:'todo-tag-pop-list'});
    allTagsCache.forEach(name=>{
      const selected=todoAddTagNames.includes(name);
      const opt=el('button',{cls:'todo-tag-opt'+(selected?' selected':''),onClick:()=>{
        if(todoAddTagNames.includes(name))todoAddTagNames=todoAddTagNames.filter(n=>n!==name);
        else todoAddTagNames.push(name);
        updateTagBtn();renderTagPop();
      }},el('span',{cls:'tto-ck'},'✓'),name);
      list.appendChild(opt);
    });
    tagPop=el('div',{cls:'todo-tag-pop'},list);
    wrapper.appendChild(tagPop);
  }
  tagBtn.addEventListener('click',e=>{e.stopPropagation();closeDatePop();if(tagPop){closeTagPop();return}renderTagPop()});

  // --- Close popovers on outside click ---
  document.addEventListener('click',()=>{closeDatePop();closeTagPop()});
  wrapper.addEventListener('click',e=>e.stopPropagation());

  // --- Submit ---
  inp.addEventListener('keydown',async e=>{
    if(e.key==='Enter'&&inp.value.trim()){
      let title=inp.value.trim();
      // Parse !shortcuts
      for(const[kw,fn] of Object.entries(DATE_SHORTCUTS)){
        if(title.toLowerCase().includes(kw)){todoAddDue=fn();title=title.replace(new RegExp(kw.replace('!','\\!'),'gi'),'').trim();updateDueBtn()}
      }
      // Parse #tags
      const hashTags=[...title.matchAll(/#(\S+)/g)].map(m=>m[1]);
      if(hashTags.length){todoAddTagNames=[...new Set([...todoAddTagNames,...hashTags])];title=title.replace(/#\S+/g,'').trim()}
      if(!title)return;
      inp.value='';inp.disabled=true;
      await api('POST','/todos',{title,due_date:todoAddDue||null,tags:todoAddTagNames});
      inp.disabled=false;inp.focus();
      todoAddDue='';todoAddTagNames=[];updateDueBtn();updateTagBtn();
      loadTodos();
    }
  });

  wrapper.appendChild(el('span',{cls:'todo-add-icon'},'+'));
  wrapper.appendChild(inp);
  wrapper.appendChild(el('div',{cls:'todo-add-meta'},dueBtn,tagBtn));
  c.appendChild(wrapper);
}
function newTd(){$('todo-inp')?.focus()}

let openTodoId=null;
function closeTdDetail(){
  openTodoId=null;
  $('todo-detail').classList.add('hidden');
  $('todo-detail').innerHTML='';
  document.querySelectorAll('.todo-item.selected').forEach(e=>e.classList.remove('selected'));
}
async function showTdDetail(id){
  openTodoId=id;
  document.querySelectorAll('.todo-item.selected').forEach(e=>e.classList.remove('selected'));
  const row=document.querySelector(`.todo-item[data-id="${id}"]`);
  if(row)row.classList.add('selected');

  const r=await api('GET','/todos/'+id),t=r.data;
  const panel=$('todo-detail');
  panel.classList.remove('hidden');
  panel.innerHTML='';

  const titleInp=el('input',{cls:'td-title-input',value:t.title});
  const descInp=el('textarea',{cls:'td-desc',placeholder:'Add notes...'});
  descInp.textContent=t.description||'';
  const dateInp=el('input',{cls:'td-date-input',type:'date',value:t.due_date||''});

  // Tags
  let detailTags=(t.tags||[]).map(x=>x.name);
  const tagsWrap=el('div',{cls:'td-tags',style:'position:relative'});
  function renderDetailTags(){
    tagsWrap.innerHTML='';
    detailTags.forEach(name=>{
      const chip=el('span',{cls:'td-tag-chip'},name,el('button',{onClick:()=>{detailTags=detailTags.filter(n=>n!==name);renderDetailTags()}},'×'));
      tagsWrap.appendChild(chip);
    });
    const addBtn=el('button',{cls:'td-tag-add',onClick:e=>{
      e.stopPropagation();
      const existing=tagsWrap.querySelector('.td-tag-dropdown');
      if(existing){existing.remove();return}
      const dd=el('div',{cls:'td-tag-dropdown'});
      const available=allTagsCache.filter(n=>!detailTags.includes(n));
      if(!available.length){dd.appendChild(el('div',{style:'padding:8px;color:var(--text3);font-size:.75rem'},'No more tags'));
      }else{available.forEach(name=>{
        dd.appendChild(el('button',{onClick:()=>{detailTags.push(name);renderDetailTags();dd.remove()}},name));
      })}
      tagsWrap.appendChild(dd);
      document.addEventListener('click',()=>dd.remove(),{once:true});
    }},'+');
    tagsWrap.appendChild(addBtn);
  }
  renderDetailTags();

  // Auto-save on blur
  async function save(){
    await api('PUT','/todos/'+id,{title:titleInp.value,description:descInp.value,status:t.status,due_date:dateInp.value||null,tags:detailTags});
    loadTodos();
  }
  titleInp.addEventListener('blur',save);
  descInp.addEventListener('blur',save);
  dateInp.addEventListener('change',save);

  panel.appendChild(el('div',{cls:'td-head'},
    el('span',{style:'font-size:.75rem;color:var(--text3)'},t.status==='done'?'Completed':'Open'),
    el('button',{cls:'td-close',onClick:closeTdDetail},'×')
  ));
  panel.appendChild(titleInp);
  panel.appendChild(el('div',{cls:'td-field'},el('span',{cls:'td-label'},'Notes'),descInp));
  panel.appendChild(el('div',{cls:'td-field'},el('span',{cls:'td-label'},'Due Date'),dateInp));
  panel.appendChild(el('div',{cls:'td-field'},el('span',{cls:'td-label'},'Tags'),tagsWrap));
  panel.appendChild(el('div',{cls:'td-actions'},
    el('button',{cls:'btn-danger',onClick:async()=>{if(confirm('Delete?')){await api('DELETE','/todos/'+id);closeTdDetail();loadTodos()}}},'Delete'),
    el('div',{style:'flex:1'}),
    el('button',{cls:'btn-primary',onClick:async()=>{await save();closeTdDetail()}},'Done')
  ));
  titleInp.focus();
}

// === EVENT ===
function newEv(){
  // Open calendar and show today's panel
  navigateTo('calendar');
  const n=new Date();showDayDetail(n.getFullYear(),n.getMonth(),n.getDate());
  setTimeout(()=>$('cd-ne-t')?.focus(),100);
}

async function showEvDetail(id){
  navigateTo('calendar');
  showEvInPanel(id);
}

// === TOOL MODAL ===
function newTl(){showM('New Tool',el('div',null,
  fg('Name','text','ntl-n','Name'),
  fg('URL','url','ntl-u','https://...'),
  fg('Icon','text','ntl-i','Emoji or letter'),
  fa(el('button',{cls:'btn-ghost',onClick:closeM},'Cancel'),
     el('button',{cls:'btn-primary',onClick:async()=>{await api('POST','/tools',{name:$('ntl-n').value,url:$('ntl-u').value,icon:$('ntl-i').value});closeM();loadTools()}},'Create'))
));$('ntl-n').focus()}

// === MODAL ===
function showM(t,b){$('modal-title').textContent=t;const mb=$('modal-body');mb.innerHTML='';mb.appendChild(b);$('modal-bg').classList.remove('hidden')}
function closeM(){$('modal-bg').classList.add('hidden')}
$('modal-x').addEventListener('click',closeM);
$('modal-bg').addEventListener('click',e=>{if(e.target===$('modal-bg'))closeM()});

// === COMMAND PALETTE ===
let cmdCache=[];
async function loadCmdCache(){
  const [mm,td,ev]=await Promise.all([api('GET','/memos?limit=50'),api('GET','/todos?limit=100'),api('GET','/events?limit=30')]);
  cmdCache=[
    ...(mm.data||[]).map(m=>({type:'memo',title:m.title||'Untitled',id:m.id})),
    ...(td.data||[]).map(t=>({type:'todo',title:t.title,id:t.id})),
    ...(ev.data||[]).map(e=>({type:'event',title:e.title,id:e.id})),
  ];
}
function openCmd(){$('cmd-bg').classList.remove('hidden');$('cmd-input').value='';$('cmd-input').focus();renderCmd('');loadCmdCache()}
function closeCmd(){$('cmd-bg').classList.add('hidden')}
function renderCmd(q){
  const c=$('cmd-results');c.innerHTML='';
  const lq=q.toLowerCase();
  const filtered=q?cmdCache.filter(x=>x.title.toLowerCase().includes(lq)):cmdCache.slice(0,15);
  if(!filtered.length){c.appendChild(el('div',{cls:'empty'},q?'No results':'Start typing to search...'));return}
  filtered.slice(0,20).forEach(item=>{
    c.appendChild(el('div',{cls:'cmd-result',onClick:()=>{
      closeCmd();
      if(item.type==='memo'){navigateTo('memos');showMmDetail(item.id)}
      else if(item.type==='todo'){navigateTo('todos');showTdDetail(item.id)}
      else if(item.type==='event'){navigateTo('calendar');showEvDetail(item.id)}
    }},
      el('span',{cls:'cmd-result-type'},item.type),
      el('span',null,item.title)
    ));
  });
}
$('cmd-input').addEventListener('input',e=>renderCmd(e.target.value));
$('cmd-bg').addEventListener('click',e=>{if(e.target===$('cmd-bg'))closeCmd()});
$('cmd-trigger').addEventListener('click',openCmd);

// === KEYBOARD SHORTCUTS ===
document.addEventListener('keydown',e=>{
  if(e.key==='Escape'){if(openTodoId)closeTdDetail();else if(selectedDay)closeCalDetail();else{closeM();closeCmd()}}
  if((e.metaKey||e.ctrlKey)&&e.key==='k'){e.preventDefault();openCmd()}
});

// === THEME SWITCHER ===
const savedTheme=localStorage.getItem('konbu-theme')||'konbu';
document.body.dataset.theme=savedTheme;
document.querySelectorAll('.theme-dot').forEach(d=>d.classList.toggle('active',d.dataset.theme===savedTheme));
document.querySelector('.theme-switcher').addEventListener('click',e=>{
  const d=e.target.closest('.theme-dot');if(!d)return;
  document.body.dataset.theme=d.dataset.theme;
  localStorage.setItem('konbu-theme',d.dataset.theme);
  document.querySelectorAll('.theme-dot').forEach(x=>x.classList.toggle('active',x===d));
});

// === INIT ===
initCal();
loadHome();
})();
