"use strict";

// Pure-function tests for the cloudflared bundler's platform branching and
// manifest shape. The actual fetch + sha256 + extract path runs against the
// network on a real install, so these tests pin the offline-verifiable bits:
// platform mapping, target filename, and that the manifest covers every
// platform we claim to support.

const { describe, test, afterEach } = require("node:test");
const assert = require("node:assert");
const path = require("node:path");
const {
  detectManifestKey,
  downloadCloudflared,
  loadManifest,
  targetBinaryFilename,
} = require("./download-cloudflared");

const origPlatform = process.platform;
const origArch = process.arch;

function setPlatform(platform, arch) {
  Object.defineProperty(process, "platform", {
    value: platform,
    configurable: true,
  });
  Object.defineProperty(process, "arch", { value: arch, configurable: true });
}

afterEach(() => {
  setPlatform(origPlatform, origArch);
});

describe("detectManifestKey", () => {
  test("maps darwin/x64 -> darwin-amd64", () => {
    setPlatform("darwin", "x64");
    assert.strictEqual(detectManifestKey(), "darwin-amd64");
  });
  test("maps darwin/arm64 -> darwin-arm64", () => {
    setPlatform("darwin", "arm64");
    assert.strictEqual(detectManifestKey(), "darwin-arm64");
  });
  test("maps linux/x64 -> linux-amd64", () => {
    setPlatform("linux", "x64");
    assert.strictEqual(detectManifestKey(), "linux-amd64");
  });
  test("maps linux/arm64 -> linux-arm64", () => {
    setPlatform("linux", "arm64");
    assert.strictEqual(detectManifestKey(), "linux-arm64");
  });
  test("maps win32/x64 -> windows-amd64", () => {
    setPlatform("win32", "x64");
    assert.strictEqual(detectManifestKey(), "windows-amd64");
  });
  test("returns windows-arm64 key but manifest has no entry, so install is silently skipped", async () => {
    setPlatform("win32", "arm64");
    // cloudflared has no published windows-arm64 asset; the bundler skips
    // silently and the runtime surfaces a clear "missing" error if the
    // user clicks Start tunnel.
    assert.strictEqual(detectManifestKey(), "windows-arm64");
    // Sanity-check that the manifest does NOT have an entry for it, so
    // downloadCloudflared() falls through the "no entry" branch.
    const manifest = loadManifest();
    assert.strictEqual(manifest.platforms["windows-arm64"], undefined);
    // Direct assertion of the silent-skip contract: downloadCloudflared
    // returns null without touching the network. If a future refactor
    // turns this branch into an error or starts a fetch, the test fails
    // here instead of silently letting the regression through.
    assert.strictEqual(await downloadCloudflared({ silent: true }), null);
  });
});

describe("targetBinaryFilename", () => {
  test("appends .exe on win32", () => {
    setPlatform("win32", "x64");
    assert.strictEqual(targetBinaryFilename(), "cloudflared.exe");
  });
  test("bare name on linux", () => {
    setPlatform("linux", "x64");
    assert.strictEqual(targetBinaryFilename(), "cloudflared");
  });
});

describe("loadManifest", () => {
  test("pins a version + per-platform sha256", () => {
    const manifest = loadManifest();
    assert.ok(/^\d{4}\.\d+\.\d+$/.test(manifest.version));
    for (const key of [
      "darwin-amd64",
      "darwin-arm64",
      "linux-amd64",
      "linux-arm64",
      "windows-amd64",
    ]) {
      const entry = manifest.platforms[key];
      assert.ok(entry !== undefined);
      assert.ok(/^cloudflared-/.test(entry.asset));
      // Case-insensitive: download-cloudflared.js compares hashes after
      // .toLowerCase() on both sides, so an uppercase-hex manifest entry is
      // semantically valid; this assertion must match.
      assert.ok(/^[0-9a-f]{64}$/i.test(entry.sha256));
    }
  });

  test("manifest is included in the published package.json files array", () => {
    // download-cloudflared.js is published; the JSON sits next to it and
    // MUST be in package.json `files`. If it isn't, postinstall fails on
    // a freshly-installed npm package because require() blows up. Verify
    // the file is present in the checkout AND listed in the publish
    // manifest — fs.existsSync alone catches neither a packaging
    // regression nor a stale checkout.
    const fs = require("node:fs");
    const expected = path.join(__dirname, "cloudflared.json");
    assert.strictEqual(fs.existsSync(expected), true);
    const pkg = JSON.parse(
      fs.readFileSync(path.join(__dirname, "..", "package.json"), "utf8"),
    );
    assert.strictEqual(Array.isArray(pkg.files), true);
    assert.ok(pkg.files.includes("scripts/cloudflared.json"));
  });
});
