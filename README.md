# Runbooks design system

The shared design language for Runbooks: the token layer and the self-hosted
fonts, published so every repository (the app, the marketing site, anything
later) consumes the same values by **pinned version** rather than copying them.

## What's here

- `tokens.css`: the token layer. Colours for both themes, spacing, structural
  sizes, type scale, leading, radius, shadows, motion. Values only.
- `fonts.css`: the `@font-face` declarations for the self-hosted fonts.
- `fonts/`: Atkinson Hyperlegible Next + Mono (variable, latin; SIL OFL) and the
  Shade Mono subset of Noto Sans Mono (box drawing + shade blocks; SIL OFL), each
  beside its licence file.
- `brand/`: the generated brand assets: logomark, wordmark, horizontal and
  stacked lockups (light and dark), and the rounded icon, as text-outlined SVG
  plus 512/1024 PNG. `favicon.svg` is the icon, ready to serve as a site
  favicon. `social-card-on-dark` is the 1200x630 link preview (the OG image),
  dark-only because a preview crawler never themes it.
- `cmd/brand`: the generator (`mise run brand`). It reads `fonts/` and the
  palette in `tokens.css`, and writes `brand/` and `favicon.svg`. Needs
  `woff2_decompress`, `inkscape` and `rsvg-convert`; the output is committed.

The contract (what the tokens mean and how to use them) is in
[DESIGN.md](DESIGN.md). The full component contract lives with the app's
styleguide.

## Consuming it

The design language is distributed as **pinned release assets**. Fetch the tag you
want:

```
https://github.com/runbooks-help/design-system/releases/download/vX.Y.Z/tokens.css
https://github.com/runbooks-help/design-system/releases/download/vX.Y.Z/fonts.css
https://github.com/runbooks-help/design-system/releases/download/vX.Y.Z/fonts.tar.gz
https://github.com/runbooks-help/design-system/releases/download/vX.Y.Z/brand.tar.gz
https://github.com/runbooks-help/design-system/releases/download/vX.Y.Z/favicon.svg
```

`fonts.css` expects a `fonts/` directory beside it (the `fonts.tar.gz` unpacks to
one), so a consumer drops the three pieces together and links both stylesheets.
`brand.tar.gz` unpacks to `brand/` plus `favicon.svg`. Pin the tag; a floating
`latest` is not published.

## Licence

The token layer and `fonts.css` are FSL-1.1-MIT. The fonts are SIL OFL 1.1; see
`fonts/*.LICENSE`.
