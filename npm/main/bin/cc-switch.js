#!/usr/bin/env node
const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

// 1. Priority: Use ~/.claude/cc-switch if exists
const claudePath = path.join(
  os.homedir(),
  '.claude',
  'cc-switch',
  process.platform === 'win32' ? 'cc-switch.exe' : 'cc-switch'
);

if (fs.existsSync(claudePath)) {
  const result = spawnSync(claudePath, process.argv.slice(2), {
    stdio: 'inherit',
    shell: false
  });
  process.exit(result.status || 0);
}

// 2. Fallback: Use npm package binary
const platform = process.platform;
const arch = process.arch;

// Map platform and architecture for package names
let platformKey = `${platform}-${arch}`;

// Handle architecture mapping
if (arch === 'ia32') {
  platformKey = `${platform}-x64`; // Use 64-bit for 32-bit systems
}

const packageMap = {
  'darwin-x64': '@hobeeliu/cc-switch-darwin-x64',
  'darwin-arm64': '@hobeeliu/cc-switch-darwin-arm64',
  'linux-x64': '@hobeeliu/cc-switch-linux-x64',
  'linux-arm64': '@hobeeliu/cc-switch-linux-arm64',
  'win32-x64': '@hobeeliu/cc-switch-win32-x64',
  'win32-ia32': '@hobeeliu/cc-switch-win32-x64', // Use 64-bit for 32-bit systems
};

const packageName = packageMap[platformKey];
if (!packageName) {
  console.error(`Error: Unsupported platform ${platformKey}`);
  console.error('Supported platforms: darwin (x64/arm64), linux (x64/arm64), win32 (x64)');
  console.error('Please visit https://github.com/HoBeedzc/cc-switch for manual installation');
  process.exit(1);
}

const binaryName = platform === 'win32' ? 'cc-switch.exe' : 'cc-switch';
const binaryPath = path.join(__dirname, '..', 'node_modules', packageName, binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(`Error: Binary not found at ${binaryPath}`);
  console.error('This might indicate a failed installation or unsupported platform.');
  console.error('Please try reinstalling: npm install -g @hobeeliu/cc-switch');
  console.error(`Expected package: ${packageName}`);
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  shell: false
});

process.exit(result.status || 0);