# Kassie Documentation

This directory contains the VitePress documentation for Kassie.

## Development

Install dependencies:
```bash
cd docs
npm install
```

Run development server:
```bash
npm run dev
```

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## Structure

```
docs/
├── .vitepress/          # VitePress configuration
│   ├── config.ts        # Main config
│   ├── components/      # Vue components
│   │   └── VersionInfo.vue  # Dynamic version display
│   └── theme/           # Custom theme
├── scripts/             # Build scripts
│   └── generate-version.js  # Version info generator
├── public/              # Static assets
├── guide/               # User guides
├── reference/           # API reference
├── architecture/        # Architecture docs
├── development/         # Developer docs
├── examples/            # Examples
├── version.json         # Generated version info (gitignored)
└── index.md             # Landing page
```

## Dynamic Versioning

The documentation uses dynamic versioning that automatically pulls version information from git:

**How it works:**

1. `scripts/generate-version.js` reads git tags and commit info
2. Generates `version.json` with current version, commit hash, and build date
3. VitePress config imports and uses this data
4. `<VersionInfo />` component displays version dynamically in docs

**Usage in markdown:**

```markdown
<VersionInfo />
```

This displays:
```
Kassie v0.1.1
Commit: 00c76c0
Built: February 9, 2026
```

**Version generation runs automatically:**

- Before `npm run dev` (via `predev` script)
- Before `npm run build` (via `prebuild` script)

**Manual regeneration:**

```bash
node scripts/generate-version.js
```

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch in the `docs/` directory.

See `.github/workflows/deploy-docs.yml` for deployment configuration.

## Contributing

When adding new pages:

1. Create the markdown file in the appropriate section
2. Add it to the sidebar in `.vitepress/config.ts`
3. Follow the existing structure and style
4. Test locally with `npm run dev`

## Style Guide

- Use clear, concise language
- Include code examples where applicable
- Add navigation links at the end of pages
- Use appropriate markdown formatting
- Include screenshots when helpful (place in `public/screenshots/`)

## Links

- [VitePress Documentation](https://vitepress.dev/)
- [Markdown Extensions](https://vitepress.dev/guide/markdown)
- [Default Theme Config](https://vitepress.dev/reference/default-theme-config)
