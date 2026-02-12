"use strict";const p=require("./compiler-DaDr0jon.cjs");function a(){return{name:"vite-plugin-zeno",transform(e,c){if(!c.endsWith(".blade"))return;const n=e.match(/<script>([\s\S]*?)<\/script>/),r=e.match(/<template>([\s\S]*?)<\/template>/);if(!n||!r)return;const s=n[1].trim(),i=r[1].trim(),o=p.compile(i);let t=s.replace(/export\s+default/,"const script =");return t+=`

script.render = function() {
${o}
};
`,t+=`
export default script;
`,{code:t,map:null}}}}module.exports=a;
