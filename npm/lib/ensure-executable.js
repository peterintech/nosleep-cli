const fs = require("node:fs");

function ensureExecutable(binaryPath, platform = process.platform, fileSystem = fs) {
  if (platform !== "darwin") {
    return { ok: true, changed: false };
  }

  try {
    fileSystem.accessSync(binaryPath, fileSystem.constants.X_OK);
    return { ok: true, changed: false };
  } catch (accessError) {
    try {
      fileSystem.chmodSync(binaryPath, 0o755);
      fileSystem.accessSync(binaryPath, fileSystem.constants.X_OK);
      return { ok: true, changed: true };
    } catch (repairError) {
      return {
        ok: false,
        error: [
          `Unable to execute nosleepp binary: ${binaryPath}`,
          `macOS returned ${repairError.code || accessError.code || "a permission error"}.`,
          `Try running: chmod +x "${binaryPath}"`
        ].join(" ")
      };
    }
  }
}

module.exports = { ensureExecutable };
