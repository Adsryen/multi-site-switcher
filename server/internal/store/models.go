package store

type Site struct {
	Key      string `db:"key" json:"key"`
	Name     string `db:"name" json:"name"`
	LoginURL string `db:"login_url" json:"loginUrl"`
	Created  int64  `db:"created_at" json:"createdAt"`
	Updated  int64  `db:"updated_at" json:"updatedAt"`
}

type Account struct {
	ID       string `db:"id" json:"id"`
	SiteKey  string `db:"site_key" json:"siteKey"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
	Extra    string `db:"extra" json:"extra"`
	Created  int64  `db:"created_at" json:"createdAt"`
	Updated  int64  `db:"updated_at" json:"updatedAt"`
}

type SiteFieldSchema struct {
	SiteKey      string `db:"site_key" json:"siteKey"`
	Field        string `db:"field" json:"field"`
	Type         string `db:"type" json:"type"`
	Required     int    `db:"required" json:"required"` // 0/1
	DefaultValue string `db:"default_value" json:"default"`
	Regex        string `db:"regex" json:"regex"`
	Choices      string `db:"choices" json:"choices"` // JSON array
	Secret       int    `db:"secret" json:"secret"`   // 0/1
	Order        int    `db:"order" json:"order"`
	UIHint       string `db:"ui_hint" json:"uiHint"`
}
