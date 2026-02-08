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
│   └── theme/           # Custom theme
├── public/              # Static assets
├── guide/               # User guides
├── reference/           # API reference
├── architecture/        # Architecture docs
├── development/         # Developer docs
├── examples/            # Examples
└── index.md             # Landing page
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
