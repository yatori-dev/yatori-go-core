package ctype

type CardType string

const (
	Video      CardType = "insertvideo"
	Work       CardType = "work"
	Insertdoc  CardType = "insertdoc"
	Document   CardType = "document"
	InsertBook CardType = "insertbook"
	Hyperlink  CardType = "hyperlink"
	Insertlive CardType = "insertlive"
)
