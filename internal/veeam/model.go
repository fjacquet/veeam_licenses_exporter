package veeam

// licenseInfo is the subset of the Veeam Enterprise Manager GET /api/licensing
// response we consume. Field names/json tags are the documented Veeam licensing
// terms; they are NOT yet verified against a live Enterprise Manager, so this file
// is deliberately isolated — correcting a tag here must not require touching the
// parser. Absent fields decode to their zero value and are handled by the parser
// (absent-not-zero).
type licenseInfo struct {
	Edition                 string `json:"Edition"`
	Status                  string `json:"Status"`
	ExpirationDate          string `json:"ExpirationDate"` // RFC3339, e.g. "2027-01-31T00:00:00Z"
	LicensedInstancesNumber int    `json:"LicensedInstancesNumber"`
	UsedInstancesNumber     int    `json:"UsedInstancesNumber"`
}
