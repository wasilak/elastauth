const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["_astro/mermaid.core.CYO9-z7c.js","_astro/preload-helper.Cgw7_4aY.js"])))=>i.map(i=>d[i]);
import{_ as L}from"./preload-helper.Cgw7_4aY.js";const A={},v=new Set,l=new WeakSet;let g=!0,y,p=!1;function S(e){p||(p=!0,g??=!1,y??="hover",T(),M(),C(),P())}function T(){for(const e of["touchstart","mousedown"])document.addEventListener(e,t=>{u(t.target,"tap")&&f(t.target.href,{ignoreSlowConnection:!0})},{passive:!0})}function M(){let e;document.body.addEventListener("focusin",r=>{u(r.target,"hover")&&t(r)},{passive:!0}),document.body.addEventListener("focusout",a,{passive:!0}),h(()=>{for(const r of document.getElementsByTagName("a"))l.has(r)||u(r,"hover")&&(l.add(r),r.addEventListener("mouseenter",t,{passive:!0}),r.addEventListener("mouseleave",a,{passive:!0}))});function t(r){const i=r.target.href;e&&clearTimeout(e),e=setTimeout(()=>{f(i)},80)}function a(){e&&(clearTimeout(e),e=0)}}function C(){let e;h(()=>{for(const t of document.getElementsByTagName("a"))l.has(t)||u(t,"viewport")&&(l.add(t),e??=O(),e.observe(t))})}function O(){const e=new WeakMap;return new IntersectionObserver((t,a)=>{for(const r of t){const i=r.target,o=e.get(i);r.isIntersecting?(o&&clearTimeout(o),e.set(i,setTimeout(()=>{a.unobserve(i),e.delete(i),f(i.href)},300))):o&&(clearTimeout(o),e.delete(i))}})}function P(){h(()=>{for(const e of document.getElementsByTagName("a"))u(e,"load")&&f(e.href)})}function f(e,t){e=e.replace(/#.*/,"");const a=t?.ignoreSlowConnection??!1;if(I(e,a))if(v.add(e),document.createElement("link").relList?.supports?.("prefetch")&&t?.with!=="fetch"){const r=document.createElement("link");r.rel="prefetch",r.setAttribute("href",e),document.head.append(r)}else{const r=new Headers;for(const[i,o]of Object.entries(A))r.set(i,o);fetch(e,{priority:"low",headers:r})}}function I(e,t){if(!navigator.onLine||!t&&k())return!1;try{const a=new URL(e,location.href);return location.origin===a.origin&&(location.pathname!==a.pathname||location.search!==a.search)&&!v.has(e)}catch{}return!1}function u(e,t){if(e?.tagName!=="A")return!1;const a=e.dataset.astroPrefetch;return a==="false"?!1:t==="tap"&&(a!=null||g)&&k()?!0:a==null&&g||a===""?t===y:a===t}function k(){if("connection"in navigator){const e=navigator.connection;return e.saveData||/2g/.test(e.effectiveType)}return!1}function h(e){e();let t=!1;document.addEventListener("astro:page-load",()=>{if(!t){t=!0;return}e()})}const b=()=>document.querySelectorAll("pre.mermaid").length>0;b()?(console.log("[astro-mermaid] Mermaid diagrams detected, loading mermaid.js..."),L(()=>import("./mermaid.core.CYO9-z7c.js").then(e=>e.b8),__vite__mapDeps([0,1])).then(async({default:e})=>{const t=[];if(t&&t.length>0){console.log("[astro-mermaid] Registering",t.length,"icon packs");const o=t.map(s=>({name:s.name,loader:new Function("return "+s.loader)()}));await e.registerIconPacks(o)}const a={startOnLoad:!1,theme:"default"},r={light:"default",dark:"dark"};async function i(){console.log("[astro-mermaid] Initializing mermaid diagrams...");const o=document.querySelectorAll("pre.mermaid");if(console.log("[astro-mermaid] Found",o.length,"mermaid diagrams"),o.length===0)return;let s=a.theme;{const n=document.documentElement.getAttribute("data-theme"),c=document.body.getAttribute("data-theme");s=r[n||c]||a.theme,console.log("[astro-mermaid] Using theme:",s,"from",n?"html":"body")}e.initialize({...a,theme:s,gitGraph:{mainBranchName:"main",showCommitLabel:!0,showBranches:!0,rotateCommitLabel:!0}});for(const n of o){if(n.hasAttribute("data-processed"))continue;n.hasAttribute("data-diagram")||n.setAttribute("data-diagram",n.textContent||"");const c=n.getAttribute("data-diagram")||"",d="mermaid-"+Math.random().toString(36).slice(2,11);console.log("[astro-mermaid] Rendering diagram:",d);try{const m=document.getElementById(d);m&&m.remove();const{svg:E}=await e.render(d,c);n.innerHTML=E,n.setAttribute("data-processed","true"),console.log("[astro-mermaid] Successfully rendered diagram:",d)}catch(m){console.error("[astro-mermaid] Mermaid rendering error for diagram:",d,m),n.innerHTML=`<div style="color: red; padding: 1rem; border: 1px solid red; border-radius: 0.5rem;">
            <strong>Error rendering diagram:</strong><br/>
            ${m.message||"Unknown error"}
          </div>`,n.setAttribute("data-processed","true")}}}i();{const o=new MutationObserver(s=>{for(const n of s)n.type==="attributes"&&n.attributeName==="data-theme"&&(document.querySelectorAll("pre.mermaid[data-processed]").forEach(c=>{c.removeAttribute("data-processed")}),i())});o.observe(document.documentElement,{attributes:!0,attributeFilter:["data-theme"]}),o.observe(document.body,{attributes:!0,attributeFilter:["data-theme"]})}document.addEventListener("astro:after-swap",()=>{b()&&i()})}).catch(e=>{console.error("[astro-mermaid] Failed to load mermaid:",e)})):console.log("[astro-mermaid] No mermaid diagrams found on this page, skipping mermaid.js load");const w=document.createElement("style");w.textContent=`
            /* Prevent layout shifts by setting minimum height */
            pre.mermaid {
              display: flex;
              justify-content: center;
              align-items: center;
              margin: 2rem 0;
              padding: 1rem;
              background-color: transparent;
              border: none;
              overflow: auto;
              min-height: 200px; /* Prevent layout shift */
              position: relative;
            }
            
            /* Loading state with skeleton loader */
            pre.mermaid:not([data-processed]) {
              background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
              background-size: 200% 100%;
              animation: shimmer 1.5s infinite;
            }
            
            /* Dark mode skeleton loader */
            [data-theme="dark"] pre.mermaid:not([data-processed]) {
              background: linear-gradient(90deg, #2a2a2a 25%, #3a3a3a 50%, #2a2a2a 75%);
              background-size: 200% 100%;
            }
            
            @keyframes shimmer {
              0% {
                background-position: -200% 0;
              }
              100% {
                background-position: 200% 0;
              }
            }
            
            /* Show processed diagrams with smooth transition */
            pre.mermaid[data-processed] {
              animation: none;
              background: transparent;
              min-height: auto; /* Allow natural height after render */
            }
            
            /* Ensure responsive sizing for mermaid SVGs */
            pre.mermaid svg {
              max-width: 100%;
              height: auto;
            }
            
            /* Optional: Add subtle background for better visibility */
            @media (prefers-color-scheme: dark) {
              pre.mermaid[data-processed] {
                background-color: rgba(255, 255, 255, 0.02);
                border-radius: 0.5rem;
              }
            }
            
            @media (prefers-color-scheme: light) {
              pre.mermaid[data-processed] {
                background-color: rgba(0, 0, 0, 0.02);
                border-radius: 0.5rem;
              }
            }
            
            /* Respect user's color scheme preference */
            [data-theme="dark"] pre.mermaid[data-processed] {
              background-color: rgba(255, 255, 255, 0.02);
              border-radius: 0.5rem;
            }
            
            [data-theme="light"] pre.mermaid[data-processed] {
              background-color: rgba(0, 0, 0, 0.02);
              border-radius: 0.5rem;
            }
          `;document.head.appendChild(w);S();
