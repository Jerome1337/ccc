#!/usr/bin/env node

const fs = require('node:fs')
const path = require('node:path')

const PACKAGE_ROOT = path.resolve(__dirname, '..')
const TARGET_DIR = process.cwd()

const FILES = [
  { src: 'CLAUDE.md', dest: 'CLAUDE.md' },
  { src: 'AGENTS.md', dest: 'AGENTS.md' },
  { src: 'RTK.md', dest: 'RTK.md' },
]

const RULES_DIR = 'rules'
const RULES_DEST = path.join('.claude', 'rules')

const args = process.argv.slice(2)
const command = args.find((a) => !a.startsWith('-')) || 'init'
const force = args.includes('--force') || args.includes('-f')

const bold = (s) => `\x1b[1m${s}\x1b[0m`
const green = (s) => `\x1b[32m${s}\x1b[0m`
const yellow = (s) => `\x1b[33m${s}\x1b[0m`
const dim = (s) => `\x1b[2m${s}\x1b[0m`

const copyFile = (srcPath, destPath) => {
  const destDir = path.dirname(destPath)
  if (!fs.existsSync(destDir)) {
    fs.mkdirSync(destDir, { recursive: true })
  }

  if (fs.existsSync(destPath) && !force) {
    console.log(`  ${yellow('skip')}  ${path.relative(TARGET_DIR, destPath)} ${dim('(already exists, use --force to overwrite)')}`)
    return 'skipped'
  }

  fs.copyFileSync(srcPath, destPath)
  console.log(`  ${green('copy')}  ${path.relative(TARGET_DIR, destPath)}`)
  return 'copied'
}

const init = () => {
  console.log(`\n${bold('@jerome1337/ccc')} — Setting up Claude Code rules\n`)

  let copied = 0
  let skipped = 0

  for (const file of FILES) {
    const result = copyFile(
      path.join(PACKAGE_ROOT, file.src),
      path.join(TARGET_DIR, file.dest),
    )
    if (result === 'copied') copied++
    else skipped++
  }

  const rulesSrcDir = path.join(PACKAGE_ROOT, RULES_DIR)
  if (fs.existsSync(rulesSrcDir)) {
    const ruleFiles = fs.readdirSync(rulesSrcDir).filter((f) => f.endsWith('.md'))
    for (const file of ruleFiles) {
      const result = copyFile(
        path.join(rulesSrcDir, file),
        path.join(TARGET_DIR, RULES_DEST, file),
      )
      if (result === 'copied') copied++
      else skipped++
    }
  }

  console.log(`\n  ${bold(`${copied} copied`)}${skipped > 0 ? `, ${skipped} skipped` : ''}\n`)

  if (copied > 0) {
    console.log(`${dim('  Edit the files to match your project, then commit them.')}\n`)
  }
}

if (command === 'init') {
  init()
} else {
  console.log(`\nUsage: ${bold('ccc init')} [--force]\n`)
  console.log(`  ${bold('init')}     Copy Claude Code rules into the current project`)
  console.log(`  ${bold('--force')}  Overwrite existing files\n`)
}
