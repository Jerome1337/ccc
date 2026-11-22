#!/usr/bin/env node

const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const readline = require('node:readline/promises');

const YAML = require('yaml');

const isGlobal = process.argv.includes('--global') || process.argv.includes('-g');
const PACKAGE_ROOT = path.resolve(__dirname, '..');
const TARGET_DIR = isGlobal ? path.join(os.homedir(), '.claude') : process.cwd();
const CLAUDE_DIR = isGlobal ? '' : '.claude';

const CONFIG_FILE = 'ccc.yml';

const FILES = [
  { src: 'CLAUDE.md', dest: 'CLAUDE.md' },
  { src: 'AGENTS.md', dest: 'AGENTS.md', scope: 'project' },
  { src: 'RTK.md', dest: 'RTK.md' },
  { src: 'CODING_STYLE.md', dest: 'CODING_STYLE.md' },
  { src: CONFIG_FILE, dest: CONFIG_FILE, scope: 'global' },
].filter((file) => file.scope !== (isGlobal ? 'project' : 'global'));

const RULES_DIR = 'rules';
const RULES_DEST = path.join(CLAUDE_DIR, 'rules');

const SKILLS_DIR = 'skills';
const SKILLS_DEST = path.join(CLAUDE_DIR, 'skills');

const MCP_ADD = ['mcp', 'add', '-s', 'user'];

const args = process.argv.slice(2);
const command = args.find((a) => !a.startsWith('-')) || 'init';
const force = args.includes('--force') || args.includes('-f');

const bold = (s) => `\x1b[1m${s}\x1b[0m`;
const green = (s) => `\x1b[32m${s}\x1b[0m`;
const yellow = (s) => `\x1b[33m${s}\x1b[0m`;
const dim = (s) => `\x1b[2m${s}\x1b[0m`;

const confirmOverwrite = async (relativePath) => {
  // readline exits silently on closed stdin, so only prompt on a terminal
  if (!process.stdin.isTTY) {
    return false;
  }

  const prompt = readline.createInterface({ input: process.stdin, output: process.stdout });
  try {
    const answer = await prompt.question(`  ${yellow('diff')}  ${relativePath} ${dim('overwrite? (y/N)')} `);

    return answer.trim().toLowerCase() === 'y';
  } catch {
    console.log();
    process.exit(1);
  } finally {
    prompt.close();
  }
};

const getSkipReason = async (srcPath, destPath) => {
  if (!fs.existsSync(destPath) || force) {
    return null;
  }

  if (command !== 'update') {
    return 'already exists, use --force to overwrite';
  }

  if (fs.readFileSync(srcPath).equals(fs.readFileSync(destPath))) {
    return 'up to date';
  }

  if (await confirmOverwrite(path.relative(TARGET_DIR, destPath))) {
    return null;
  }

  return 'changed, use --force to overwrite';
};

const copyFile = async (srcPath, destPath) => {
  const destDir = path.dirname(destPath);
  if (!fs.existsSync(destDir)) {
    fs.mkdirSync(destDir, { recursive: true });
  }

  const skipReason = await getSkipReason(srcPath, destPath);
  if (skipReason) {
    console.log(`  ${yellow('skip')}  ${path.relative(TARGET_DIR, destPath)} ${dim(`(${skipReason})`)}`);
    return 'skipped';
  }

  fs.copyFileSync(srcPath, destPath);
  console.log(`  ${green('copy')}  ${path.relative(TARGET_DIR, destPath)}`);
  return 'copied';
};

const copyDirectory = async (srcDir, destDir) => {
  let copied = 0;
  let skipped = 0;

  const srcPath = path.join(PACKAGE_ROOT, srcDir);
  if (fs.existsSync(srcPath)) {
    const mdFiles = fs.readdirSync(srcPath).filter((f) => f.endsWith('.md'));
    for (const file of mdFiles) {
      const result = await copyFile(path.join(srcPath, file), path.join(TARGET_DIR, destDir, file));
      if (result === 'copied') {
        copied++;
      } else {
        skipped++;
      }
    }
  }

  return { copied, skipped };
};

const isText = (value) => typeof value === 'string' && value.trim() !== '';

const toList = (value) => [value ?? []].flat().filter(isText);

const readConfig = (configPath) => {
  try {
    return YAML.parse(fs.readFileSync(configPath, 'utf8')) ?? {};
  } catch (error) {
    console.log(`  ${yellow('skip')}  ${CONFIG_FILE} ${dim(`(${error.message.split('\n')[0]})`)}\n`);

    return null;
  }
};

const toMcpCommand = (name, server) => {
  if (isText(server?.url)) {
    return [...MCP_ADD, '--transport', 'http', name, server.url];
  }

  if (!isText(server?.command)) {
    console.log(`  ${yellow('skip')}  ${name} ${dim('(needs a url or a command)')}`);

    return null;
  }

  const envKeys = toList(server.env);
  const missingKeys = envKeys.filter((key) => !process.env[key]);
  if (missingKeys.length > 0) {
    console.log(`  ${yellow('skip')}  ${name} ${dim(`(set ${missingKeys.join(', ')} and run again)`)}`);

    return null;
  }

  const env = envKeys.flatMap((key) => ['-e', `${key}=${process.env[key]}`]);

  return [...MCP_ADD, name, ...env, '--', ...server.command.trim().split(/\s+/)];
};

const installClaudeExtensions = () => {
  const configPath = path.join(TARGET_DIR, CONFIG_FILE);
  console.log(`${bold('MCP servers and plugins')} ${dim(`from ${configPath}`)}\n`);

  const config = readConfig(configPath);
  if (!config) {
    return;
  }

  const commands = [
    ...Object.entries(config.mcp ?? {}).map(([name, server]) => toMcpCommand(name, server)),
    ...toList(config.marketplaces).map((source) => ['plugin', 'marketplace', 'add', source]),
    ...toList(config.plugins).map((plugin) => ['plugin', 'install', plugin]),
  ].filter(Boolean);

  for (const claudeArgs of commands) {
    const result = spawnSync('claude', claudeArgs, { stdio: 'inherit' });
    if (result.error) {
      console.log(`  ${yellow('skip')}  ${dim('claude not found, install Claude Code first')}\n`);

      return;
    }
  }

  console.log();
};

const init = async () => {
  console.log(`\n${bold('@jerome1337/ccc')} — Setting up Claude Code rules\n`);

  let copied = 0;
  let skipped = 0;

  for (const file of FILES) {
    const result = await copyFile(path.join(PACKAGE_ROOT, file.src), path.join(TARGET_DIR, file.dest));
    if (result === 'copied') {
      copied++;
    } else {
      skipped++;
    }
  }

  const rulesResult = await copyDirectory(RULES_DIR, RULES_DEST);
  copied += rulesResult.copied;
  skipped += rulesResult.skipped;

  const skillsResult = await copyDirectory(SKILLS_DIR, SKILLS_DEST);
  copied += skillsResult.copied;
  skipped += skillsResult.skipped;

  console.log(`\n  ${bold(`${copied} copied`)}${skipped > 0 ? `, ${skipped} skipped` : ''}\n`);

  if (copied > 0) {
    console.log(`${dim('  Edit the files to match your project, then commit them.')}\n`);
  }

  if (isGlobal) {
    installClaudeExtensions();
  }
};

if (command === 'init' || command === 'update') {
  init();
} else {
  console.log(`\nUsage: ${bold('ccc init|update')} [--force] [--global]\n`);
  console.log(`  ${bold('init')}      Copy Claude Code rules into the current project`);
  console.log(`  ${bold('update')}    Add new files and ask before overwriting changed ones`);
  console.log(`  ${bold('--force')}   Overwrite existing files`);
  console.log(`  ${bold('--global')}  Use ~/.claude and install MCP servers and plugins\n`);
}
