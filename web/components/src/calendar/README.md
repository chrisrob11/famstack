# FamStack Daily Calendar Component - Development Plan

## Project Overview
Build a Web Component (using Lit) for a family-focused daily calendar view that displays on a shared horizontal tablet. Focus: glanceable "heads up" view of today's events for the whole family.

**Target Timeline:** 20-25 sessions over 6-8 weeks
**Development Strategy:** Build in `/calendar-dev` experiment page, then integrate into main app

## Data Model Review (Prerequisite)
**Status:** ❌ Not Started

**Before Milestone 1:** Review and validate calendar event data structure
- [ ] Audit current API response format from calendar integrations
- [ ] Ensure we have: `id`, `title`, `start_time`, `end_time`, `all_day`, `attendees`, `source`
- [ ] Add missing fields if needed (location, description, color hints)
- [ ] Create test data sets for development
- [ ] Validate family member data includes color/initial info

---

## Development Milestones

### Milestone 1: Foundation Setup (Sessions 1-2)
**Status:** ✅ Complete
**Calendar Support Level:** Empty container with "Hello World"
**User Validation:** Navigate to `/calendar-dev`, see basic component rendered

- [x] Install and configure Lit in existing TypeScript build
- [x] Create basic `<daily-calendar>` component structure
- [x] Create `/calendar-dev` experiment page
- [x] Set up development workflow and hot reload
- [x] Basic component registration and HTML integration

### Milestone 2: Core Layout & Time Grid (Sessions 3-4)
**Status:** ✅ Complete
  **Calendar Support Level:** Empty time grid with hour markings, no events
  **User Validation:** See clean 24-hour time grid with hour labels, scrolled to a reasonable morning hour

- [x] Implement a full 24-hour time grid (15-min blocks), with the view auto-scrolling to 7 AM on initial load
- [x] Add hour line separators and time labels- [x] Responsive horizontal tablet layout
- [x] CSS structure with custom properties

### Milestone 3: Event Data & API Integration (Sessions 5-6)
**Status:** ✅ Complete
**Calendar Support Level:** Loads real/mock event data, no visual events yet
**User Validation:** API call returns event data with attendees

- [x] Review/update calendar event data model
- [x] Create mock test data for development
- [x] Integrate with calendar API endpoint
- [x] Handle authentication and error states
- [x] Process events into all-day vs. timed

### Milestone 4: Basic Event Rendering (Sessions 7-8)
**Status:** ❌ Not Started
**Calendar Support Level:** Simple events - title only, single color, stacked if overlapping
**User Validation:** See events appear in correct time slots, all-day events at top

- [ ] Create `<event-card>` component
- [ ] Render events as simple blocks in correct time slots
- [ ] All-day events section at top
- [ ] Basic positioning (no overlap handling)

### Milestone 5: Person Identification System (Sessions 9-10)
**Status:** ❌ Not Started
**Calendar Support Level:** Events with colored person indicators, multi-person support
**User Validation:** Each event shows person circles, colors are consistent per person

- [ ] Family member color coding system
- [ ] Add person circles/dots with initials to events
- [ ] Multi-person events show multiple circles
- [ ] Consistent color per family member

### Milestone 6: Event Overlap & Positioning (Sessions 11-13)
**Status:** ❌ Not Started
**Calendar Support Level:** Sophisticated overlap handling, events don't stack
**User Validation:** Overlapping events appear side-by-side, not on top of each other

- [ ] Detect overlapping events
- [ ] Side-by-side layout for concurrent events
- [ ] Dynamic width calculation
- [ ] Smart positioning algorithms

### Milestone 7: Person Filtering (Sessions 14-15)
**Status:** ❌ Not Started
**Calendar Support Level:** Full filtering, shows individual or family view
**User Validation:** Click person circle → see only their events, click again → see all

- [ ] One-tap person filter functionality
- [ ] Show/hide based on selected person
- [ ] Visual feedback for active filters
- [ ] "Show All" toggle

### Milestone 8: Polish & Tablet Optimization (Sessions 16-18)
**Status:** ❌ Not Started
**Calendar Support Level:** Production-ready daily view, tablet-optimized
**User Validation:** Smooth performance on tablet, visually polished

- [ ] Horizontal tablet layout refinements
- [ ] Touch-friendly targets and interactions
- [ ] Visual polish (shadows, spacing, typography)
- [ ] Performance optimization

### Milestone 9: Integration & Real Data Testing (Sessions 19-20)
**Status:** ❌ Not Started
**Calendar Support Level:** Full integration with live calendar data
**User Validation:** Works in main app with real Google/Apple/Microsoft calendar events

- [ ] Move from experiment page to main app
- [ ] Real calendar integration testing
- [ ] Multi-family member testing
- [ ] Cross-browser compatibility

### Milestone 10: MVP Production Ready (Sessions 21-22)
**Status:** ❌ Not Started
**Calendar Support Level:** Family-ready daily dashboard
**User Validation:** Family can use it daily as primary calendar view

- [ ] Final UX polish and edge cases
- [ ] Error handling and resilience
- [ ] Ready for family daily use
- [ ] Performance monitoring

---

## Calendar Support Evolution

| Milestone | Visual | Overlap Handling | Person Support | Filtering |
|-----------|---------|------------------|----------------|-----------|
| 1-2 | Empty grid | N/A | N/A | N/A |
| 3-4 | Simple blocks | Stacked | None | None |
| 5 | With person dots | Stacked | Colored circles | None |
| 6 | Smart layout | Side-by-side | Multi-person events | None |
| 7+ | Full featured | Dynamic positioning | Full color coding | Interactive |

---

## Technical Architecture

### Component Structure
```
<daily-calendar>
├── <all-day-section>
├── <time-grid>
│   ├── <time-labels>
│   └── <events-container>
│       └── <event-card> (multiple)
└── <person-filter-bar>
```

### Key Data Interfaces
```typescript
interface CalendarEvent {
  id: string;
  title: string;
  start_time: string;
  end_time: string;
  all_day: boolean;
  attendees: string[]; // family member IDs
  source: 'google' | 'apple' | 'microsoft';
}

interface FamilyMember {
  id: string;
  name: string;
  initial: string;
  color: string;
}
```

### File Structure
```
web/components/src/calendar/
├── README.md                   (this file)
├── daily-calendar.ts           (main component)
├── components/
│   ├── all-day-section.ts
│   ├── time-grid.ts
│   ├── event-card.ts
│   └── person-filter.ts
├── services/
│   └── calendar-api.ts
├── utils/
│   ├── time-utils.ts
│   ├── layout-utils.ts
│   └── color-utils.ts
└── styles/
    └── calendar-styles.ts
```

---

## Progress Tracking

**Current Status:** Milestone 3 Complete - Data Integration
**Next Session:** Start Milestone 4 - Basic Event Rendering
**Sessions Completed:** 2/25
**Estimated Completion:** 6-8 weeks

### Session Log
- **Session 0 (Planning):** Created development plan and roadmap.
- **Session 1 (Foundation Setup):** ✅ Complete
  - Integrated Lit framework with TypeScript build system.
  - Created basic `<daily-calendar>` Web Component.
  - Set up `/calendar-dev` experiment page.
- **Session 2 (Data & Refactoring):** ✅ Complete
  - Implemented API data fetching for calendar events, including attendees.
  - Created a dedicated `CalendarApiService` for separation of concerns.
  - Added a reusable database seed script for test data.
  - Refactored component to improve i18n and remove magic numbers.
  - Enhanced accessibility by adding ARIA roles to the calendar grid.
  - Updated `README.md` to reflect design improvements.

---

*Last Updated: September 22, 2025*
*This document is updated after each development session to track progress*