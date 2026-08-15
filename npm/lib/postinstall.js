const { resolveBinary } = require("./resolve-binary");
const { ensureExecutable } = require("./ensure-executable");

if (process.platform === "darwin" || process.platform === "linux") {
  const resolved = resolveBinary(process.platform, process.arch, __dirname + "/..");
  if (!resolved.ok) {
    console.error(resolved.error);
    process.exitCode = 1;
  } else {
    const permission = ensureExecutable(resolved.path);
    if (!permission.ok) {
      console.error(permission.error);
      process.exitCode = 1;
    }
  }
}
