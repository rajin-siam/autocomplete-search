package nominatim

type RankConfig struct {
	RankToLayer map[int]string
	RankToField map[int]string
}

func DefaultRankConfig() *RankConfig {
	return &RankConfig{
		RankToLayer: map[int]string{
			4: "country", 8: "state", 12: "county",
			16: "city", 17: "city", 18: "city", 19: "city", 20: "city", 21: "city",
			22: "district", 23: "district", 24: "district", 25: "district",
			26: "locality", 27: "locality", 28: "locality", 29: "locality",
			30: "house",
		},
		RankToField: map[int]string{
			8: "state", 12: "county",
			16: "city", 17: "city", 18: "city", 19: "city", 20: "city",
			22: "district", 23: "district", 24: "district", 25: "district",
			26: "locality", 27: "locality", 28: "locality",
		},
	}
}
