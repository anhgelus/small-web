package backend

type data struct {
	title   string
	Article bool
}

func (d data) Title() string {
	title := "anhgelus"
	if d.Article {
		title += " - log entry"
	}
	if len(d.title) != 0 {
		title += " - " + d.title
	}
	return title
}
