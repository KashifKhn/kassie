#!/usr/bin/env node

import { execSync } from 'child_process'
import { writeFileSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

function getLatestTag() {
  try {
    const tag = execSync('git describe --tags --abbrev=0', { encoding: 'utf-8' }).trim()
    return tag
  } catch (error) {
    console.warn('No tags found, using v0.0.0')
    return 'v0.0.0'
  }
}

function getCommitHash() {
  try {
    const hash = execSync('git rev-parse --short HEAD', { encoding: 'utf-8' }).trim()
    return hash
  } catch (error) {
    return 'unknown'
  }
}

function getBuildDate() {
  return new Date().toISOString()
}

function getCommitCount() {
  try {
    const count = execSync('git rev-list --count HEAD', { encoding: 'utf-8' }).trim()
    return parseInt(count, 10)
  } catch (error) {
    return 0
  }
}

function main() {
  const version = getLatestTag()
  const commit = getCommitHash()
  const buildDate = getBuildDate()
  const commitCount = getCommitCount()

  const versionInfo = {
    version,
    commit,
    buildDate,
    commitCount,
    generatedAt: new Date().toISOString()
  }

  const outputPath = join(__dirname, '..', 'version.json')
  writeFileSync(outputPath, JSON.stringify(versionInfo, null, 2))

  console.log('Generated version.json:')
  console.log(JSON.stringify(versionInfo, null, 2))
}

main()
