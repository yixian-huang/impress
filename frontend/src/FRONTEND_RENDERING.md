# Frontend Rendering Configuration Guide

## Overview

This document describes the configuration-driven rendering system for the **Inkless** frontend. All public pages and global sections fetch content from the backend CMS rather than hardcoded i18n keys.

**Status**: As of feature FE-108, all public pages have been migrated to config-driven rendering.

## Architecture

### Data Flow

```
Backend CMS (Published Config)
  ↓ (HTTP GET /public/content/{pageKey}?locale=zh|en)
Public Content API Client (src/api/publicContent.ts)
  ↓ (usePublicContent hook with auto-normalization)
React Components (Page-level)
  ↓ (TypeScript-typed config interfaces)
Rendered UI (with locale fallback)
```

### Key Components

1. **API Client** (`src/api/publicContent.ts`): Core HTTP client and normalization utilities
2. **React Hook** (`src/hooks/usePublicContent.ts`): Stateful hook for React components
3. **Page Components** (`src/pages/*/page.tsx`): Consumer components with typed config interfaces
4. **Global Components** (`src/components/feature/Header.tsx`, `Footer.tsx`): Config-driven navigation and footer

## Page Configuration Schema

### Supported Page Keys

All pages follow the schema defined in `docs/data-model.md`:

- `home` - Homepage with hero, about, advantages, core services
- `about` - Company profile and information blocks
- `advantages` - Competitive advantages list
- `core-services` - Service offerings detail
- `cases` - Case studies and success stories
- `experts` - Expert team profiles
- `contact` - Contact form and information
- `global` - Site-wide navigation and footer

### LocalizedText Contract

All bilingual text fields follow the `LocalizedText` interface:

```typescript
interface LocalizedText {
  zh: string;  // Chinese (required at publish time)
  en: string;  // English (required at publish time)
}
```

**Runtime Behavior**:
- Backend returns published config with both `zh` and `en` values (enforced at publish gate)
- Frontend normalization layer applies locale selection based on current i18n language
- If `en` value is empty/missing at runtime, system falls back to `zh` (defensive coding)
- Fallback does NOT mutate source data; it's applied during normalization only

## Locale Fallback Policy

As documented in `docs/data-model.md` section 7:

### Backend Guarantees
- Publish gate blocks releases when required `LocalizedText` fields have `missing` or `stale` status
- Both `zh` and `en` values must be present and synchronized for required fields

### Frontend Behavior
```typescript
// 1. Fetch config for current locale
const { config } = usePublicContent('home', {
  locale: i18n.language,  // 'zh' or 'en'
  autoNormalize: true     // Apply locale selection
});

// 2. Normalization applies getLocalizedText() to all LocalizedText objects
//    - Returns text[locale] if present
//    - Falls back to text.zh if locale === 'en' and en is missing
//    - Returns empty string if both missing (defensive)

// 3. Normalized config has string values instead of { zh, en } objects
const hero = config.hero as { title: string; subtitle: string };
return <h1>{hero.title}</h1>;  // Direct string access
```

### Fallback Rationale
- **Publish time**: Strict validation ensures bilingual completeness
- **Runtime**: Defensive fallback prevents UI breakage if data is corrupted or incomplete
- **Migration**: Allows gradual content population without blocking development

## Component Integration Patterns

### Basic Page Pattern

All migrated pages follow this structure:

```typescript
import { useTranslation } from 'react-i18next';
import { usePublicContent } from '@/hooks/usePublicContent';
import type { Locale } from '@/api/publicContent';

interface HeroConfig {
  title?: string;
  subtitle?: string;
  backgroundImage?: string;
}

interface PageConfig {
  hero?: HeroConfig;
  // ... other sections
}

export default function MyPage() {
  // 1. Get current locale from i18n context
  const { i18n } = useTranslation('common');
  const locale = (i18n.language === 'zh' || i18n.language.startsWith('zh')
    ? 'zh'
    : 'en') as Locale;

  // 2. Fetch and normalize config
  const { loading, error, config } = usePublicContent('my-page', {
    locale,
    autoNormalize: true,
  });

  // 3. Handle loading and error states
  if (loading) {
    return <div className="min-h-screen bg-white flex items-center justify-center">
      <div className="text-gray-600">Loading...</div>
    </div>;
  }

  if (error) {
    return <div className="min-h-screen bg-white flex items-center justify-center">
      <div className="text-red-600">Failed to load page content</div>
    </div>;
  }

  // 4. Extract typed config with defaults
  const pageConfig = (config as PageConfig) || {};
  const hero = pageConfig.hero || {};

  // 5. Render with conditional checks for optional fields
  return (
    <div className="min-h-screen bg-white">
      <Header />

      {hero.title && (
        <section className="hero">
          <h1>{hero.title}</h1>
          {hero.subtitle && <p>{hero.subtitle}</p>}
        </section>
      )}

      <Footer />
    </div>
  );
}
```

### Global Component Pattern

Header and Footer components fetch from the `global` pageKey:

```typescript
export default function Header() {
  const { i18n, changeLanguage } = useTranslation('common');
  const locale = (i18n.language === 'zh' || i18n.language.startsWith('zh')
    ? 'zh'
    : 'en') as Locale;

  const { config } = usePublicContent('global', {
    locale,
    autoNormalize: true,
  });

  const globalConfig = (config as GlobalConfig) || {};
  const nav = globalConfig.nav || {};
  const items = nav.items || [];

  return (
    <header>
      {/* Logo from config.branding.logo */}
      <nav>
        {items.map((item, idx) => (
          <a key={idx} href={item.href}>{item.label}</a>
        ))}
      </nav>
      <LanguageSwitcher />
    </header>
  );
}
```

## i18n Compatibility Layer

### Current Status
- `react-i18next` is **retained** for language switching mechanism and compatibility
- Page content i18n keys (`common.hero.title`, `common.about.description`, etc.) are **deprecated**
- Only `useTranslation()` for accessing `i18n.language` remains actively used in pages

### Migration Path
```typescript
// ❌ OLD: Hardcoded i18n keys (DEPRECATED)
const { t } = useTranslation('common');
return <h1>{t('hero.title')}</h1>;

// ✅ NEW: Config-driven with locale from i18n
const { i18n } = useTranslation('common');
const locale = i18n.language === 'zh' ? 'zh' : 'en';
const { config } = usePublicContent('home', { locale, autoNormalize: true });
return <h1>{config.hero.title}</h1>;
```

### Deprecated i18n Keys

The following i18n key groups in `src/i18n/local/{zh,en}/common.ts` are **obsolete** but retained for compatibility:

- `common.hero.*` - Replaced by `home` pageKey config
- `common.about.*` - Replaced by `home` and `about` pageKey configs
- `common.advantages.*` - Replaced by `home` and `advantages` pageKey configs
- `common.coreServices.*` - Replaced by `home` and `core-services` pageKey configs
- `common.coreServicesPage.*` - Replaced by `core-services` pageKey config
- `common.advantagesPage.*` - Replaced by `advantages` pageKey config
- `common.caseListPage.*` - Replaced by `cases` pageKey config
- `common.expertTeamPage.*` - Replaced by `experts` pageKey config
- `common.aboutPage.*` - Replaced by `about` pageKey config
- `common.contactPage.*` - Replaced by `contact` pageKey config
- `common.footer.*` - Replaced by `global` pageKey config (partially)

### Still Active i18n Keys

These keys remain in use and are NOT deprecated:

- `common.nav.*` - May still be referenced by legacy components (migration to `global` config in progress)
- Any admin-specific UI strings not managed by CMS

## Type Safety

### Page Config Interfaces

Each page defines TypeScript interfaces for type-safe config access:

```typescript
// Example: src/pages/home/page.tsx
interface HeroConfig {
  title?: string;
  subtitle?: string;
  backgroundImage?: string;
}

interface AboutConfig {
  title?: string;
  description?: string;
  description2?: string;
  description3?: string;
  button?: string;
  image?: string;
}

interface HomePageConfig {
  hero?: HeroConfig;
  about?: AboutConfig;
  advantages?: AdvantagesConfig;
  coreServices?: CoreServicesConfig;
}
```

**Key Properties**:
- All fields are `optional` (`?`) for graceful degradation
- String types represent normalized (locale-selected) values
- Array types for repeating sections (cards, services, etc.)

### Type Assertion Pattern

```typescript
const pageConfig = (config as HomePageConfig) || {};
const hero = pageConfig.hero || {};

// Safe access with fallback to empty string or default image
<img src={hero.backgroundImage || '/images/default-hero.png'} />
<h1>{hero.title || ''}</h1>
```

## Graceful Degradation

### Missing Optional Fields

All components implement conditional rendering for optional fields:

```typescript
// ✅ GOOD: Conditional rendering
{hero.subtitle && <p>{hero.subtitle}</p>}

// ✅ GOOD: Section-level conditional
{advantages.title && (
  <section>
    <h2>{advantages.title}</h2>
    {advantages.cards && advantages.cards.length > 0 && (
      <div className="grid">
        {advantages.cards.map((card, idx) => (...))}
      </div>
    )}
  </section>
)}

// ❌ BAD: No conditional check
<p>{hero.subtitle}</p>  // Breaks if undefined
```

### Image Fallbacks

Components provide fallback image paths when config images are missing:

```typescript
<img
  src={card.image || `/images/default-card-${index + 1}.png`}
  alt={card.title || `Card ${index + 1}`}
/>
```

### Loading and Error States

All pages implement consistent loading and error UI:

```typescript
if (loading) {
  return (
    <div className="min-h-screen bg-white flex items-center justify-center">
      <div className="text-gray-600">Loading...</div>
    </div>
  );
}

if (error) {
  return (
    <div className="min-h-screen bg-white flex items-center justify-center">
      <div className="text-red-600">Failed to load page content</div>
    </div>
  );
}
```

## Static Assets

### Current Approach
- Static images remain in `/public/images/` directory
- Config stores relative URLs: `/images/hero-bg.png`
- No signed upload or object storage in current phase

### URL Format
```typescript
interface MediaRef {
  url: string;           // e.g., "/images/hero-bg.png"
  alt: {
    zh: string;
    en: string;
  };
}
```

### Migration Notes
- Legacy hardcoded image references (e.g., `ABOUT_IMAGES`, `ADVANTAGE_IMAGES` constants) have been removed
- All image URLs now come from backend config or fallback to default paths
- Future: May migrate to object storage with CDN URLs in backend config

## Performance Considerations

### Caching Strategy
- Backend implements standard HTTP caching headers for `/public/content/*` endpoints
- Frontend does NOT implement client-side caching (relies on browser cache)
- SWR/React Query not currently integrated (future consideration)

### Bundle Size
- `usePublicContent` hook is lightweight (~2KB)
- Auto-normalization adds minimal overhead (~1KB)
- No significant impact on page load time vs. hardcoded content

## Error Handling

### API Error Structure
```typescript
interface PublicContentError {
  code: string;       // e.g., "NOT_FOUND", "NETWORK_ERROR"
  message: string;    // Human-readable error
  details?: unknown;  // Optional structured details
}
```

### Common Error Scenarios
1. **404 Not Found**: Page not published or invalid pageKey
2. **Network Error**: Backend unreachable or timeout
3. **Invalid Locale**: Unexpected locale parameter (backend returns 400)

### Error Recovery
- Components display generic error message to user
- No automatic retry mechanism (user must refresh)
- Errors logged to console for debugging

## Testing Considerations

### Component Testing
When testing pages, mock `usePublicContent`:

```typescript
import { vi } from 'vitest';

vi.mock('@/hooks/usePublicContent', () => ({
  usePublicContent: () => ({
    loading: false,
    error: null,
    config: {
      hero: { title: 'Test Title', subtitle: 'Test Subtitle' },
      // ... mock config
    },
  }),
}));
```

### Integration Testing
- Ensure backend CMS has seeded data for all pageKeys
- Test language switching triggers refetch with correct locale
- Verify fallback images display when config images missing

## Troubleshooting

### Issue: Page shows "Failed to load page content"
**Cause**: Backend API not running or pageKey not published

**Solution**:
1. Check backend server is running (`pnpm server` or Go binary)
2. Verify page is published via admin panel
3. Check browser console for API error details

### Issue: English content shows Chinese text
**Cause**: English locale missing in published config (fallback applied)

**Solution**:
1. Check admin panel for translation status (should be `done`, not `missing` or `stale`)
2. Re-publish page after ensuring both `zh` and `en` fields are complete
3. Verify backend seed data includes bilingual content

### Issue: Images not displaying
**Cause**: Config image URL incorrect or file missing

**Solution**:
1. Check image URL in admin editor (should start with `/images/`)
2. Verify image file exists in `/public/images/` directory
3. Confirm fallback image paths in component code

## Future Improvements

### Planned Enhancements
1. **Client-side caching**: Integrate SWR or React Query for better UX
2. **Incremental Static Generation**: Pre-render pages at build time with Vite SSG
3. **Image optimization**: Migrate to object storage with CDN and responsive images
4. **Admin preview**: Live preview in admin panel before publish
5. **A/B testing**: Support multiple config variants per page

### Deprecation Timeline
1. **Phase 1 (Current)**: Config-driven rendering active, i18n keys retained
2. **Phase 2**: Remove obsolete i18n keys after 3-month stability period
3. **Phase 3**: Migrate remaining admin UI strings to backend if needed

## References

- **Data Model**: `docs/data-model.md` - Page config schema and translation state
- **API Spec**: `docs/api-spec.md` - Backend API contracts
- **Architecture**: `docs/architecture.md` - Overall system design
- **Development Plan**: `docs/development-plan.md` - Migration roadmap
- **API Client Docs**: `src/api/README.md` - Usage examples and patterns
