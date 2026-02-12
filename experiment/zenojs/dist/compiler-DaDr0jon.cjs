"use strict";const r={TEXT:"TEXT",ECHO_START:"ECHO_START",ECHO_END:"ECHO_END",DIRECTIVE:"DIRECTIVE",PAREN_START:"PAREN_START",PAREN_END:"PAREN_END",TAG_OPEN:"TAG_OPEN",TAG_CLOSE:"TAG_CLOSE",TAG_END_OPEN:"TAG_END_OPEN",EOF:"EOF"};class T{constructor(t){this.input=t,this.pos=0,this.len=t.length}nextToken(){if(this.pos>=this.len)return{type:r.EOF};const t=this.input[this.pos];if(t==="{"&&this.input[this.pos+1]==="{")return this.pos+=2,{type:r.ECHO_START};if(t==="}"&&this.input[this.pos+1]==="}")return this.pos+=2,{type:r.ECHO_END};if(t==="@"){let e=this.pos+1,s=e;for(;s<this.len&&/[a-zA-Z0-9_]/.test(this.input[s]);)s++;if(s>e){const l=this.input.slice(e,s);return this.pos=s,{type:r.DIRECTIVE,value:l}}}if(t==="<"){if(this.input.startsWith("<x-",this.pos)){let e=this.pos+1,s=e;for(;s<this.len&&/[a-zA-Z0-9\-\.]/.test(this.input[s]);)s++;const l=this.input.slice(e,s);return this.pos=s,{type:r.TAG_OPEN,value:l}}if(this.input.startsWith("</x-",this.pos)){let e=this.pos+2,s=e;for(;s<this.len&&/[a-zA-Z0-9\-\.]/.test(this.input[s]);)s++;const l=this.input.slice(e,s);if(this.input[s]===">")return this.pos=s+1,{type:r.TAG_END_OPEN,value:l}}}if(t===">")return this.pos++,{type:r.TAG_CLOSE};if(t==="(")return this.pos++,{type:r.PAREN_START};if(t===")")return this.pos++,{type:r.PAREN_END};let n=this.pos+1;for(;n<this.len;){const e=this.input[n];if(e==="{"&&this.input[n+1]==="{"||e==="}"&&this.input[n+1]==="}"||e==="@"||e==="<"&&this.input.startsWith("<x-",n)||e==="<"&&this.input.startsWith("</x-",n)||e==="("||e===")"||e===">")break;n++}const i=this.input.slice(this.pos,n);return this.pos=n,{type:r.TEXT,value:i}}}function y(c){const t=new T(c),n={type:"Root",children:[]},i=[n];let e=t.nextToken();for(;e.type!==r.EOF;){const s=i[i.length-1];if(e.type===r.TEXT)s.children.push({type:"Text",value:e.value});else if(e.type===r.ECHO_START){let l="";for(e=t.nextToken();e.type!==r.ECHO_END&&e.type!==r.EOF;)e.value&&(l+=e.value),e.type===r.PAREN_START&&(l+="("),e.type===r.PAREN_END&&(l+=")"),e.type===r.DIRECTIVE&&(l+="@"+e.value),e=t.nextToken();s.children.push({type:"Echo",value:l.trim()})}else if(e.type===r.DIRECTIVE){const l=e.value;if(l.startsWith("end")){const o=l.substring(3);s.type==="Directive"&&(s.name===o||o==="unless"&&s.name==="unless"||o==="empty"&&s.name==="empty"||o==="isset"&&s.name==="isset"||o==="switch"&&s.name==="switch")?i.pop():s.children.push({type:"Text",value:"@"+l})}else if(l==="else"||l==="elseif"||l==="case"||l==="default"||l==="break"){let o=null;const a=t.nextToken();a.type===r.PAREN_START&&(o=f(t));const p={type:"Directive",name:l,args:o,children:[]};s.children.push(p),o!==null||a.type!==r.PAREN_START&&d(s,a)}else{let o=null;const a=t.nextToken();a.type===r.PAREN_START?o=f(t):a.type!==r.EOF&&d(s,a);const p={type:"Directive",name:l,args:o,children:[]};s.children.push(p),["if","unless","isset","empty","switch","foreach","push","component"].includes(l)&&i.push(p)}}else if(e.type===r.TAG_OPEN){const l=e.value;let o="",a=!1;for(e=t.nextToken();e.type!==r.TAG_CLOSE&&e.type!==r.EOF;){if(e.type===r.TEXT){let u=e.value;u.trim().endsWith("/")&&(a=!0,u=u.replace("/","")),o+=u}else e.type===r.ECHO_START?o+="{{":e.value&&(o+=e.value);e=t.nextToken()}const p={type:"Component",tagName:l,attrs:o.trim(),children:[]};s.children.push(p),a||i.push(p)}else if(e.type===r.TAG_END_OPEN){const l=e.value;s.type==="Component"&&s.tagName===l?i.pop():s.children.push({type:"Text",value:`</${l}>`})}e=t.nextToken()}return n}function f(c){let t="",n=1,i=c.nextToken();for(;n>0&&i.type!==r.EOF&&(i.type===r.PAREN_START&&n++,!(i.type===r.PAREN_END&&(n--,n===0)));)i.value?t+=i.value:i.type===r.ECHO_START?t+="{{":i.type===r.ECHO_END?t+="}}":i.type===r.PAREN_START?t+="(":i.type===r.PAREN_END&&(t+=")"),i=c.nextToken();return t}function d(c,t){t.type===r.TEXT?c.children.push({type:"Text",value:t.value}):t.type===r.ECHO_START?c.children.push({type:"Text",value:"{{"}):t.type===r.PAREN_START&&c.children.push({type:"Text",value:"("})}function E(c){const t=y(c);let n=null,i={};const e=t.children,s=[];for(const o of e)if(o.type==="Directive"&&o.name==="extends")n=o.args.replace(/['"]/g,"");else if(o.type==="Directive"&&o.name==="section"){const a=o.args.replace(/['"]/g,"");i[a]=o.children}else s.push(o);let l=`let _out = '';
`;if(l+=`const _helpers = this.$helpers || {};
`,l+=`const _renderComponent = this.renderComponent || (() => '');
`,l+=`const _renderLayout = this.renderLayout || (() => '');
`,l+=`const _renderInclude = this.renderInclude || (() => '');
`,l+=`with(this) {
`,n){const o=s.filter(p=>p.type==="Directive"&&(p.name==="push"||p.name==="inject"));for(const p of o)l+=h(p);let a="{";for(const[p,u]of Object.entries(i)){const _=u.map(h).join("");a+=`
"${p}": (() => {
let _out='';
${_}
return _out;
}),`}a+="}",l+=`_out += _renderLayout('${n}', ${a});
`}else l+=h({type:"Root",children:e});return l+=`}
`,l+="return _out;",l}function h(c){let t="";if(c.type==="Root")for(const n of c.children)t+=h(n);else if(c.type==="Text")c.value&&(t+=`_out += \`${$(c.value)}\`;
`);else if(c.type==="Echo")t+=`_out += ${c.value};
`;else if(c.type==="Directive"){const n=c.name,i=c.args?c.args.trim():"",e=c.children.map(h).join("");if(n==="if")t+=`if (${i}) {
${e}
}
`;else if(n==="elseif")t+=`} else if (${i}) {
${e}
`;else if(n==="else")t+=`} else {
${e}
`;else if(!["endif","endunless","endisset","endempty","endswitch","endforeach","endsection","endpush","endauth","endguest","enderror"].includes(n))if(n==="unless")t+=`if (!(${i})) {
${e}
}
`;else if(n==="isset")t+=`if (typeof ${i} !== 'undefined' && ${i} !== null) {
${e}
}
`;else if(n==="empty")t+=`if (!${i} || (Array.isArray(${i}) && ${i}.length === 0)) {
${e}
}
`;else if(n==="auth")t+=`if (this.auth && this.auth.check()) {
${e}
}
`;else if(n==="guest")t+=`if (!this.auth || this.auth.guest()) {
${e}
}
`;else if(n==="error"){const s=i.replace(/^["']|["']$/g,"");t+=`if (this.$errors && this.$errors['${s}']) {

                 // Inject $message
                 const message = this.$errors['${s}'];
                 ${e}
             }
`}else if(n==="switch")t+=`switch (${i}) {
${e}
}
`;else if(n==="case")t+=`case ${i}:
${e}
`;else if(n==="break")t+=`break;
`;else if(n==="default")t+=`default:
${e}
`;else if(n==="foreach"){const[s,l]=i.split(/\s+as\s+/).map(o=>o.trim());t+=`
                let _loopIndex = 0;
                const _loopList = ${s} || [];
                const _loopCount = _loopList.length;
                for (let ${l} of _loopList) {
                    const loop = {
                        index: _loopIndex,
                        iteration: _loopIndex + 1,
                        remaining: _loopCount - (_loopIndex + 1),
                        count: _loopCount,
                        first: _loopIndex === 0,
                        last: _loopIndex === _loopCount - 1,
                        even: (_loopIndex + 1) % 2 === 0,
                        odd: (_loopIndex + 1) % 2 !== 0,
                        depth: 1
                    };
                    ${e}
                    _loopIndex++;
                }
`}else if(n==="json")t+=`_out += JSON.stringify(${i});
`;else if(n==="class")t+=`_out += 'class="' + _helpers.classNames(${i}) + '"';
`;else if(n==="style")t+=`_out += 'style="' + _helpers.styleNames(${i}) + '"';
`;else if(["checked","selected","disabled","readonly","required"].includes(n))t+=`if (${i}) { _out += '${n}="${n}"'; }
`;else if(n==="click"){const s=i.replace(/^["']|["']$/g,"");t+=`_out += 'data-z-click="${s}"';
`}else if(n==="model"){const s=i.replace(/^["']|["']$/g,"");t+=`_out += 'value="' + (${s}) + '" data-z-model="${s}"';
`}else if(n==="yield"){const s=i.replace(/['"]/g,"");t+=`if (this.$sections && this.$sections['${s}']) { _out += this.$sections['${s}'](); }
`}else if(n==="section")t+=e;else if(n==="push"){const s=i.replace(/['"]/g,"");t+=`
             {
                 let _pushOut = '';
                 this.$push('${s}', (() => { let _out=''; ${e} return _out; })());
             }

`}else if(n==="stack"){const s=i.replace(/['"]/g,"");t+=`_out += this.$stack('${s}');
`}else if(n==="include")t+=`_out += _renderInclude(${i});
`;else if(n==="inject"){const[s,l]=i.split(",").map(o=>o.trim().replace(/['"]/g,""));t+=`var ${s} = this.$services['${l}'];
`}else t+=`_out += '@${n}';
`}else if(c.type==="Component"){const n=c.tagName.substring(2),i=m(c.attrs||""),e={};let s=[];for(const o of c.children)if(o.type==="Component"&&o.tagName==="x-slot"){const a=o.attrs.match(/name=["'](.*?)["']/),p=a?a[1]:"default";e[p]=o.children.map(h).join("")}else s.push(o);s.length>0&&(e.default=s.map(h).join(""));let l="{";for(const[o,a]of Object.entries(e))l+=`
"${o}": (() => {
let _out='';
${a}
return _out;
}),`;l+="}",t+=`_out += _renderComponent('${n}', {${i}}, ${l});
`}return t}function $(c){return c.replace(/`/g,"\\`").replace(/\$/g,"\\$")}function m(c){const t=[],n=/(:)?([a-zA-Z0-9_-]+)=["'](.*?)["']/g;let i;for(;(i=n.exec(c))!==null;){const e=!!i[1],s=i[2],l=i[3];e?t.push(`"${s}": ${l}`):t.push(`"${s}": "${l}"`)}return t.join(", ")}exports.compile=E;
