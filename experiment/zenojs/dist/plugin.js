import { c as i } from "./compiler-Dkyzkoo9.js";
function l() {
  return {
    name: "vite-plugin-zeno",
    transform(n, c) {
      if (!c.endsWith(".blade")) return;
      const e = n.match(/<script>([\s\S]*?)<\/script>/), r = n.match(/<template>([\s\S]*?)<\/template>/);
      if (!e || !r)
        return;
      const o = e[1].trim(), s = r[1].trim(), a = i(s);
      let t = o.replace(/export\s+default/, "const script =");
      return t += `

script.render = function() {
${a}
};
`, t += `
export default script;
`, {
        code: t,
        map: null
      };
    }
  };
}
export {
  l as default
};
