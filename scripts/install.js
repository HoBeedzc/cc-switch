#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// 获取平台和架构信息
const platform = os.platform();
const arch = os.arch();

// 平台映射
const platformMap = {
  'darwin': 'darwin',
  'linux': 'linux',
  'win32': 'windows'
};

const archMap = {
  'x64': 'amd64',
  'arm64': 'arm64'
};

// 获取目标平台
const targetPlatform = platformMap[platform];
const targetArch = archMap[arch];

if (!targetPlatform || !targetArch) {
  console.error(`Unsupported platform: ${platform}-${arch}`);
  process.exit(1);
}

// 构建二进制文件名
const binaryName = platform === 'win32' ? 'cc-switch.exe' : 'cc-switch';
const sourceBinary = `cc-switch-${targetPlatform}-${targetArch}${platform === 'win32' ? '.exe' : ''}`;

// 路径设置
const binDir = path.join(__dirname, '..', 'bin');
const sourcePath = path.join(binDir, sourceBinary);
const targetPath = path.join(binDir, binaryName);

try {
  // 检查源文件是否存在
  if (!fs.existsSync(sourcePath)) {
    console.error(`Binary not found: ${sourcePath}`);
    console.error('This package may not support your platform.');
    process.exit(1);
  }

  // 复制并重命名二进制文件
  fs.copyFileSync(sourcePath, targetPath);

  // 设置执行权限 (Unix系统)
  if (platform !== 'win32') {
    fs.chmodSync(targetPath, 0o755);
  }

  console.log('✓ cc-switch installed successfully!');
  console.log('');
  console.log('Usage:');
  console.log('  cc-switch list                # List all configurations');
  console.log('  cc-switch new <name>          # Create new configuration');
  console.log('  cc-switch use <name>          # Switch to configuration');
  console.log('  cc-switch delete <name>       # Delete configuration');
  console.log('  cc-switch current             # Show current configuration');
  console.log('');
  console.log('Run "cc-switch --help" for more information.');

} catch (error) {
  console.error('Installation failed:', error.message);
  process.exit(1);
}