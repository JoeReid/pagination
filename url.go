package pagination

import (
	"net/url"
	"strconv"
)

func maxItems(u url.URL) (int, bool) {
	maxItems, err := strconv.Atoi(u.Query().Get("maxItems"))
	if err != nil {
		return 0, false
	}

	return maxItems, true
}

func setMaxItems(u url.URL, maxItems int) url.URL {
	q := u.Query()
	q.Set("maxItems", strconv.Itoa(maxItems))

	u.RawQuery = q.Encode()
	return u
}

func page(u url.URL) (string, bool) {
	return u.Query().Get("page"), true
}

func setPage(u url.URL, page string) url.URL {
	q := u.Query()
	q.Set("page", page)

	u.RawQuery = q.Encode()
	return u
}
