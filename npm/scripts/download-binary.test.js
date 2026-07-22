"use strict";

// Pure-function tests for the npm wrapper's platform branching. Without
// these, the `process.platform === "win32"` arms in download-binary.js
// regress silently — they're only exercised in the wild on actual Windows
// machines, never by darwin/linux CI. Run with `node`.

const { describe, test, afterEach } = require("node:test");
const assert = require("node:assert");
const {
  detectPlatform,
  archiveExtension,
  archiveName,
  binaryFilename,
} = require("./download-binary");

const origPlatform = process.platform;
const origArch = process.arch;

function setPlatform(platform, arch) {
  Object.defineProperty(process, "platform", { value: platform, configurable: true });
  Object.defineProperty(process, "arch", { value: arch, configurable: true });
}

afterEach(() => {
  setPlatform(origPlatform, origArch);
});

describe("detectPlatform", () => {
  test("maps darwin/x64 to darwin/amd64", () => {
    setPlatform("darwin", "x64");
    assert.deepStrictEqual(detectPlatform(), { os: "darwin", arch: "amd64" });
  });

  test("maps linux/arm64 to linux/arm64", () => {
    setPlatform("linux", "arm64");
    assert.deepStrictEqual(detectPlatform(), { os: "linux", arch: "arm64" });
  });

  test("maps win32/x64 to windows/amd64", () => {
    setPlatform("win32", "x64");
    assert.deepStrictEqual(detectPlatform(), { os: "windows", arch: "amd64" });
  });

  test("maps win32/arm64 to windows/arm64", () => {
    setPlatform("win32", "arm64");
    assert.deepStrictEqual(detectPlatform(), { os: "windows", arch: "arm64" });
  });

  test("rejects unknown platform with a clear error", () => {
    setPlatform("freebsd", "x64");
    assert.throws(() => detectPlatform(), /Unsupported platform: freebsd/);
  });

  test("rejects unknown arch with a clear error", () => {
    setPlatform("linux", "ia32");
    assert.throws(() => detectPlatform(), /Unsupported architecture: ia32/);
  });
});

describe("archiveExtension", () => {
  test("returns zip on Windows", () => {
    setPlatform("win32", "x64");
    assert.strictEqual(archiveExtension(), "zip");
  });

  test("returns tar.gz on darwin", () => {
    setPlatform("darwin", "arm64");
    assert.strictEqual(archiveExtension(), "tar.gz");
  });

  test("returns tar.gz on linux", () => {
    setPlatform("linux", "x64");
    assert.strictEqual(archiveExtension(), "tar.gz");
  });
});

describe("archiveName", () => {
  test("Windows release name uses .zip + windows/amd64", () => {
    setPlatform("win32", "x64");
    assert.strictEqual(archiveName("0.9.0"), "wuphf_0.9.0_windows_amd64.zip");
  });

  test("Windows arm64 release name uses .zip + windows/arm64", () => {
    setPlatform("win32", "arm64");
    assert.strictEqual(archiveName("1.2.3"), "wuphf_1.2.3_windows_arm64.zip");
  });

  test("darwin release name uses .tar.gz", () => {
    setPlatform("darwin", "arm64");
    assert.strictEqual(archiveName("0.9.0"), "wuphf_0.9.0_darwin_arm64.tar.gz");
  });

  test("linux release name uses .tar.gz", () => {
    setPlatform("linux", "x64");
    assert.strictEqual(archiveName("0.9.0"), "wuphf_0.9.0_linux_amd64.tar.gz");
  });
});

describe("binaryFilename", () => {
  test("returns wuphf.exe on Windows (CreateProcess requires the .exe suffix)", () => {
    setPlatform("win32", "x64");
    assert.strictEqual(binaryFilename(), "wuphf.exe");
  });

  test("returns bare wuphf on Unix", () => {
    setPlatform("darwin", "arm64");
    assert.strictEqual(binaryFilename(), "wuphf");
    setPlatform("linux", "x64");
    assert.strictEqual(binaryFilename(), "wuphf");
  });
});
