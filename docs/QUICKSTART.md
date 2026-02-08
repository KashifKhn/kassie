# Quick Start: Kassie Documentation

## Install and Preview

```bash
# Install dependencies
cd docs
npm install

# Start development server
npm run dev
```

Visit http://localhost:5173

## Deploy to GitHub Pages

### 1. Enable GitHub Pages
1. Go to repo Settings > Pages
2. Set Source to "GitHub Actions"

### 2. Push to Main
```bash
git add .
git commit -m "docs: add VitePress documentation"
git push origin main
```

The workflow will automatically build and deploy.

### 3. Access
Docs will be available at: `https://<username>.github.io/kassie/`

## Update Base URL (if needed)

If your repo name is not "kassie", update `.vitepress/config.ts`:

```typescript
export default defineConfig({
  base: '/your-repo-name/',
  // ...
})
```

## Commands

```bash
npm run dev      # Start dev server
npm run build    # Build for production
npm run preview  # Preview production build
```

## Next Steps

1. Add screenshots to `docs/public/screenshots/`
2. Replace logo at `docs/public/logo.svg`
3. Expand architecture/development/examples sections
4. Add actual product screenshots

See `DOCS_SETUP_COMPLETE.md` for detailed information.
