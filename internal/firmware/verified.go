package firmware

type Release struct {
	Version string
	Build   string
}

var verified = []Release{
	{Version: "1.0.0", Build: "1295"},
}

func Verified() []Release {
	return append([]Release(nil), verified...)
}

func IsVerified(r Release) bool {
	for _, v := range verified {
		if v == r {
			return true
		}
	}
	return false
}
