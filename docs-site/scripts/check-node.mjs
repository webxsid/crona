const required = [22, 12, 0];
const actual = process.versions.node.split(".").map(Number);

const supported = actual[0] > required[0] ||
  (actual[0] === required[0] && actual[1] > required[1]) ||
  (actual[0] === required[0] && actual[1] === required[1] && actual[2] >= required[2]);

if (!supported) {
  console.error(`Crona docs require Node.js >= ${required.join(".")}; found ${process.versions.node}.`);
  process.exit(1);
}
