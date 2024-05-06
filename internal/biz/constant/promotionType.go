package constant

// TorrentPromotionTypes represents the different promotion types available.
var TorrentPromotionTypes = map[int64]PromotionType{
	TORRENT_PROMOTION_NORMAL: {
		Text:           "Normal",
		UpMultiplier:   1,
		DownMultiplier: 1,
		Color:          "",
	},
	TORRENT_PROMOTION_FREE: {
		Text:           "Free",
		UpMultiplier:   1,
		DownMultiplier: 0,
		Color:          "linear-gradient(to right, rgba(0,52,206,0.5), rgba(0,52,206,1), rgba(0,52,206,0.5))",
	},
	TORRENT_PROMOTION_TWO_TIMES_UP: {
		Text:           "2X",
		UpMultiplier:   2,
		DownMultiplier: 1,
		Color:          "linear-gradient(to right, rgba(0,153,0,0.5), rgba(0,153,0,1), rgba(0,153,0,0.5))",
	},
	TORRENT_PROMOTION_FREE_TWO_TIMES_UP: {
		Text:           "2X Free",
		UpMultiplier:   2,
		DownMultiplier: 0,
		Color:          "linear-gradient(to right, rgba(0,153,0,1), rgba(0,52,206,1)",
	},
	TORRENT_PROMOTION_HALF_DOWN: {
		Text:           "50%",
		UpMultiplier:   1,
		DownMultiplier: 0.5,
		Color:          "linear-gradient(to right, rgba(220,0,3,0.5), rgba(220,0,3,1), rgba(220,0,3,0.5))",
	},
	TORRENT_PROMOTION_HALF_DOWN_TWO_TIMES_UP: {
		Text:           "2X 50%",
		UpMultiplier:   2,
		DownMultiplier: 0.5,
		Color:          "linear-gradient(to right, rgba(0,153,0,1), rgba(220,0,3,1)",
	},
	TORRENT_PROMOTION_ONE_THIRD_DOWN: {
		Text:           "30%",
		UpMultiplier:   1,
		DownMultiplier: 0.3,
		Color:          "linear-gradient(to right, rgba(65,23,73,0.5), rgba(65,23,73,1), rgba(65,23,73,0.5))",
	},
}

// PromotionType represents a single promotion type.
type PromotionType struct {
	Text           string  `json:"text"`
	UpMultiplier   float64 `json:"up_multiplier"`
	DownMultiplier float64 `json:"down_multiplier"`
	Color          string  `json:"color"`
}
