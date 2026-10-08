---
name: IBM Plex Corporate
colors:
  surface: '#fbf9f8'
  surface-dim: '#dbd9d9'
  surface-bright: '#fbf9f8'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f5f3f3'
  surface-container: '#efeded'
  surface-container-high: '#eae8e7'
  surface-container-highest: '#e4e2e2'
  on-surface: '#1b1c1c'
  on-surface-variant: '#424656'
  inverse-surface: '#303030'
  inverse-on-surface: '#f2f0f0'
  outline: '#737687'
  outline-variant: '#c3c6d8'
  surface-tint: '#0052dd'
  primary: '#004ccd'
  on-primary: '#ffffff'
  primary-container: '#0f62fe'
  on-primary-container: '#f3f3ff'
  inverse-primary: '#b4c5ff'
  secondary: '#5e5e5e'
  on-secondary: '#ffffff'
  secondary-container: '#e1dfdf'
  on-secondary-container: '#626263'
  tertiary: '#006527'
  on-tertiary: '#ffffff'
  tertiary-container: '#198038'
  on-tertiary-container: '#d4ffd2'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#dbe1ff'
  primary-fixed-dim: '#b4c5ff'
  on-primary-fixed: '#00174c'
  on-primary-fixed-variant: '#003da9'
  secondary-fixed: '#e4e2e2'
  secondary-fixed-dim: '#c7c6c6'
  on-secondary-fixed: '#1b1c1c'
  on-secondary-fixed-variant: '#464747'
  tertiary-fixed: '#96f8a1'
  tertiary-fixed-dim: '#7bdb87'
  on-tertiary-fixed: '#002108'
  on-tertiary-fixed-variant: '#00531e'
  background: '#fbf9f8'
  on-background: '#1b1c1c'
  surface-variant: '#e4e2e2'
typography:
  headline-lg:
    fontFamily: IBM Plex Sans
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
  body-md:
    fontFamily: IBM Plex Sans
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  label-md:
    fontFamily: IBM Plex Sans
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  gutter: 1rem
  margin: 1.5rem
  space-xs: 0.25rem
  space-sm: 0.5rem
  space-md: 1rem
  space-lg: 1.5rem
  space-xl: 2rem
---

# Design System Document: IBM Plex Corporate

## Brand & Style
The design system adopts a **Corporate / Modern** style, heavily inspired by IBM's Design Language and enterprise-grade design principles. It emphasizes clarity, efficiency, reliability, and precision. The aesthetic is clean and professional, tailored for complex applications where usability and information density are paramount.

## Colors
The color palette is built using a semantic tonal spot approach optimized for enterprise usability.
- **Primary Color (`#0f62fe`):** A vibrant, authoritative blue that anchors interactive elements, primary actions, and key focal points.
- **Secondary Color (`#6f6f6f`):** A neutral gray utilized for supporting elements, borders, and secondary text hierarchies.
- **Tertiary Color (`#198038`):** A functional green reserved for success states, positive indicators, and confirmation actions.
- **Neutral Color (`#525252`):** Dark slate neutrals providing high contrast for text and structural interfaces in light mode.

## Typography
The typography relies entirely on **IBM Plex Sans**, bringing an open-source, neutral, and grooved neo-grotesque spirit that is engineered for legibility across digital interfaces.
- **Headlines:** Set in medium-to-semibold weights to establish clear hierarchical structure.
- **Body:** Optimized at comfortable reading sizes (14px base) with robust line heights.
- **Labels:** Crisp and compact to support dense data displays and form controls.

## Layout & Spacing
A structured grid and spacing rhythm based on a standard multiplier scale ensure consistent layout alignment. 
- **Grid:** Fluid layout model using 16px (1rem) gutters and 24px (1.5rem) outer margins.
- **Spacing Scale:** Utilizes a standard scale (`space-xs` through `space-xl`) to maintain predictable spatial relationships between dense enterprise components.

## Elevation & Depth
Elevation is handled via clean, low-contrast outlines and subtle tonal layers. Rather than heavy drop shadows, the system relies on structured borders and background surface shifts to separate content hierarchies, ensuring a flat yet deeply readable enterprise user experience.

## Shapes
The shape language uses a **Soft** roundedness level (`roundedness: 1`), providing slight corner radiuses (0.25rem for base elements, up to 0.75rem for larger containers) that soften the strict corporate aesthetic without losing geometric precision.

## Components
- **Buttons:** Rectangular with subtle rounding (0.25rem). Primary buttons use the vibrant blue (`#0f62fe`), while secondary buttons utilize neutral outlines.
- **Inputs:** Clean border outlines with high-contrast text and clear focus states.
- **Cards & Containers:** Structured surfaces defined by light neutral borders and soft corner radii.
- **Chips & Badges:** Compact pill or softly rounded tags indicating metadata or status using the tertiary green or neutral tones.