package abc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//readInformationField returns an error if an instruction was in the wrong syntax.
//unknown information fields are skipped as per the specification.
//if multiple information fields are used, it will be overwritten and the latter is used.
//An exception on this is the change of meter in a Tune.
func (d *Decoder) readInformationField(inline bool) error {
	//information field starts with a letter (capital or non-capital), followed by a ":"
	//information fields that have <instruction>, require a specified syntax.
	//the instructions may also change the Decoder structure to advance through the tunes in the file.
	//therefore, they are checked on top.
	d.skipComments()
	fullLine := ""
	var err error = nil
	if !inline {
		fullLine, err = d.readLine()
	} else {
		fullLine, err = d.readInline()
	}
	if err != nil {
		return err
	}
	informationCharacter := fullLine[0]
	if fullLine[1] != ':' {
		fmt.Printf("found text!")
		return nil //this probably was some kind of text...
	}
	line := string(fullLine[2:len(fullLine)]) //skipping ":"

	switch string(informationCharacter) {
	//I: instruction        <instruction>
	case "I":

	//K: key                <instruction>
	case "K":
		d.tuneHeaderDone = true
		d.Tunes[len(d.Tunes)-1].Key = line

	//L: unit note length   <instruction>
	case "L":

	//M: meter              <instruction>
	case "M":
		divisor := strings.IndexByte(line, '/')
		if divisor == -1 {
			return errors.New("Meter not properly formatted")
		}
		top, err := strconv.ParseUint(line[0:divisor], 10, 64)
		if err != nil {
			return err
		}
		bottom, err := strconv.ParseUint(line[divisor+1:], 10, 64)
		if err != nil {
			return err
		}

		if d.inFileHeader {
			d.MeterTop = top
			d.MeterBottom = bottom

		} else {
			current := &d.Tunes[len(d.Tunes)-1]

			if !d.tuneHeaderDone {
				current.MeterTop = top
				current.MeterBottom = bottom
			} else {
				current.Measures[len(current.Measures)-1].MeterTop = top
				current.Measures[len(current.Measures)-1].MeterBottom = bottom
			}
		}

	//m: macro              <instruction>
	case "m":

	//Q: tempo              <instruction>
	case "Q":

	//s: symbol line        <instruction>
	case "s":

	//U: user defined       <instruction>
	case "U":

	//X: reference number   <instruction>
	case "X":
		d.tuneHeaderDone = false
		num, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return errors.Wrap(err, "reference number of tune could not be parsed")
		}
		d.Tunes = append(d.Tunes, Tune{ReferenceNumber: uint64(num)})
	//V: voice              <instruction>
	case "V":

	//P: parts              <instruction>
	case "P":

	//+ continuation character!
	case "+":
		switch d.lastInformationField {
		case "D":
			if d.inFileHeader {
				d.Discography += "\n" + line
			} else {
				d.Tunes[len(d.Tunes)-1].Discography += "\n" + line
			}
		case "H":
			if d.inFileHeader {
				d.History += "\n" + line
			} else {
				d.Tunes[len(d.Tunes)-1].History += "\n" + line
			}
		case "W":
			fallthrough
		case "w":
			d.Tunes[len(d.Tunes)-1].Words += "\n" + line

		}

	//A: area
	case "A":
		if d.inFileHeader {
			d.Area = line
		} else {
			d.Tunes[len(d.Tunes)-1].Area = line
		}

	//B: book
	case "B":
		if d.inFileHeader {
			d.Book = line
		} else {
			d.Tunes[len(d.Tunes)-1].Book = line
		}

	//C: composer
	case "C":
		if d.inFileHeader {
			d.Composer = line
		} else {
			d.Tunes[len(d.Tunes)-1].Composer = line
		}

	//D: discography
	case "D":
		d.lastInformationField = string(informationCharacter)
		if d.inFileHeader {
			d.Discography = line
		} else {
			d.Tunes[len(d.Tunes)-1].Discography = line
		}
	//F: file url
	case "F":
		if d.inFileHeader {
			d.FileURL = line
		} else {
			d.Tunes[len(d.Tunes)-1].FileURL = line
		}
	//G: group
	case "G":
		if d.inFileHeader {
			d.Group = line
		} else {
			d.Tunes[len(d.Tunes)-1].Group = line
		}
	//H: history
	case "H":
		d.lastInformationField = string(informationCharacter)
		if d.inFileHeader {
			d.History = line
		} else {
			d.Tunes[len(d.Tunes)-1].History = line
		}

	//N: notes
	case "N":
		if d.inFileHeader {
			d.NoteText = line
		} else {
			d.Tunes[len(d.Tunes)-1].NoteText = line
		}
	//O: origin
	case "O":
		if d.inFileHeader {
			d.Origin = line
		} else {
			d.Tunes[len(d.Tunes)-1].Origin = line
		}

	//R: rhythm
	case "R":
		if d.inFileHeader {
			d.Rhythm = line
		} else {
			d.Tunes[len(d.Tunes)-1].Rhythm = line
		}
	//r: remark
	case "r":
		if d.inFileHeader {
			d.Remark = line
		} else {
			d.Tunes[len(d.Tunes)-1].Remark = line
		}
	//S: source
	case "S":
		if d.inFileHeader {
			d.Source = line
		} else {
			d.Tunes[len(d.Tunes)-1].Source = line
		}
	//T: tune title
	case "T":
		if d.Tunes[len(d.Tunes)-1].Title == "" {
			d.Tunes[len(d.Tunes)-1].Title = line
		} else {
			d.Tunes[len(d.Tunes)-1].Title += " \n" + line
		}

	//W: words
	//w: words
	case "W":
		fallthrough
	case "w":
		d.lastInformationField = string(informationCharacter)
		d.Tunes[len(d.Tunes)-1].Words = line
	//Z: transcription
	case "Z":
		if d.inFileHeader {
			d.Transcription = line
		} else {
			d.Tunes[len(d.Tunes)-1].Transcription = line
		}
	}

	//all other information fields are ignored.
	//information fields inlined in a tune body may be enclosed with []

	return nil
}
