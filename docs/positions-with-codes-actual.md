# Positions with Actual Position Codes from Database

**Last Updated:** 2026-01-18  
**Source:** Database query

---

## All Positions with Position Codes

### Branch-Type Positions (for quota configuration)

| Position Code | Position Name (Thai) | English Translation | Display Order |
|---------------|---------------------|---------------------|---------------|
| B-MGR | ผู้จัดการสาขา | Branch Manager | 1 |
| B-AMGR | รองผู้จัดการสาขา | Assistant Branch Manager | 2 |
| B-AMGR-T | รองผู้จัดการสาขาและล่าม | Assistant Branch Manager & Interpreter | 3 |
| B-FR3 | ผู้ช่วยผู้จัดการสาขา | Assistant Branch Manager (variant) | 4 |
| B-CCO | ผู้ประสานงานคลินิก (Clinic Coordination Officer) | Clinic Coordination Officer | 5 |
| B-FRLS | พนักงานต้อนรับ (Laser Receptionist) | Laser Receptionist | 6 |
| B-FRPC | พนักงานต้อนรับ (Pico Laser Receptionist) | Pico Laser Receptionist | 7 |
| B-DRA | ผู้ช่วยแพทย์ | Doctor Assistant | 20 |
| *(not set)* | ผู้ช่วยแพทย์ Pico | Doctor Assistant Pico | 20 |
| B-APC | ผู้ช่วยแพทย์ Pico Laser | Doctor Assistant Pico Laser | 20 |
| B-NUR | พยาบาล | Nurse | 30 |
| B-ALS | ผู้ช่วย Laser Specialist | Laser Specialist Assistant | 999 |
| B-HKP | แม่บ้านประจำสาขา | Branch Housekeeper | 999 |

### Rotation-Type Positions (not for quota configuration)

| Position Code | Position Name (Thai) | English Translation | Display Order |
|---------------|---------------------|---------------------|---------------|
| AMGR | ผู้จัดการเขต | District Manager | 10 |
| DMGR | ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา | Department Manager & Branch Development Supervisor | 10 |
| R-DRA | ผู้ช่วยแพทย์วนสาขา | Doctor Assistant Rotation | 20 |
| DRAMGR | หัวหน้าผู้ช่วยแพทย์ | Head Doctor Assistant | 20 |
| R-FR-T | Front+ล่ามวนสาขา | Front + Interpreter Rotation | 999 |
| R-DRA-S | ผู้ช่วยพิเศษ | Special Assistant | 999 |
| R-FR | ฟร้อนท์วนสาขา | Front Rotation | 999 |

---

## CSV Format (for easy copy-paste)

```
Position Code,Position Name (Thai),Position Type,Display Order
B-MGR,ผู้จัดการสาขา,branch,1
B-AMGR,รองผู้จัดการสาขา,branch,2
B-AMGR-T,รองผู้จัดการสาขาและล่าม,branch,3
B-FR3,ผู้ช่วยผู้จัดการสาขา,branch,4
B-CCO,ผู้ประสานงานคลินิก (Clinic Coordination Officer),branch,5
B-FRLS,พนักงานต้อนรับ (Laser Receptionist),branch,6
B-FRPC,พนักงานต้อนรับ (Pico Laser Receptionist),branch,7
B-DRA,ผู้ช่วยแพทย์,branch,20
,ผู้ช่วยแพทย์ Pico,branch,20
B-APC,ผู้ช่วยแพทย์ Pico Laser,branch,20
B-NUR,พยาบาล,branch,30
B-ALS,ผู้ช่วย Laser Specialist,branch,999
B-HKP,แม่บ้านประจำสาขา,branch,999
AMGR,ผู้จัดการเขต,rotation,10
DMGR,ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา,rotation,10
R-DRA,ผู้ช่วยแพทย์วนสาขา,rotation,20
DRAMGR,หัวหน้าผู้ช่วยแพทย์,rotation,20
R-FR-T,Front+ล่ามวนสาขา,rotation,999
R-DRA-S,ผู้ช่วยพิเศษ,rotation,999
R-FR,ฟร้อนท์วนสาขา,rotation,999
```

---

## Positions Missing Position Codes

⚠️ **1 position needs a position code set:**

- **ผู้ช่วยแพทย์ Pico** (Doctor Assistant Pico) - ID: `10000000-0000-0000-0000-000000000020`

Please set the position code in the Position Management page (`/positions`).

---

## Excel Import Format for Position Quotas

When importing position quotas, use these codes in Column B:

**Example:**
```
TMA,B-MGR,1,1
TMA,B-AMGR,1,0
TMA,B-DRA,2,2
TMA,B-NUR,2,1
CPN,B-MGR,1,1
CPN,B-DRA,3,2
```

**Note:** Only branch-type positions (codes starting with `B-`) can be used for quota configuration. Rotation positions (codes starting with `R-` or other codes) are not used for branch quotas.

---

## Quick Reference for Common Branch Positions

| Code | Position |
|------|----------|
| B-MGR | Branch Manager |
| B-AMGR | Assistant Branch Manager |
| B-DRA | Doctor Assistant |
| B-NUR | Nurse |
| B-CCO | Coordinator |
| B-FRLS | Laser Receptionist |
| B-FRPC | Pico Laser Receptionist |
| B-HKP | Branch Housekeeper |
