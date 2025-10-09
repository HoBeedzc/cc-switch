#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const os = require('os');

// 获取当前平台和架构信息
const platform = os.platform();
const arch = os.arch();

// 映射Node.js架构名称到Go架构名称
const archMap = {
  'x64': 'amd64',
  'arm64': 'arm64'
};

// 映射平台名称
const platformMap = {
  'darwin': 'darwin',
  'linux': 'linux',
  'win32': 'windows'
};

const goArch = archMap[arch] || 'amd64';
const goPlatform = platformMap[platform] || 'linux';

// 构建二进制文件名
const extension = platform === 'win32' ? '.exe' : '';
const binaryName = `cc-switch-${goPlatform}-${goArch}${extension}`;
const binaryPath = path.join(__dirname, binaryName);

// 如果找不到对应平台的二进制文件，尝试使用默认的
const fs = require('fs');
let finalBinaryPath = binaryPath;
if (!fs.existsSync(binaryPath)) {
  const defaultBinary = path.join(__dirname, 'cc-switch');
  if (fs.existsSync(defaultBinary)) {
    finalBinaryPath = defaultBinary;
  } else {
    console.error(`错误: 找不到适用于 ${platform}-${arch} 的二进制文件`);
    console.error(`尝试查找: ${binaryPath}`);
    process.exit(1);
  }
}

// 执行二进制文件
const child = spawn(finalBinaryPath, process.argv.slice(2), {
  stdio: 'inherit'
});

child.on('exit', (code) => {
  process.exit(code);
});

child.on('error', (err) => {
  console.error(`执行错误: ${err.message}`);
  process.exit(1);
});
