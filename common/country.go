package common

type Country struct {
	Short   string
	Aliases []string
}

var countries = make(map[string]*Country)

func CountryShort(country string) string {
	c := countries[country]
	if c != nil {
		return c.Short
	}
	return ""
}

func NewCountry(short string) *Country {
	return &Country{
		Short: short,
	}
}

func NewCountryWithAliases(short string, aliases []string) *Country {
	c := NewCountry(short)
	c.Aliases = aliases
	return c
}

func init() {
	countries["China"] = NewCountryWithAliases("CN", []string{"CHN"})
}
