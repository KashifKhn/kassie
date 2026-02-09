<script setup lang="ts">
import versionInfo from '../../version.json'

const formatDate = (isoDate: string) => {
  return new Date(isoDate).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}
</script>

<template>
  <div class="version-info">
    <pre><code>Kassie {{ versionInfo.version }}
Commit: {{ versionInfo.commit }}
Built: {{ formatDate(versionInfo.buildDate) }}</code></pre>
  </div>
</template>

<style scoped>
.version-info {
  margin: 1em 0;
}

.version-info pre {
  background-color: var(--vp-code-block-bg);
  border-radius: 6px;
  padding: 1em;
  overflow-x: auto;
}

.version-info code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.875em;
  line-height: 1.7;
  color: var(--vp-c-text-1);
}
</style>
