const fs = require('fs');
const path = require('path');
const os = require('os');

// Silent mode detection
const silent = process.env.npm_config_loglevel === 'silent' ||
               process.env.CC_SWITCH_SKIP_PREUNINSTALL === '1';

// Check if full cleanup is requested
const fullCleanup = process.env.CC_SWITCH_FULL_CLEANUP === '1';

/**
 * Recursively remove a directory and all its contents
 */
function removeDirectory(dirPath) {
  if (!fs.existsSync(dirPath)) {
    return false;
  }

  const entries = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      removeDirectory(fullPath);
    } else {
      fs.unlinkSync(fullPath);
    }
  }
  fs.rmdirSync(dirPath);
  return true;
}

/**
 * Safely remove a file if it exists
 */
function removeFile(filePath) {
  if (fs.existsSync(filePath)) {
    fs.unlinkSync(filePath);
    return true;
  }
  return false;
}

try {
  const homeDir = os.homedir();
  const claudeDir = path.join(homeDir, '.claude');
  
  // cc-switch binary directory
  const ccSwitchBinDir = path.join(claudeDir, 'cc-switch');
  
  // cc-switch data directory (profiles)
  const profilesDir = path.join(claudeDir, 'profiles');
  
  // Files to clean up
  const cleanupItems = [];

  if (!silent) {
    console.log('🧹 Cleaning up CC-Switch...');
    console.log('');
  }

  // 1. Remove cc-switch binary directory
  if (fs.existsSync(ccSwitchBinDir)) {
    if (removeDirectory(ccSwitchBinDir)) {
      cleanupItems.push(`Binary directory: ${ccSwitchBinDir}`);
    }
  }

  // 2. Remove cc-switch data files (profiles, templates, etc.)
  if (fullCleanup && fs.existsSync(profilesDir)) {
    if (removeDirectory(profilesDir)) {
      cleanupItems.push(`Profiles directory: ${profilesDir}`);
    }
  } else if (fs.existsSync(profilesDir)) {
    // Partial cleanup: only remove cc-switch internal files, keep user profiles
    const internalFiles = [
      '.current',
      '.history', 
      '.empty_mode',
      '.empty_backup_settings.json'
    ];
    
    for (const file of internalFiles) {
      const filePath = path.join(profilesDir, file);
      if (removeFile(filePath)) {
        cleanupItems.push(`Internal file: ${filePath}`);
      }
    }

    // Remove templates directory (created by cc-switch)
    const templatesDir = path.join(profilesDir, 'templates');
    if (fs.existsSync(templatesDir)) {
      if (removeDirectory(templatesDir)) {
        cleanupItems.push(`Templates directory: ${templatesDir}`);
      }
    }

    // Try to remove profiles directory if empty
    try {
      const remaining = fs.readdirSync(profilesDir);
      if (remaining.length === 0) {
        fs.rmdirSync(profilesDir);
        cleanupItems.push(`Profiles directory (empty): ${profilesDir}`);
      }
    } catch {
      // Directory not empty or removal failed, that's okay
    }
  }

  // 3. Remove legacy files (from old versions)
  const legacyFiles = [
    path.join(claudeDir, '.current'),
    path.join(claudeDir, '.empty_mode')
  ];
  
  for (const legacyFile of legacyFiles) {
    if (removeFile(legacyFile)) {
      cleanupItems.push(`Legacy file: ${legacyFile}`);
    }
  }

  // Show cleanup summary
  if (!silent) {
    if (cleanupItems.length > 0) {
      console.log('✓ Cleaned up the following:');
      for (const item of cleanupItems) {
        console.log(`  - ${item}`);
      }
      console.log('');
    }

    // Check if there are remaining profile files
    if (!fullCleanup && fs.existsSync(profilesDir)) {
      const remaining = fs.readdirSync(profilesDir).filter(f => f.endsWith('.json'));
      if (remaining.length > 0) {
        console.log('ℹ️  User profiles were preserved:');
        console.log(`   ${profilesDir}`);
        console.log(`   Files: ${remaining.join(', ')}`);
        console.log('');
        console.log('   To remove all data, reinstall and uninstall with:');
        console.log('   CC_SWITCH_FULL_CLEANUP=1 npm uninstall @hobeeliu/cc-switch');
        console.log('');
        console.log('   Or manually remove:');
        if (process.platform === 'win32') {
          console.log(`   rmdir /s /q "${profilesDir}"`);
        } else {
          console.log(`   rm -rf "${profilesDir}"`);
        }
        console.log('');
      }
    }

    console.log('✨ CC-Switch cleanup complete!');
  }
} catch (error) {
  // Silent failure - don't break uninstallation
  if (!silent) {
    console.log('Note: Could not fully clean up CC-Switch files');
    console.log('You can manually remove: ~/.claude/cc-switch/ and ~/.claude/profiles/');
    if (process.env.DEBUG) {
      console.error('Error:', error.message);
    }
  }
}
