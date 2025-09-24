const fs = require('fs');
const path = require('path');
const os = require('os');

// Silent mode detection
const silent = process.env.npm_config_loglevel === 'silent' ||
               process.env.CC_SWITCH_SKIP_POSTINSTALL === '1';

if (!silent) {
  console.log('🚀 Setting up CC-Switch for Claude Code...');
}

try {
  const platform = process.platform;
  const arch = process.arch;
  const homeDir = os.homedir();
  const claudeDir = path.join(homeDir, '.claude', 'cc-switch');

  // Create directory
  fs.mkdirSync(claudeDir, { recursive: true });

  // Determine platform key
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
    'win32-ia32': '@hobeeliu/cc-switch-win32-x64', // Use 64-bit for 32-bit
  };

  const packageName = packageMap[platformKey];
  if (!packageName) {
    if (!silent) {
      console.log(`Platform ${platformKey} not supported for auto-setup`);
    }
    process.exit(0);
  }

  const binaryName = platform === 'win32' ? 'cc-switch.exe' : 'cc-switch';
  const targetPath = path.join(claudeDir, binaryName);

  // Multiple path search strategies for different package managers
  const findBinaryPath = () => {
    const possiblePaths = [
      // npm/yarn: nested in node_modules
      path.join(__dirname, '..', 'node_modules', packageName, binaryName),
      // pnpm: try require.resolve first
      (() => {
        try {
          const packagePath = require.resolve(packageName + '/package.json');
          return path.join(path.dirname(packagePath), binaryName);
        } catch {
          return null;
        }
      })(),
      // pnpm: flat structure fallback with version detection
      (() => {
        const currentPath = __dirname;
        const pnpmMatch = currentPath.match(/(.+\.pnpm)[\\/]([^\\//]+)[\\/]/);
        if (pnpmMatch) {
          const pnpmRoot = pnpmMatch[1];
          const packageNameEncoded = packageName.replace('/', '+');

          try {
            // Try to find any version of the package
            const pnpmContents = fs.readdirSync(pnpmRoot);
            const packagePattern = new RegExp(`^${packageNameEncoded.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}@`);
            const matchingPackage = pnpmContents.find(dir => packagePattern.test(dir));

            if (matchingPackage) {
              return path.join(pnpmRoot, matchingPackage, 'node_modules', packageName, binaryName);
            }
          } catch {
            // Fallback to current behavior if directory reading fails
          }
        }
        return null;
      })()
    ].filter(p => p !== null);

    for (const testPath of possiblePaths) {
      if (fs.existsSync(testPath)) {
        return testPath;
      }
    }
    return null;
  };

  const sourcePath = findBinaryPath();
  if (!sourcePath) {
    if (!silent) {
      console.log('Binary package not installed, skipping Claude Code setup');
      console.log('The global cc-switch command will still work via npm');
    }
    process.exit(0);
  }

  // Copy or link the binary
  if (platform === 'win32') {
    // Windows: Copy file
    fs.copyFileSync(sourcePath, targetPath);
  } else {
    // Unix: Try hard link first, fallback to copy
    try {
      if (fs.existsSync(targetPath)) {
        fs.unlinkSync(targetPath);
      }
      fs.linkSync(sourcePath, targetPath);
    } catch {
      fs.copyFileSync(sourcePath, targetPath);
    }
    fs.chmodSync(targetPath, '755');
  }

  if (!silent) {
    console.log('✨ CC-Switch is ready for Claude Code!');
    console.log(`📍 Location: ${targetPath}`);
    console.log('🎉 You can now use: cc-switch --help');
    console.log('');
    console.log('Quick Start:');
    console.log('  cc-switch list                # List all configurations');
    console.log('  cc-switch new <name>          # Create new configuration');
    console.log('  cc-switch use <name>          # Switch to configuration');
    console.log('  cc-switch web                 # Web interface');
  }
} catch (error) {
  // Silent failure - don't break installation
  if (!silent) {
    console.log('Note: Could not auto-configure for Claude Code');
    console.log('The global cc-switch command will still work.');
    console.log('You can manually copy cc-switch to ~/.claude/cc-switch/ if needed');
  }
}