package components

// TitleCard renders the brand header that sits above the table.
func TitleCard() string {
	brand := TitleBrandStyle.Render("mash")
	sub := TitleSubStyle.Render("The Connection Manager")
	return TitleCardStyle.Render(" " + brand + "  " + sub)
}
