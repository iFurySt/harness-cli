#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { mkdirSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const rootDir = dirname(dirname(fileURLToPath(import.meta.url)));
const version = process.env.npm_package_version || "dev";
const executable = process.platform === "win32" ? "harness-cli.exe" : "harness-cli";
const distDir = join(rootDir, "dist");
const outputPath = join(distDir, executable);

const goVersion = spawnSync("go", ["version"], { stdio: "ignore" });
if (goVersion.status !== 0) {
  console.error("Go is required to install @ifuryst/harness-cli. Install Go and rerun npm install.");
  process.exit(goVersion.status ?? 1);
}

rmSync(distDir, { recursive: true, force: true });
mkdirSync(distDir, { recursive: true });

const result = spawnSync(
  "go",
  [
    "build",
    "-trimpath",
    "-ldflags",
    `-s -w -X main.version=${version}`,
    "-o",
    outputPath,
    ".",
  ],
  {
    cwd: rootDir,
    stdio: "inherit",
  },
);

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
