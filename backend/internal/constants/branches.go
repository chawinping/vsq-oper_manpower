package constants

// StandardBranchCodes contains the list of branch codes that must always be available in the system.
// These codes are hardcoded and should not be deleted or disabled.
var StandardBranchCodes = []string{
	"CPN",
	"CPN-LS",
	"CTR",
	"PNK",
	"CNK",
	"BNA",
	"CLP",
	"SQR",
	"BKP",
	"CMC",
	"CSA",
	"EMQ",
	"ESV",
	"GTW",
	"MGA",
	"MTA",
	"PRM",
	"RCT",
	"RST",
	"TMA",
	"MBA",
	"SCN",
	"CWG",
	"CRM",
	"CWT",
	"PSO",
	"RCP",
	"CRA",
	"CTW",
	"ONE",
	"DCP",
	"MNG",
	"TLR",
	"TLR-LS",
	"TLR-WN",
}

// IsStandardBranchCode checks if a branch code is in the standard list
func IsStandardBranchCode(code string) bool {
	for _, stdCode := range StandardBranchCodes {
		if stdCode == code {
			return true
		}
	}
	return false
}

// GetStandardBranchCodes returns a copy of the standard branch codes list
func GetStandardBranchCodes() []string {
	codes := make([]string, len(StandardBranchCodes))
	copy(codes, StandardBranchCodes)
	return codes
}

