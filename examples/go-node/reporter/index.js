function report(value) {
  const n = Number(value);
  return { ok: true, doubled: n * 2, source: "node" };
}

function ping() {
  return "ok";
}

function countUp(n) {
  const count = Number(n) || 0;
  const out = [];
  for (let i = 0; i < count; i++) out.push(i);
  return out;
}

module.exports = { report, ping, countUp };
