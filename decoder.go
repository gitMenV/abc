package abc

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//Tune is a tune in an ABC file.
type Tune struct {
	ReferenceNumber uint64 `json:"referenceNumber,omitempty"`
	Title           string `json:"title"`

	Area           string `json:"area,omitempty"`
	Book           string `json:"book,omitempty"`
	Composer       string `json:"composer,omitempty"`
	Discography    string `json:"discography,omitempty"`
	FileURL        string `json:"fileURL,omitempty"`
	Group          string `json:"group,omitempty"`
	History        string `json:"history,omitempty"`
	Key            string `json:"key,omitempty"`
	UnitNoteLength string `json:"unitNoteLength,omitempty"`
	MeterTop       uint64 `json:"meterTop,omitempty"`
	MeterBottom    uint64 `json:"meterBottom,omitempty"`
	Macro          string `json:"macro,omitempty"`
	NoteText       string `json:"notes,omitempty"`
	Origin         string `json:"origin,omitempty"`
	Parts          string `json:"parts,omitempty"`
	Tempo          string `json:"tempo,omitempty"`
	Rhythm         string `json:"rhythm,omitempty"`
	Remark         string `json:"remark,omitempty"`
	Source         string `json:"source,omitempty"`
	UserDefined    string `json:"userDefined,omitempty"`
	Voice          string `json:"voice,omitempty"`
	Words          string `json:"words,omitempty"`
	Transcription  string `json:"transcription,omitempty"`

	Measures []Measure
}

//Decoder contains structured ABC-file information after decoding.
type Decoder struct {
	r                    *bufio.Reader
	fileCheck            bool
	inFileHeader         bool
	tuneHeaderDone       bool
	lastInformationField string
	halveDuration        bool
	doubleDuration       bool

	Version float32 `json:"abc-version,omitempty"`

	Area           string `json:"area,omitempty"`
	Book           string `json:"book,omitempty"`
	Composer       string `json:"composer,omitempty"`
	Discography    string `json:"discography,omitempty"`
	FileURL        string `json:"fileURL,omitempty"`
	Group          string `json:"group,omitempty"`
	History        string `json:"history,omitempty"`
	UnitNoteLength string `json:"unitNoteLength,omitempty"`
	MeterTop       uint64 `json:"meterTop,omitempty"`
	MeterBottom    uint64 `json:"meterBottom,omitempty"`
	Macro          string `json:"macro,omitempty"`
	NoteText       string `json:"notes,omitempty"`
	Origin         string `json:"origin,omitempty"`
	Rhythm         string `json:"rhythm,omitempty"`
	Remark         string `json:"remark,omitempty"`
	Source         string `json:"source,omitempty"`
	UserDefined    string `json:"userDefined,omitempty"`
	Transcription  string `json:"transcription,omitempty"`

	Tunes []Tune
}

//readInline starts reading from '[' and consumes all up to ']'
//If no '[' is detected, it will do nothing.
func (d *Decoder) readInline() (string, error) {
	b, err := d.r.Peek(1)
	if err != nil {
		return "", err
	}
	if b[0] != '[' {
		return "", nil
	}
	b, err = d.r.ReadBytes(']')
	if err != nil {
		return "", errors.Wrap(err, "reading inline field failed")
	}
	return string(b[0 : len(b)-1]), nil

}

//readMagicNumber reads the first line and returns an error if not an ABC file.
func (d *Decoder) readMagicNumber() error {
	line, err := d.r.ReadBytes('\n')
	if err != nil {
		return errors.Wrap(err, "could not read abc-file version")
	}
	if len(line) <= 4 {
		return errors.Wrap(errors.New("first line too short"), "no abc-file found")
	}
	abc := line[0:4]
	if bytes.Compare(abc, []byte("%abc")) != 0 {
		return errors.Wrap(errors.New("first line not starting with %abc"), "no abc-file found")
	}
	version := line[5 : len(line)-1] //stripping \n
	if len(version) != 0 {
		bigFloat, err := strconv.ParseFloat(string(version), 32)
		if err != nil {
			return errors.Wrap(err, "could not figure out abc-file version number")
		}
		d.Version = float32(bigFloat)
		if d.Version <= 2.0 {
			fmt.Println("WARNING: abc-version is not 2.1+")
		}
	}

	return nil
}

//readLine does not only read a line, but it also trims the comments!
func (d *Decoder) readLine() (string, error) {
	b, err := d.r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	commentStart := len(b) - 1
	for i := 0; i < len(b); i++ {

		if b[i] == '%' {
			commentStart = i
			break
		}
	}
	if commentStart != len(b)-1 {
		return strings.TrimSpace(string(b[0:commentStart])), nil
	}
	return strings.TrimSpace(string(b[0 : len(b)-1])), nil

}

func (d *Decoder) skipSpaces() {
	b, _ := d.r.ReadByte()
	if b == ' ' {
		d.skipSpaces()
	} else {
		d.r.UnreadByte()
	}
}

func (d *Decoder) skipComments() {
	b, _ := d.r.Peek(1)
	if bytes.Compare(b, []byte("%")) == 0 {
		//is comment
		d.r.ReadBytes('\n')
		d.skipComments()
	}
}

//NewDecoder returns a decoder that can be used to decode an ABC file.
//if a filecheck needs to be omitted, set second parameter to true.
func NewDecoder(reader bufio.Reader, disableFileCheck bool) *Decoder {
	return &Decoder{r: &reader, fileCheck: disableFileCheck}
}

//skips the Byte Order Mark if present
func (d *Decoder) skipBOM() error {
	bom, _, err := d.r.ReadRune()
	if err != nil {
		return errors.Wrap(err, "Read operation failed")
	}
	if bom != '\uFEFF' {
		d.r.UnreadRune() //skip the BOM
	}
	return nil
}

//Decode will decode the ABC file and stores everything in the Decoder struct
func (d *Decoder) Decode() error {
	//here, decode the thing. Return err if necessary.

	//abc file structure
	//first character: possibly Byte Order Mark (BOM) ignore this.
	err := d.skipBOM()
	if err != nil {
		return err
	}

	//first line: %abc-<version number>
	if !d.fileCheck {
		err = d.readMagicNumber()
		if err != nil {
			return err
		}
	}

	d.skipComments()

	//optional file header
	err = d.readFileHeader()
	if err != nil {
		return err
	}
	//if there was a file header, the next line is just a newline character.
	b, err := d.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	if len(b) != 1 {
		return errors.Wrap(errors.New("not abc file"), "file header was not followed by an empty newline")
	}

	err = d.readTuneHeader()
	if err != nil {
		return err
	}
	//abc Tunes
	//      tune header
	//            tune header starts with "X:"(reference number) followed by "T:" (title) and finish with "K:" (key)
	//      tune body
	//            music codes
	//termintated by either EOF or newline

	err = d.readTuneBody()
	//a line starting with "%" is a comment and should be ignored.
	//a line starting with "r:" is a remark and behaves like a comment.

	//in abc versions 2.0 or less, there was a dedicated "line continuation character", but not in 2.1+
	//in 2.1+, there are 4 different continuation characters.
	//     for music code, it's backslash "\" or by using I:linebreak/I:linebreak <none> to ignore linebreaks
	//     for information fields, it's "+:" at the following line
	//     for comments, it's just %, so nothing special
	//     for stylesheet directives, it could be I:<directive>

	// line, err := d.r.ReadBytes('\n')
	// if err != nil {
	// 	return err
	// }
	line, _ := d.r.ReadBytes('\n')
	fmt.Println(string(line))
	// fmt.Println(line)

	return nil
}

func (d *Decoder) readTuneHeader() error {
	before := len(d.Tunes)

	//first line must be reference number X
	err := d.readInformationField(false)
	if err != nil {
		return err
	}
	if len(d.Tunes) == before {
		return errors.Wrap(errors.New("not abc file"), "tune header must start with reference number field")
	}

	//second line must be Title of the tune!
	err = d.readInformationField(false)
	if err != nil {
		return err
	}
	if d.Tunes[len(d.Tunes)-1].Title == "" {
		return errors.Wrap(errors.New("not abc file"), "second line of tune header must be Title field")
	}

	for !d.tuneHeaderDone {
		err = d.readInformationField(false)
		if err != nil {
			return err
		}
	}

	return nil
}

//readTuneBody starts reading after the Key information field
//it ends after reading the newline.
//The tune body is read using the BNF from the abc.ebnf file!
//in the BNF, readTuneBody corresponds to abc-music, which is
//abc-music ::= abc-line+
func (d *Decoder) readTuneBody() error {
	//initializing the first measure with one notegroup! Otherwise, it can not be appended to
	d.Tunes[len(d.Tunes)-1].Measures = make([]Measure, 1)
	d.Tunes[len(d.Tunes)-1].Measures[0].NoteGroups = make([]NoteGroup, 1)

	d.skipComments()
	b, err := d.r.Peek(1)
	if err != nil {
		return errors.Wrap(err, "could not read tune body")
	}

	for bytes.Compare(b, []byte("\n")) != 0 {
		err = d.readABCLine()
		if err != nil {
			return err
		}
		d.skipComments()
		nl, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		if nl != '\n' {
			return nil
		}
		b, err = d.r.Peek(1)
		if err != nil {
			return err
		}

	}
	return nil
}

// musicLine ::= comment | (tuneBodyInfoField, lineFeed) | (element, {element} , lineFeed) ;
//can not be comment as these were skipped in previous section.
func (d *Decoder) readABCLine() error {
	b, err := peekLexToken(d.r)
	if err != nil {
		return err
	}
	if b.isTuneBodyInfoField() {
		err = d.readInformationField(false)
		if err != nil {
			return err
		}
	}
	for b.isElement() {
		err = d.readElement()
		if err != nil {
			return err
		}
		b, err = peekLexToken(d.r)
	}

	return nil
}
func (d *Decoder) readPitch() (string, error) {
	rePitch := regexp.MustCompile(`[a-gA-G]`)
	reOctave := regexp.MustCompile(`[',]`)
	retString := ""
	b, err := d.r.ReadByte()
	if err != nil {
		return "", err
	}
	if rePitch.Match([]byte(string(b))) {
		retString += string(b)
	} else {
		return "", errors.Wrap(errors.New("reading note failed"), "still need to implement accidentals in notes.")
	}
	for {
		b, err = d.r.ReadByte()
		if rePitch.Match([]byte(string(b))) {

			d.r.UnreadByte()
			return retString, nil
		} else if reOctave.Match([]byte(string(b))) {
			retString += string(b)
		} else {
			d.r.UnreadByte()
			return retString, nil
		}
	}

}

//element ::= note | annotation | barline | space | inLine | repeat | chord | brokenRhythm;
func (d *Decoder) readElement() error {
	tuneMeasures := &d.Tunes[len(d.Tunes)-1].Measures
	currentMeasure := &(*tuneMeasures)[len((*tuneMeasures))-1]
	// fmt.Println("length of measures:", len((*tuneMeasures)))

	b, err := peekLexToken(d.r)

	if err != nil {
		return err
	}

	if b.isNote() {
		if b.isPitch() {
			pitch, err := d.readPitch() //TODO: this must be done properly.
			if err != nil {
				return err
			}
			duration, err := d.readDuration()
			if err != nil {
				return err
			}
			if d.halveDuration {
				duration /= 2.0
				d.halveDuration = false
			} else if d.doubleDuration {
				duration *= 2.0
				d.doubleDuration = false
			}
			currentMeasure.NoteGroups[len(currentMeasure.NoteGroups)-1].addUnit(&Note{Value: pitch, Duration: duration})
		} else if b.isRest() {

		}
	} else if b.isAnnotation() {
		_, err = d.r.ReadByte() //Can't be anything other than '"'
		if err != nil {
			return err
		}
		fullAnnotation, err := d.r.ReadBytes('"') //reads the full annotation, including the last '"'
		if err != nil {
			return err
		}
		annotation := fullAnnotation[0 : len(fullAnnotation)-2] //stripping the '"'
		fmt.Println("annotation:", annotation)

	} else if b.isSpace() {
		//start a new notegroup, skip space
		currentMeasure.NoteGroups = append(currentMeasure.NoteGroups, NoteGroup{})
		_, err := d.r.ReadByte()
		if err != nil {
			return err
		}
	} else if b.isBarline() {
		//start a new measure, skip barline (could potentially be start/end of repeat)
		barline, err := d.r.ReadByte()
		if err != nil {
			return err
		}

		(*tuneMeasures) = append((*tuneMeasures), Measure{NoteGroups: make([]NoteGroup, 1)})

		currentMeasure := &(*tuneMeasures)[len((*tuneMeasures))-2]
		newMeasure := &(*tuneMeasures)[len((*tuneMeasures))-1]
		switch barline {
		case '[':
			currentMeasure.ThickEnd = true
		case ':':
			currentMeasure.RepeatEnd = true
		}
		newBarline, err := d.r.Peek(1)
		if err != nil {
			return err
		}
		switch newBarline[0] {
		case '|':
			newMeasure.BarlineStart = true
			_, err = d.r.ReadByte()

		case '[':
			newMeasure.ThickStart = true
			_, err = d.r.ReadByte()

		case ':':
			newMeasure.RepeatStart = true
			_, err = d.r.ReadByte()

		}
		if err != nil {
			return err
		}

	} else if b.isInline() {
		err = d.readInformationField(true)
		if err != nil {
			return err
		}
	} else if b.isRepeat() {

	} else if b.isChord() {
		_, err = d.r.ReadByte()
		if err != nil {
			return err
		}
		fullChord, err := d.r.ReadBytes(']')
		if err != nil {
			return err
		}
		chord := fullChord[0 : len(fullChord)-2] //stripping the ending ']'
		fmt.Println("chord:", chord)

	} else if b.isBrokenRhythm() {
		brokenRhythm, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		switch brokenRhythm {
		case '>':
			currNoteGroup := &currentMeasure.NoteGroups[len(currentMeasure.NoteGroups)-1]
			currNoteGroup.Units[len(currNoteGroup.Units)-1].SetDuration(currNoteGroup.Units[len(currNoteGroup.Units)-1].GetDuration() * 2.0)
			d.halveDuration = true
		case '<':
			currNoteGroup := &currentMeasure.NoteGroups[len(currentMeasure.NoteGroups)-1]
			currNoteGroup.Units[len(currNoteGroup.Units)-1].SetDuration(currNoteGroup.Units[len(currNoteGroup.Units)-1].GetDuration() / 2.0)
			d.doubleDuration = true
		}
	}

	return nil
}

//readDuration returns a float with value '1.0' if no duration was found!
func (d *Decoder) readDuration() (float64, error) {
	var nominatorString []byte
	var denominatorString []byte

	var isFraction bool = false
	re := regexp.MustCompile(`[0-9]`)

	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return 0, err
		}
		if re.Match([]byte(string(b))) {
			nominatorString = append(nominatorString, b)
		} else {
			isFraction = (b == '/')
			break
		}
	}

	if len(nominatorString) == 0 && !isFraction {
		d.r.UnreadByte()
		return 1, nil //no duration given.
	}
	var nominator float64 = 0
	var err error
	if len(nominatorString) > 0 {
		nominator, err = strconv.ParseFloat(string(nominatorString), 64)
		if err != nil {
			return 0, err
		}
	}

	if !isFraction {
		d.r.UnreadByte()
		return nominator, nil
	}
	//fraction sign '/' has been read.
	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return 0, err
		}
		if re.Match([]byte(string(b))) {
			denominatorString = append(denominatorString, b)
		} else {
			d.r.UnreadByte()
			break
		}
	}
	if len(denominatorString) == 0 {
		return 0, errors.New("no denominator found in duration")
	}
	denominator, err := strconv.ParseFloat(string(denominatorString), 64)
	if err != nil {
		return 0, err
	}
	if nominator == 0 {
		nominator++
	}
	return nominator / denominator, nil
}

//readfileHeader reads the file header if there is any.
func (d *Decoder) readFileHeader() error {
	b, err := d.r.Peek(1)
	if err != nil {
		return errors.Wrap(err, "read operation failed")
	}
	if string(b) == "X" { //No file header present
		return nil
	}
	d.inFileHeader = true
	var done bool = false
	for !done {
		//forloop read informationFields
		err = d.readInformationField(false)
		b, err = d.r.Peek(1)
		if err != nil {
			return err
		}
		if bytes.Compare(b, []byte("\n")) == 0 {
			//fileHeader is done.
			done = true
		}
	}
	d.inFileHeader = false
	return nil
}
