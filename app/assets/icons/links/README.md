# Link icons

Software-link and contact icons are stored together in this directory. Technology
icons remain in the neighboring `technologies` directory.

## Existing SVG artwork

These files were extracted from the existing inline SVG artwork in
`app/assets/js/app.js`; they are not newly downloaded or newly designed logos:

- `github.svg`: the existing software-link GitHub mark, shared by software links
  and the profile contact link.
- `repository.svg`, `package.svg`, `container.svg`, `archive.svg`, `webpage.svg`:
  generic software-link icons, not site-specific brand logos.
- `email.svg`, `address.svg`: contact glyphs.
- `orcid.svg`, `scholar.svg`, `linkedin.svg`: existing contact artwork with its
  original colors.
- `orcid-monochrome.svg`: the existing ORCID mark with transparent letter cutouts
  for CV contact links that follow the text color.

Monochrome SVGs are rendered as CSS masks so their color continues to follow the
surrounding link. Color artwork on the main site is rendered with an `img`
element to retain its colors. Generic software-link SVGs include their original rounded stroke styling
so they also work as standalone files.

CV contact links render all icons as masks in the text color. The ORCID variant
preserves transparent letters for this use.

The existing artwork's provenance and rights are unchanged by this extraction.
Brand marks identify their respective services; extracting them does not grant
additional rights to those marks.

## food4Rhino

- File: `food4rhino.svg`
- Original source: <https://www.rhino3d.com/images/f4r_icon_01.svg>
- Official usage: the Food4Rhino link in the footer of <https://www.rhino3d.com/>.
- Retrieved: 2026-09-07
- Format: original SVG artwork; viewBox `0 0 147.2 116.7`; no embedded raster image.
- Appearance: dark-red rhinoceros/puzzle-piece mark on a transparent background,
  without a wordmark. The official paths, colors, and proportions are preserved.
- Purpose: identify links to the official food4Rhino website. The artwork remains the property of its respective owner; it is not covered by this project's code license.

Keep these icons local so rendering a link does not make a third-party image request.
