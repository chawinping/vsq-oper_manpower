# Allocation Scoring System UI - Two Alternative Designs

## Overview

The allocation system has been updated to use a **point-based scoring system** with three fixed scoring groups that use lexicographic ordering. Unlike the previous priority-based system where criteria could be reordered, the new system has a fixed priority order:

1. **Group 1** (Highest Priority): Position Quota - Minimum Shortage (negative points)
2. **Group 2** (Second Priority): Daily Staff Constraints - Minimum Shortage (negative points)
3. **Group 3** (Third Priority): Position Quota - Preferred Shortage (positive points)

The UI needs to be updated to reflect this new system and explain how the scoring works.

---

## Alternative 1: Information Dashboard Style

### Design Philosophy
Transform the configuration page into an informational dashboard that explains the scoring system. Since priorities are fixed, the focus shifts from configuration to understanding and transparency.

### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│  Allocation Scoring System Configuration                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ℹ️ The allocation system uses a point-based scoring system       │
│  with three fixed priority groups. Priorities cannot be changed    │
│  as they ensure critical staffing needs are always addressed first.│
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  SCORING GROUPS (Fixed Priority Order)                             │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ 🔴 GROUP 1: Position Quota - Minimum Shortage               │  │
│  │ ─────────────────────────────────────────────────────────── │  │
│  │ Priority: Highest (checked first)                           │  │
│  │ Scoring: -1 point per staff below minimum                   │  │
│  │                                                              │  │
│  │ Example:                                                     │  │
│  │   • Position needs 5 minimum, has 3 → -2 points            │  │
│  │   • Position needs 3 minimum, has 3 → 0 points             │  │
│  │                                                              │  │
│  │ Purpose: Identifies critical shortages where branches are   │  │
│  │           below minimum staffing requirements.              │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ 🟠 GROUP 2: Daily Staff Constraints - Minimum Shortage       │  │
│  │ ─────────────────────────────────────────────────────────── │  │
│  │ Priority: Second (checked after Group 1)                     │  │
│  │ Scoring: -1 point per staff group below minimum             │  │
│  │                                                              │  │
│  │ Example:                                                     │  │
│  │   • Staff Group "Nurses": Needs 3, has 1 → -2 points       │  │
│  │   • Staff Group "Managers": Needs 2, has 2 → 0 points       │  │
│  │                                                              │  │
│  │ Purpose: Identifies shortages in staff group-based         │  │
│  │           constraints (Daily Staff Constraints).            │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ 🟢 GROUP 3: Position Quota - Preferred Shortage             │  │
│  │ ─────────────────────────────────────────────────────────── │  │
│  │ Priority: Third (checked after Group 1 and Group 2)         │  │
│  │ Scoring: +1 point per staff below preferred                 │  │
│  │            (only if at/above minimum)                       │  │
│  │                                                              │  │
│  │ Example:                                                     │  │
│  │   • Position needs 5 preferred, has 3, minimum is 2        │  │
│  │     → +2 points (above minimum)                             │  │
│  │   • Position needs 4 preferred, has 1, minimum is 2        │  │
│  │     → 0 points (below minimum, handled by Group 1)          │  │
│  │                                                              │  │
│  │ Purpose: Identifies opportunities to reach preferred       │  │
│  │           staffing levels (nice-to-have, not critical).     │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  RANKING LOGIC                                                      │
│                                                                     │
│  The system ranks branch-position combinations using strict        │
│  lexicographic ordering:                                           │
│                                                                     │
│  1. Sort by Group 1 Score (ascending - more negative = higher)    │
│  2. If tied, sort by Group 2 Score (ascending - more negative)   │
│  3. If tied, sort by Group 3 Score (descending - more positive)  │
│  4. If still tied, sort by Branch Code (alphabetical)             │
│                                                                     │
│  Example Ranking:                                                  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ Rank │ Branch │ Position │ Group 1 │ Group 2 │ Group 3 │    │  │
│  ├─────────────────────────────────────────────────────────────┤  │
│  │  1   │  A01   │  Nurse   │   -5    │   -2    │    0    │    │  │
│  │  2   │  B02   │  Nurse   │   -3    │   -1    │   +1    │    │  │
│  │  3   │  C03   │  Nurse   │   -3    │    0    │   +2    │    │  │
│  │  4   │  D04   │  Nurse   │   -2    │   -3    │   +1    │    │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  KEY FEATURES                                                       │
│                                                                     │
│  ✓ Magnitude matters: 2 staff below minimum = -2 points           │
│  ✓ Separate display: All three groups shown independently          │
│  ✓ Fixed priorities: Cannot be changed (ensures critical needs    │
│                       are always addressed first)                  │
│  ✓ Deterministic: Always produces consistent ranking              │
│                                                                     │
│  [View Documentation]  [Test Scoring System]                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Structure

**Main Sections:**
1. **Header**: Title and explanation that priorities are fixed
2. **Scoring Groups**: Three expandable cards, one for each group
3. **Ranking Logic**: Visual explanation of how ranking works
4. **Example Table**: Shows example rankings
5. **Key Features**: Bullet points highlighting important aspects

**Visual Design:**
- Color-coded groups: Red (Group 1), Orange (Group 2), Green (Group 3)
- Expandable cards with detailed explanations
- Example calculations shown inline
- Visual ranking table with example data

### Advantages
- ✅ **Clear and educational**: Explains how the system works
- ✅ **Transparent**: Shows exactly how scores are calculated
- ✅ **No confusion**: Makes it clear priorities are fixed
- ✅ **Helpful examples**: Provides concrete examples
- ✅ **Professional**: Dashboard-style presentation

### Disadvantages
- ❌ **Less interactive**: No configuration options (but that's by design)
- ❌ **More verbose**: Takes more space
- ❌ **May feel static**: No drag-and-drop interaction

---

## Alternative 2: Visual Scoring Calculator Style

### Design Philosophy
Create an interactive, visual representation of the scoring system that allows users to see how different scenarios would be scored. Focus on understanding through visualization and examples.

### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│  Allocation Scoring System                                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  The allocation system uses a point-based scoring system with      │
│  three fixed priority groups. More negative scores = higher         │
│  priority (more urgent).                                            │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  SCORING GROUPS                                                     │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │              │  │              │  │              │           │
│  │   GROUP 1    │  │   GROUP 2    │  │   GROUP 3    │           │
│  │              │  │              │  │              │           │
│  │ 🔴 Priority 1│  │ 🟠 Priority 2│  │ 🟢 Priority 3│           │
│  │              │  │              │  │              │           │
│  │ Position     │  │ Daily        │  │ Position     │           │
│  │ Quota - Min  │  │ Constraints  │  │ Quota - Pref│           │
│  │              │  │ - Min        │  │              │           │
│  │              │  │              │  │              │           │
│  │ -1 pt per    │  │ -1 pt per    │  │ +1 pt per    │           │
│  │ shortage     │  │ shortage     │  │ shortage     │           │
│  │              │  │              │  │              │           │
│  │ [Details ▼]  │  │ [Details ▼]  │  │ [Details ▼]  │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  INTERACTIVE SCORING CALCULATOR                                     │
│                                                                     │
│  Try different scenarios to see how they score:                    │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ Scenario: Branch A - Nurse Position                         │  │
│  ├─────────────────────────────────────────────────────────────┤  │
│  │                                                              │  │
│  │ Position Quota Configuration:                               │  │
│  │   Minimum Required:  [5]  Current: [3]  → Shortage: 2      │  │
│  │   Preferred Quota:   [7]  Current: [3]  → Shortage: 4      │  │
│  │                                                              │  │
│  │ Daily Staff Constraints:                                     │  │
│  │   Staff Group "Nurses":                                      │  │
│  │   Minimum Required:  [3]  Actual: [1]  → Shortage: 2       │  │
│  │                                                              │  │
│  │ ─────────────────────────────────────────────────────────── │  │
│  │                                                              │  │
│  │ SCORES:                                                      │  │
│  │                                                              │  │
│  │ Group 1 (Position Quota - Min):                             │  │
│  │   Position "Nurse": 2 below minimum → -2 points            │  │
│  │   Total: -2 points                                          │  │
│  │                                                              │  │
│  │ Group 2 (Daily Constraints - Min):                          │  │
│  │   Staff Group "Nurses": 2 below minimum → -2 points         │  │
│  │   Total: -2 points                                          │  │
│  │                                                              │  │
│  │ Group 3 (Position Quota - Preferred):                       │  │
│  │   Position "Nurse": 4 below preferred, but below minimum    │  │
│  │   → 0 points (only counts if at/above minimum)              │  │
│  │   Total: 0 points                                           │  │
│  │                                                              │  │
│  │ ─────────────────────────────────────────────────────────── │  │
│  │                                                              │  │
│  │ FINAL SCORE:                                                │  │
│  │   Group 1: -2  │  Group 2: -2  │  Group 3: 0                │  │
│  │                                                              │  │
│  │ Ranking Priority: HIGH (Group 1 = -2, Group 2 = -2)        │  │
│  │                                                              │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  [Try Another Scenario]  [Reset Calculator]                       │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  RANKING EXAMPLES                                                   │
│                                                                     │
│  Compare how different branches rank:                              │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │ Branch │ Position │ G1  │ G2  │ G3  │ Rank │ Explanation    │  │
│  ├─────────────────────────────────────────────────────────────┤  │
│  │  A01   │  Nurse   │ -5  │ -2  │  0  │  1   │ Highest G1    │  │
│  │  B02   │  Nurse   │ -3  │ -1  │ +1  │  2   │ Lower G1      │  │
│  │  C03   │  Nurse   │ -3  │  0  │ +2  │  3   │ Same G1,      │  │
│  │        │          │     │     │     │      │ lower G2      │  │
│  │  D04   │  Nurse   │ -2  │ -3  │ +1  │  4   │ Lower G1      │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  Key: G1 = Group 1, G2 = Group 2, G3 = Group 3                    │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  HOW IT WORKS                                                       │
│                                                                     │
│  1. Calculate Group 1 score (Position Quota - Minimum)             │
│     → More negative = more urgent                                  │
│                                                                     │
│  2. If Group 1 scores are equal, check Group 2                     │
│     → More negative = more urgent                                  │
│                                                                     │
│  3. If Group 1 & 2 are equal, check Group 3                       │
│     → More positive = less urgent (nice-to-have)                   │
│                                                                     │
│  4. If all groups are equal, use Branch Code                       │
│     → Alphabetical order for consistency                            │
│                                                                     │
│  [View Full Documentation]                                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Structure

**Main Sections:**
1. **Header**: Brief explanation of the system
2. **Scoring Groups**: Three compact cards showing the groups side-by-side
3. **Interactive Calculator**: Input fields to try different scenarios
4. **Real-time Calculation**: Shows scores as user inputs change
5. **Ranking Examples**: Table showing how different scenarios rank
6. **How It Works**: Step-by-step explanation

**Visual Design:**
- Compact group cards in a row
- Interactive input fields with live calculation
- Visual score display with color coding
- Comparison table with explanations
- Step-by-step guide

### Advantages
- ✅ **Interactive**: Users can experiment with different scenarios
- ✅ **Visual**: Shows calculations in real-time
- ✅ **Educational**: Learn by doing
- ✅ **Compact**: More information in less space
- ✅ **Engaging**: More interactive than static display

### Disadvantages
- ❌ **More complex**: Requires more development effort
- ❌ **May be overwhelming**: Too much interactivity for some users
- ❌ **Requires data**: Calculator needs realistic examples

---

## Comparison Matrix

| Aspect | Alternative 1: Dashboard | Alternative 2: Calculator |
|--------|-------------------------|---------------------------|
| **Complexity** | Low | Medium-High |
| **Interactivity** | Low (read-only) | High (interactive) |
| **Educational Value** | High (clear explanations) | Very High (learn by doing) |
| **Space Usage** | More vertical space | More compact |
| **Development Effort** | Low | Medium |
| **User Engagement** | Medium | High |
| **Clarity** | Very High | High |
| **Best For** | Quick reference, documentation | Learning, experimentation |

---

## Recommendations

### For Maximum Clarity and Documentation
**Recommended: Alternative 1 (Dashboard Style)**
- Best for users who want to quickly understand the system
- Clear, professional presentation
- Easy to reference
- Lower development effort

### For Maximum Engagement and Learning
**Recommended: Alternative 2 (Calculator Style)**
- Best for users who want to experiment and understand deeply
- Interactive learning experience
- More engaging
- Helps users understand edge cases

### Hybrid Approach
Consider combining both:
- **Main view**: Dashboard style (Alternative 1) for quick reference
- **"Try It" section**: Calculator (Alternative 2) as an expandable section
- Best of both worlds: clarity + interactivity

---

## Implementation Considerations

### Common Elements (Both Alternatives)

1. **Remove drag-and-drop**: No longer needed since priorities are fixed
2. **Update API calls**: May need to update backend endpoints (or keep them for backward compatibility)
3. **Add examples**: Both designs benefit from concrete examples
4. **Visual indicators**: Color coding (red/orange/green) for the three groups
5. **Documentation links**: Link to detailed documentation

### Alternative 1 Specific
- Simple card-based layout
- Expandable sections for details
- Static example table
- Minimal JavaScript needed

### Alternative 2 Specific
- Interactive form inputs
- Real-time calculation logic
- State management for calculator
- More complex component structure

---

## Next Steps

1. Review both alternatives with stakeholders
2. Choose primary approach (or hybrid)
3. Create detailed component specifications
4. Design mockups/wireframes
5. Implement chosen design
6. Test with users
7. Iterate based on feedback
