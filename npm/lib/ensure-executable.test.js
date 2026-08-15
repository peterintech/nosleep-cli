const assert = require("node:assert/strict");
const test = require("node:test");

const { ensureExecutable } = require("./ensure-executable");

test("repairs a non-executable macOS binary", () => {
  let accessCalls = 0;
  let chmodCall;
  const fileSystem = {
    constants: { X_OK: 1 },
    accessSync: () => {
      accessCalls += 1;
      if (accessCalls === 1) {
        const error = new Error("permission denied");
        error.code = "EACCES";
        throw error;
      }
    },
    chmodSync: (filePath, mode) => {
      chmodCall = { filePath, mode };
    }
  };

  const result = ensureExecutable("/tmp/nosleepp", "darwin", fileSystem);

  assert.deepEqual(result, { ok: true, changed: true });
  assert.deepEqual(chmodCall, { filePath: "/tmp/nosleepp", mode: 0o755 });
  assert.equal(accessCalls, 2);
});

test("repairs a non-executable Linux binary", () => {
  let accessCalls = 0;
  let chmodCall;
  const fileSystem = {
    constants: { X_OK: 1 },
    accessSync: () => {
      accessCalls += 1;
      if (accessCalls === 1) {
        const error = new Error("permission denied");
        error.code = "EACCES";
        throw error;
      }
    },
    chmodSync: (filePath, mode) => {
      chmodCall = { filePath, mode };
    }
  };

  const result = ensureExecutable("/tmp/nosleepp", "linux", fileSystem);

  assert.deepEqual(result, { ok: true, changed: true });
  assert.deepEqual(chmodCall, { filePath: "/tmp/nosleepp", mode: 0o755 });
  assert.equal(accessCalls, 2);
});

test("skips permission changes on Windows", () => {
  let touched = false;
  const fileSystem = {
    constants: { X_OK: 1 },
    accessSync: () => {
      touched = true;
    },
    chmodSync: () => {
      touched = true;
    }
  };

  const result = ensureExecutable("C:\\nosleepp.exe", "win32", fileSystem);

  assert.deepEqual(result, { ok: true, changed: false });
  assert.equal(touched, false);
});

test("returns an actionable error when repair fails", () => {
  const fileSystem = {
    constants: { X_OK: 1 },
    accessSync: () => {
      const error = new Error("permission denied");
      error.code = "EACCES";
      throw error;
    },
    chmodSync: () => {
      const error = new Error("operation not permitted");
      error.code = "EPERM";
      throw error;
    }
  };

  const result = ensureExecutable("/global/nosleepp/bin/nosleepp-darwin-arm64", "darwin", fileSystem);

  assert.equal(result.ok, false);
  assert.match(result.error, /Unable to execute nosleepp binary/);
  assert.match(result.error, /chmod \+x/);
});
