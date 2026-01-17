# Position Cleanup Summary

**Date:** 2025-01-08  
**Action:** Remove English positions and keep only Thai positions

---

## Overview

This migration removes all English position entries from the database and migrates all related data (staff, quotas, suggestions) to their Thai equivalents.

---

## Position Mappings

| English Position (ID) | Thai Position (ID) | Thai Name |
|----------------------|-------------------|-----------|
| Branch Manager (001) | ผู้จัดการสาขา (008) | ผู้จัดการสาขา |
| Assistant Branch Manager (002) | รองผู้จัดการสาขา (009) | รองผู้จัดการสาขา |
| Service Consultant (003) | ผู้ประสานงานคลินิก (011) | ผู้ประสานงานคลินิก (Clinic Coordination Officer) |
| Coordinator (004) | ผู้ประสานงานคลินิก (011) | ผู้ประสานงานคลินิก (Clinic Coordination Officer) |
| Doctor Assistant (005) | ผู้ช่วยแพทย์ (010) | ผู้ช่วยแพทย์ |
| Physiotherapist (006) | ผู้ช่วยแพทย์ (010) | ผู้ช่วยแพทย์ (closest match) |
| Nurse (007) | พยาบาล (015) | พยาบาล |
| Front 3 (028) | ฟร้อนท์วนสาขา (027) | ฟร้อนท์วนสาขา |
| Front Laser (029) | พนักงานต้อนรับ (Laser Receptionist) (013) | พนักงานต้อนรับ (Laser Receptionist) |
| Laser Assistant (030) | ผู้ช่วย Laser Specialist (012) | ผู้ช่วย Laser Specialist |

---

## Migration Steps

The migration script (`migrate_remove_english_positions.go`) performs the following steps:

1. **Update Staff Records**: Migrates all staff records from English positions to Thai positions
2. **Update Position Quotas**: Migrates position quotas, handling conflicts by keeping Thai quotas
3. **Update Allocation Suggestions**: Migrates allocation suggestions to Thai positions
4. **Update Staff Allocation Rules**: Migrates allocation rules, handling conflicts
5. **Delete English Positions**: Removes all English position entries from the database

---

## Position Mappings for Removed Positions

The following positions are mapped to existing Thai positions:

- **Front 3** → **ฟร้อนท์วนสาขา** (Front Rotation)
- **Front Laser** → **พนักงานต้อนรับ (Laser Receptionist)**

Note: The Thai versions "ฟร้อนท์ 3" and "ฟร้อนท์ Laser" are also removed and mapped to the above positions.

---

## Remaining Positions (Thai Only)

After migration, the following 20 Thai positions remain:

1. ผู้จัดการสาขา
2. รองผู้จัดการสาขา
3. ผู้ช่วยแพทย์
4. ผู้ประสานงานคลินิก (Clinic Coordination Officer)
5. ผู้ช่วย Laser Specialist
6. พนักงานต้อนรับ (Laser Receptionist)
7. แม่บ้านประจำสาขา
8. พยาบาล
9. ผู้ช่วยผู้จัดการสาขา
10. ผู้ช่วยแพทย์ Pico Laser
11. รองผู้จัดการสาขาและล่าม
12. Front+ล่ามวนสาขา
13. ผู้ช่วยแพทย์ Pico
14. พนักงานต้อนรับ (Pico Laser Receptionist)
15. ผู้จัดการเขต
16. ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา
17. หัวหน้าผู้ช่วยแพทย์
18. ผู้ช่วยพิเศษ
19. ผู้ช่วยแพทย์วนสาขา
20. ฟร้อนท์วนสาขา

---

## Frontend Updates

The `BranchPositionQuotaConfig` component has been updated to use Thai position names:

- Updated `targetPositions` array to use Thai position names
- Matching logic remains fuzzy, so it will find Thai positions correctly

---

## Notes

1. **Physiotherapist Mapping**: Physiotherapist is mapped to "ผู้ช่วยแพทย์" as there's no direct Thai equivalent. This may need review.

2. **Service Consultant Mapping**: Service Consultant is mapped to "ผู้ประสานงานคลินิก" as the closest match.

3. **Conflict Handling**: If position quotas exist for both English and Thai positions for the same branch, the Thai quota is kept and the English one is deleted.

4. **Data Integrity**: All foreign key relationships are maintained during migration.

---

## Testing Recommendations

Before running in production:

1. **Backup Database**: Create a full database backup
2. **Test Migration**: Run migration on a test database first
3. **Verify Staff Records**: Check that all staff have valid Thai position assignments
4. **Verify Quotas**: Check that position quotas are correctly migrated
5. **Verify Suggestions**: Check that allocation suggestions reference Thai positions
6. **Test Frontend**: Verify that Branch Configuration UI works with Thai positions

---

## Rollback Plan

If rollback is needed:

1. Restore from database backup
2. Or manually recreate English positions with original IDs
3. Update staff/quota/suggestion records back to English positions

---

## Files Modified

- `backend/internal/repositories/postgres/migrations.go` - Updated seed data
- `backend/internal/repositories/postgres/migrate_remove_english_positions.go` - Migration script
- `frontend/src/components/branch/BranchPositionQuotaConfig.tsx` - Updated to use Thai names

---

## Execution

The migration runs automatically when `RunMigrations()` is called, after all table creation and seed data insertion.

To run manually:
```go
err := postgres.MigrateRemoveEnglishPositions(db)
```
