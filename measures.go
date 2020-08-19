package abc

//Measure is just one measure of the song.
type Measure struct {
	MeterTop     uint64 `json:"meterTop,omitempty"`
	MeterBottom  uint64 `json:"meterBottom,omitempty"`
	RepeatStart  bool   `json:"repeatStart,omitempty"`
	RepeatEnd    bool   `json:"repeatEnd,omitempty"`
	ThickStart   bool   `json:"startThick,omitempty"`
	ThickEnd     bool   `json:"endThick,omitempty"`
	BarlineStart bool   `json:"barlineStart,omitempty"`
	NoteGroups   []NoteGroup
}

//NoteGroup denotes one group of notes that should be paired using a beam.
type NoteGroup struct {
	Units []Unit
}

func (ng *NoteGroup) addUnit(unit Unit) {
	ng.Units = append(ng.Units, unit)
}

//Unit is an interface that is implemented by either chord, note or rest
type Unit interface {
	GetValue() string
	GetDuration() float64
	SetDuration(float64)
}

//Note holds the information of a single note in music.
//This could also be a rest, in which case the value is Z or x
//The duration is the duration in terms of spaces it occupies in the measure.
//or, how many units!
//in case of a 1/4 measure, a duration of 0.5 will be an eigth note.
//A duration of 2 will be a half note.
//A duration of 3 will be a three-quarter note etc.
//this should scale to 1/128th notes. Does it? Float imprecisions...
type Note struct {
	Value    string  `json:"value"`
	Duration float64 `json:"duration"`
}

//Rest is a simple way to denote the rest in a measure.
type Rest struct {
	Duration float64 `json:"duration"`
}

//Chord holds the values of a chord.
type Chord struct {
	notes    []Note
	Value    string  `json:"value"`
	Duration float64 `json:"duration"`
}

//GetValue returns a string with a comma separated list of values
func (c *Chord) GetValue() string {
	return c.Value
}

//SetDuration is used for when '<' or '>' is encountered.
func (c *Chord) SetDuration(f float64) {
	c.Duration = f
}

//GetDuration returns the duration of the chord; each chord can only have one duration!
//So, a chord consisting of 3 notes can not have different durations for the notes.
func (c *Chord) GetDuration() float64 {
	return c.Duration
}

//GetValue returns the note value
func (n *Note) GetValue() string {
	return n.Value
}

//SetDuration is used for when '<' or '>' is encountered.
func (n *Note) SetDuration(f float64) {
	n.Duration = f
}

//GetDuration returns the duration of the note
func (n *Note) GetDuration() float64 {
	return n.Duration
}

//GetValue returns the note value
func (r *Rest) GetValue() string {
	return "z"
}

//GetDuration returns the duration of the note
func (r *Rest) GetDuration() float64 {
	return r.Duration
}

//SetDuration is used for when '<' or '>' is encountered.
func (r *Rest) SetDuration(f float64) {
	r.Duration = f
}
