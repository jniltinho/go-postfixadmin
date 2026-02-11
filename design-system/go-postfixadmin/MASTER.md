# Design System Master File

> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.
> If that file exists, its rules **override** this Master file.
> If not, strictly follow the rules below.

---

**Project:** Go-Postfixadmin
**Generated:** 2026-02-11 17:33:56
**Category:** SaaS (General)

---

## Global Rules

### Color Palette

| Role | Hex | CSS Variable |
|------|-----|--------------|
| Primary | `#3B82F6` | `--color-primary` |
| Secondary | `#60A5FA` | `--color-secondary` |
| CTA/Accent | `#F97316` | `--color-cta` |
| Background | `#F8FAFC` | `--color-background` |
| Text | `#1E293B` | `--color-text` |

**Color Notes:** Cool→Hot gradients + neutral grey

### Typography

- **Heading Font:** Fira Code
- **Body Font:** Fira Sans
- **Mood:** dashboard, data, analytics, code, technical, precise, square, brutalist-lite
- **Google Fonts:** [Fira Code + Fira Sans](https://fonts.google.com/share?selection.family=Fira+Code:wght@400;500;600;700|Fira+Sans:wght@300;400;500;600;700)

**CSS Import:**
```css
@import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap');
```

### Spacing Variables

| Token | Value | Usage |
|-------|-------|-------|
| `--space-xs` | `4px` / `0.25rem` | Tight gaps |
| `--space-sm` | `8px` / `0.5rem` | Icon gaps, inline spacing |
| `--space-md` | `16px` / `1rem` | Standard padding |
| `--space-lg` | `24px` / `1.5rem` | Section padding |
| `--space-xl` | `32px` / `2rem` | Large gaps |
| `--space-2xl` | `48px` / `3rem` | Section margins |
| `--space-3xl` | `64px` / `4rem` | Hero padding |

### Shadow Depths & Borders

| Level | Value | Usage |
|-------|-------|-------|
| `--shadow-sm` | `1px 1px 0px rgba(0,0,0,0.1)` | Thin flat shadow |
| `--shadow-md` | `2px 2px 0px rgba(0,0,0,0.1)` | Standard flat shadow |
| `--border-radius` | `0px` | Square corners |

---

## Component Specs

### Buttons

```css
/* Primary Button */
.btn-primary {
  background: #F97316;
  color: white;
  padding: 12px 24px;
  border: 2px solid #1E293B;
  border-radius: 0px;
  font-weight: 600;
  transition: all 200ms ease;
  cursor: pointer;
  box-shadow: 4px 4px 0px #1E293B;
}

.btn-primary:hover {
  transform: translate(-2px, -2px);
  box-shadow: 6px 6px 0px #1E293B;
}

/* Secondary Button */
.btn-secondary {
  background: transparent;
  color: #3B82F6;
  border: 2px solid #1E293B;
  padding: 12px 24px;
  border-radius: 0px;
  font-weight: 600;
  transition: all 200ms ease;
  cursor: pointer;
}
```

### Cards

```css
.card {
  background: white;
  border: 2px solid #1E293B;
  border-radius: 0px;
  padding: 24px;
  box-shadow: 4px 4px 0px #1E293B;
  transition: all 200ms ease;
}

.card:hover {
  transform: translate(-2px, -2px);
  box-shadow: 6px 6px 0px #1E293B;
}
```

### Inputs

```css
.input {
  padding: 12px 16px;
  border: 2px solid #1E293B;
  border-radius: 0px;
  font-size: 16px;
  transition: all 200ms ease;
}

.input:focus {
  border-color: #3B82F6;
  outline: none;
  box-shadow: 4px 4px 0px #3B82F6;
}
```

---

## Style Guidelines

**Style:** Neobrutalism / Square-Modern

**Keywords:** Sharp edges, heavy borders, flat shadows, vibrant accents, technical typography, grid-based, high contrast

**Best For:** Modern SaaS, financial dashboards, high-end corporate, lifestyle apps, modal overlays, navigation

**Key Effects:** Zero border-radius, 2px solid strokes, offset-0 shadows, high contrast text

### Page Pattern

**Pattern Name:** Waitlist/Coming Soon

- **Conversion Strategy:** Scarcity + exclusivity. Show waitlist count. Early access benefits. Referral program.
- **CTA Placement:** Email form prominent (above fold) + Sticky form on scroll
- **Section Order:** 1. Hero with countdown, 2. Product teaser/preview, 3. Email capture form, 4. Social proof (waitlist count)

---

## Anti-Patterns (Do NOT Use)

- ❌ Excessive animation
- ❌ Dark mode by default

### Additional Forbidden Patterns

- ❌ **Emojis as icons** — Use SVG icons (Heroicons, Lucide, Simple Icons)
- ❌ **Missing cursor:pointer** — All clickable elements must have cursor:pointer
- ❌ **Layout-shifting hovers** — Avoid scale transforms that shift layout
- ❌ **Low contrast text** — Maintain 4.5:1 minimum contrast ratio
- ❌ **Instant state changes** — Always use transitions (150-300ms)
- ❌ **Invisible focus states** — Focus states must be visible for a11y

---

## Pre-Delivery Checklist

Before delivering any UI code, verify:

- [ ] No emojis used as icons (use SVG instead)
- [ ] All icons from consistent icon set (Heroicons/Lucide)
- [ ] `cursor-pointer` on all clickable elements
- [ ] Hover states with smooth transitions (150-300ms)
- [ ] Light mode: text contrast 4.5:1 minimum
- [ ] Focus states visible for keyboard navigation
- [ ] `prefers-reduced-motion` respected
- [ ] Responsive: 375px, 768px, 1024px, 1440px
- [ ] No content hidden behind fixed navbars
- [ ] No horizontal scroll on mobile
