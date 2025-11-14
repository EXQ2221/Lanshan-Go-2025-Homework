package song

type Song struct {
	Name           string
	Difficulty     float32
	Score          float32
	PotentialPoint float32
}

func Getallsongs() []Song {
	songs := []Song{
		{"Testify", 12.0, 0, 0},
		{"Desingant", 11.9, 0, 0},
		{"Tempestissimo", 11.7, 0, 0},
		{"Arcana Eden", 11.6, 0, 0},
		{"Aether Crest : Astral", 11.5, 0, 0},
		{"Pentiment", 11.5, 0, 0},
		{"Lament Rain", 11.4, 0, 0},
		{"ALTER EGO", 11.3, 0, 0},
		{"Fracture Ray", 11.1, 0, 0},
		{"Grievous Lady", 11.1, 0, 0},
	}
	return songs

}
