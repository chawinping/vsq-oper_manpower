# Positions with Position Codes

**Last Updated:** 2026-01-18

This document lists all positions with their suggested position codes for Excel import.

---

## All Positions with Position Codes

### Branch-Type Positions (for quota configuration)

| Position Code | Position Name (Thai) | English Translation | Position Type | Display Order |
|---------------|---------------------|---------------------|---------------|---------------|
| BM | ผู้จัดการสาขา | Branch Manager | branch | 1 |
| ABM | รองผู้จัดการสาขา | Assistant Branch Manager | branch | 2 |
| ABM2 | ผู้ช่วยผู้จัดการสาขา | Assistant Branch Manager (variant) | branch | 2 |
| ABM-INT | รองผู้จัดการสาขาและล่าม | Assistant Branch Manager & Interpreter | branch | 2 |
| DA | ผู้ช่วยแพทย์ | Doctor Assistant | branch | 20 |
| DA-PICO | ผู้ช่วยแพทย์ Pico | Doctor Assistant Pico | branch | 20 |
| DA-PICO-LASER | ผู้ช่วยแพทย์ Pico Laser | Doctor Assistant Pico Laser | branch | 20 |
| NUR | พยาบาล | Nurse | branch | 30 |
| COORD | ผู้ประสานงานคลินิก (Clinic Coordination Officer) | Clinic Coordination Officer | branch | 50 |
| LASER-ASST | ผู้ช่วย Laser Specialist | Laser Specialist Assistant | branch | 20 |
| RECEPT-LASER | พนักงานต้อนรับ (Laser Receptionist) | Laser Receptionist | branch | 40 |
| RECEPT-PICO | พนักงานต้อนรับ (Pico Laser Receptionist) | Pico Laser Receptionist | branch | 40 |
| HOUSEKEEPER | แม่บ้านประจำสาขา | Branch Housekeeper | branch | 60 |

### Rotation-Type Positions (not for quota configuration)

| Position Code | Position Name (Thai) | English Translation | Position Type | Display Order |
|---------------|---------------------|---------------------|---------------|---------------|
| FRONT-ROT | ฟร้อนท์วนสาขา | Front Rotation | rotation | 40 |
| FRONT-INT-ROT | Front+ล่ามวนสาขา | Front + Interpreter Rotation | rotation | 40 |
| DA-ROT | ผู้ช่วยแพทย์วนสาขา | Doctor Assistant Rotation | rotation | 20 |
| DA-HEAD | หัวหน้าผู้ช่วยแพทย์ | Head Doctor Assistant | rotation | 15 |
| DM | ผู้จัดการเขต | District Manager | rotation | 10 |
| DM-DEV | ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา | Department Manager & Branch Development Supervisor | rotation | 10 |
| SA | ผู้ช่วยพิเศษ | Special Assistant | rotation | 20 |

---

## Copy-Paste Format for Excel

### CSV Format (for easy copy-paste)

```
Position Code,Position Name (Thai),English Translation,Position Type
BM,ผู้จัดการสาขา,Branch Manager,branch
ABM,รองผู้จัดการสาขา,Assistant Branch Manager,branch
ABM2,ผู้ช่วยผู้จัดการสาขา,Assistant Branch Manager (variant),branch
ABM-INT,รองผู้จัดการสาขาและล่าม,Assistant Branch Manager & Interpreter,branch
DA,ผู้ช่วยแพทย์,Doctor Assistant,branch
DA-PICO,ผู้ช่วยแพทย์ Pico,Doctor Assistant Pico,branch
DA-PICO-LASER,ผู้ช่วยแพทย์ Pico Laser,Doctor Assistant Pico Laser,branch
NUR,พยาบาล,Nurse,branch
COORD,ผู้ประสานงานคลินิก (Clinic Coordination Officer),Clinic Coordination Officer,branch
LASER-ASST,ผู้ช่วย Laser Specialist,Laser Specialist Assistant,branch
RECEPT-LASER,พนักงานต้อนรับ (Laser Receptionist),Laser Receptionist,branch
RECEPT-PICO,พนักงานต้อนรับ (Pico Laser Receptionist),Pico Laser Receptionist,branch
HOUSEKEEPER,แม่บ้านประจำสาขา,Branch Housekeeper,branch
FRONT-ROT,ฟร้อนท์วนสาขา,Front Rotation,rotation
FRONT-INT-ROT,Front+ล่ามวนสาขา,Front + Interpreter Rotation,rotation
DA-ROT,ผู้ช่วยแพทย์วนสาขา,Doctor Assistant Rotation,rotation
DA-HEAD,หัวหน้าผู้ช่วยแพทย์,Head Doctor Assistant,rotation
DM,ผู้จัดการเขต,District Manager,rotation
DM-DEV,ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา,Department Manager & Branch Development Supervisor,rotation
SA,ผู้ช่วยพิเศษ,Special Assistant,rotation
```

---

## Notes

1. **Position codes are case-sensitive** - Use uppercase letters and hyphens as shown
2. **Branch-type positions** can have quotas configured via Excel import
3. **Rotation-type positions** cannot have branch quotas (they're for rotation staff)
4. **Position codes must be unique** - Each position should have a distinct code
5. **To set position codes**: Go to `/positions` page (Admin only) and edit each position to add the code

---

## Quick Reference for Common Positions

| Code | Position |
|------|----------|
| BM | Branch Manager |
| ABM | Assistant Branch Manager |
| DA | Doctor Assistant |
| NUR | Nurse |
| COORD | Coordinator |
| RECEPT-LASER | Laser Receptionist |
| RECEPT-PICO | Pico Laser Receptionist |
| HOUSEKEEPER | Branch Housekeeper |

---

## Excel Import Format for Position Quotas

When importing position quotas, use these codes in Column B:

```
Column A: Branch Code (e.g., TMA, CPN)
Column B: Position Code (use codes from above table)
Column C: Preferred No. (designated quota)
Column D: Minimum No. (minimum required)
```

**Example:**
```
TMA,BM,1,1
TMA,ABM,1,0
TMA,DA,2,2
TMA,NUR,2,1
CPN,BM,1,1
CPN,DA,3,2
```
