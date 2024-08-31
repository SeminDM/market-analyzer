package share

type ByChange []*Share

func (s ByChange) Len() int           { return len(s) }
func (s ByChange) Less(i, j int) bool { return s[i].PriceChangePercent() < s[j].PriceChangePercent() }
func (s ByChange) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
