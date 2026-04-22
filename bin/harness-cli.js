#!/usr/bin/env node

const { existsSync } = require("node:fs");
const { join } = require("node:path");
const { spawnSync } = require("node:child_process");

const executable = process.platform === "win32" ? "harness-cli.exe" : "harness-cli";
const binaryPath = join(__dirname, "..", "dist", executable);

if (!existsSync(binaryPath)) {
  console.error(
    "harness-cli native binary is missing. Reinstall @ifuryst/harness-cli with Go available on PATH.",
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
