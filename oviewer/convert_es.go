package oviewer

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// convert_es converts an ANSI escape sequence to an ov representation.
// In ov, it is mainly to convert CSI (Control Sequence Introducer).
// Furthermore, it is to convert SGR (Set Graphics Rendition) in it to ov's style.
// (Some hyperlinks are also interpreted).
// Other than that, it hides them so that they do not interfere.

// The states of the ANSI escape code parser.
const (
	ansiText = iota
	ansiEscape
	ansiSubstring
	ansiControlSequence
	otherSequence
	systemSequence
	oscHyperLink
	oscParameter
	oscURL
)

// csiParamStart and csiParamEnd define the range of parameters in the CSI.
// Parameters outside this range will result in an error and will not be considered as CSI.
// Errors within this range will not affect the CSI.
const (
	csiParamStart = 0x20
	csiParamEnd   = 0x3F
)

const (
	// Colors256 is the index of the 256 color. 8-bit colors. 0-255.
	Colors256 = 5
	// ColorsRGB is the index of the RGB color. 24-bit colors. r:0-255 g:0-255 b:0-255.
	ColorsRGB = 2
)

// escapeSequence is a structure that holds the escape sequence.
type escapeSequence struct {
	parameter strings.Builder
	url       strings.Builder
	state     int
}

// newESConverter returns a new escape sequence converter.
func newESConverter() *escapeSequence {
	return &escapeSequence{
		parameter: strings.Builder{},
		url:       strings.Builder{},
		state:     ansiText,
	}
}

// sgrCache caches SGR escape sequences.
var sgrCache sync.Map

// sgrParams is a structure that holds the SGR parameters.
type sgrParams struct {
	code   int
	params []string
	colonF bool
}

// convert parses an escape sequence and changes state.
// Returns true if it is an escape sequence and a non-printing character.
func (es *escapeSequence) convert(st *parseState) bool {
	switch st.mainc {
	case 0x1b:
		es.state = ansiEscape
		return true
	case '\n':
		return false
	}
	return es.paraseEscapeSequence(st)
}

// parseEscapeSequence parses the escape sequence.
// convert parses an escape sequence and changes state.
func (es *escapeSequence) paraseEscapeSequence(st *parseState) bool {
	mainc := st.mainc
	switch es.state {
	case ansiEscape:
		switch mainc {
		case '[': // CSI(Control Sequence Introducer).
			es.parameter.Reset()
			es.state = ansiControlSequence
			return true
		case 'c': // Reset.
			st.style = tcell.StyleDefault
			es.state = ansiText
			return true
		case ']': // Operating System Command Sequence.
			es.state = systemSequence
			return true
		case 'P', 'X', '^', '_': // Substrings and commands.
			es.state = ansiSubstring
			return true
		case '(':
			es.state = otherSequence
			return true
		default: // Ignore.
			es.state = ansiText
		}
	case ansiSubstring:
		if mainc == 0x1b {
			es.state = ansiControlSequence
		}
		return true
	case ansiControlSequence:
		es.parseCSI(st, mainc)
		return true
	case otherSequence:
		es.state = ansiEscape
		return true
	case systemSequence:
		switch mainc {
		case '8':
			es.state = oscHyperLink
			return true
		case '\\':
			es.state = ansiText
			return true
		case 0x1b:
			// unknown but for compatibility.
			es.state = ansiControlSequence
			return true
		case 0x07:
			es.state = ansiText
			return true
		}
		log.Printf("invalid char %c", mainc)
		return true
	case oscHyperLink:
		switch mainc {
		case ';':
			es.state = oscParameter
			return true
		}
		es.state = ansiText
		return false
	case oscParameter:
		if mainc != ';' {
			es.parameter.WriteRune(mainc)
			return true
		}
		urlID := es.parameter.String()
		if urlID != "" {
			st.style = st.style.UrlId(urlID)
		}
		es.parameter.Reset()
		es.state = oscURL
		return true
	case oscURL:
		switch mainc {
		case 0x1b:
			st.style = st.style.Url(es.url.String())
			es.url.Reset()
			es.state = systemSequence
			return true
		case 0x07:
			st.style = st.style.Url(es.url.String())
			es.url.Reset()
			es.state = ansiText
			return true
		}
		es.url.WriteRune(mainc)
		return true
	}
	return false
}

// parseCSI parses the CSI(Control Sequence Introducer) escape sequence.
func (es *escapeSequence) parseCSI(st *parseState, mainc rune) {
	switch {
	case mainc == 'm': // SGR(Set Graphics Rendition).
		st.style = sgrStyle(st.style, es.parameter.String())
	case mainc == 'K': // Erase in Line.
		// CSI 0 K or CSI K maintains the style after the newline
		params := es.parameter.String()
		if params == "" || params == "0" {
			// can change the background color of the line.
			_, bg, _ := st.style.Decompose()
			st.eolStyle = st.eolStyle.Background(bg)
		}
	case mainc >= 'A' && mainc <= 'T': // Cursor Movement.
		// Ignore.
	case mainc >= csiParamStart && mainc <= csiParamEnd: // Parameters.
		es.parameter.WriteRune(mainc)
		return
	}
	// End of escape sequence.
	es.state = ansiText
}

// sgrStyle returns tcell.Style from the SGR control sequence.
func sgrStyle(style tcell.Style, paramStr string) tcell.Style {
	switch paramStr {
	case "0", "", ";":
		return tcell.StyleDefault.Normal()
	}

	if s, ok := sgrCache.Load(paramStr); ok {
		style = applyStyle(style, s.(OVStyle))
		return style
	}

	s := parseSGR(paramStr)
	sgrCache.Store(paramStr, s)
	return applyStyle(style, s)
}

// parseSGR actually parses the style and returns OVStyle.
func parseSGR(paramStr string) OVStyle {
	s := OVStyle{}
	paramList := strings.Split(paramStr, ";")
	for index := 0; index < len(paramList); index++ {
		sgr, err := toSGRCode(paramList, index)
		if err != nil {
			return OVStyle{}
		}

		switch sgr.code {
		case 0: // Reset.
			s = OVStyle{}
		case 1: // Bold On
			s.Bold = true
			s.UnBold = false
		case 2: // Dim On
			s.Dim = true
			s.UnDim = false
		case 3: // Italic On
			s.Italic = true
			s.UnItalic = false
		case 4: // Underline On
			if len(sgr.params) > 0 && sgr.params[0] != "" {
				s = underLineStyle(s, sgr.params[0])
				break
			}
			s.Underline = true
			s.UnUnderline = false
		case 5: // Blink On
			s.Blink = true
			s.UnBlink = false
		case 6: // Rapid Blink On
			s.Blink = true // Rapid Blink is the same as Blink.
			s.UnBlink = false
		case 7: // Reverse On
			s.Reverse = true
			s.UnReverse = false
		case 8: // Invisible On
			// (not implemented)
		case 9: // StrikeThrough On
			s.StrikeThrough = true
			s.UnStrikeThrough = false
		case 21: // Double Underline On
			s.Underline = true // Double Underline is the same as Underline.
			s.UnUnderline = false
		case 22: // Bold Off
			s.Bold = false
			s.UnBold = true
		case 23: // Italic Off
			s.Italic = false
			s.UnItalic = true
		case 24: // Underline Off
			s.Underline = false
			s.UnUnderline = true
		case 25: // Blink Off
			s.Blink = false
			s.UnBlink = true
		case 27: // Reverse Off
			s.Reverse = false
			s.UnReverse = true
		case 28: // Invisible Off
			// (not implemented)
		case 29: // StrikeThrough Off
			s.StrikeThrough = false
			s.UnStrikeThrough = true
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground color
			s.Foreground = colorName(sgr.code - 30)
		case 38: // Foreground color extended
			color, i, err := parseSGRColor(sgr)
			if err != nil {
				return OVStyle{}
			}
			index += i
			s.Foreground = color
		case 39: // ForegroundColorDefault
			s.Foreground = "default"
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background color
			s.Background = colorName(sgr.code - 40)
		case 48: // Background color extended
			color, i, err := parseSGRColor(sgr)
			if err != nil {
				return OVStyle{}
			}
			index += i
			s.Background = color
		case 49: // BackgroundColorDefault
			s.Background = "default"
		case 53: // Overline On
			s.OverLine = true
		case 55: // Overline Off
			s.UnOverLine = true
		case 58: // UnderlineColor
			// (not implemented). Increase index only.
			_, i, err := parseSGRColor(sgr)
			if err != nil {
				return s
			}
			index += i
		case 59: // UnderlineColorDefault
			// (not implemented).
		case 73, 74, 75: // VerticalAlignment
			// (not implemented).
		case 90, 91, 92, 93, 94, 95, 96, 97: // Bright Foreground color
			s.Foreground = colorName(sgr.code - 82)
		case 100, 101, 102, 103, 104, 105, 106, 107: // Bright Background color
			s.Background = colorName(sgr.code - 92)
		}
	}
	return s
}

// toSGRCode converts the SGR parameter to a code.
// If there is no (:) separator, use the slice of paramList.
func toSGRCode(paramList []string, index int) (sgrParams, error) {
	str := paramList[index]
	sgr := sgrParams{}
	colonLists := strings.Split(str, ":")
	code, err := sgrNumber(colonLists[0])
	if err != nil {
		return sgrParams{}, ErrNotSuuport
	}
	sgr.code = code

	// If the colon parameter is used, interpret the first paramList as the code.
	if len(colonLists) > 1 {
		sgr.params = colonLists[1:]
		sgr.colonF = true
		return sgr, nil
	}
	// If the colon parameter is not used, interpret the following paramList as parameters.
	if code == 38 || code == 48 || code == 58 {
		if len(paramList) > index+1 {
			sgr.params = paramList[index+1:]
			sgr.colonF = false
		}
	}
	return sgr, nil
}

// sgrNumber converts a string to a number.
// If the string is empty, it returns 0.
// If the string contains a non-numeric character, it returns an error.
func sgrNumber(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	if containsNonDigit(str) {
		return 0, ErrNotSuuport
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// containsNonDigit returns true if the string contains a non-numeric character.
func containsNonDigit(str string) bool {
	for _, char := range str {
		if !unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// underLineStyle sets the underline style.
func underLineStyle(s OVStyle, param string) OVStyle {
	n, err := sgrNumber(param)
	if err != nil {
		return s
	}

	// Support only Underline Off (4:0).
	if n == 0 {
		s.Underline = false
		s.UnUnderline = true
		return s
	}
	// Other than that, it is treated as Underline On.
	s.Underline = true
	s.UnUnderline = false
	return s
}

// parseSGRColor parses 256 color or RGB color.
// Returns the color name and increase in the index (the colon does not increase).
func parseSGRColor(sgr sgrParams) (string, int, error) {
	color, inc, error := convertSGRColor(sgr)
	if sgr.colonF { // In the case of colon, index does not increase.
		inc = 0
	}
	return color, inc, error
}

// convertSGRColor converts the SGR color to a string that can be used to specify the color of tcell.
// There are three ways to specify the extended color:
// 38;5;n or 38;2;n;n;n semi-colon separated.
// 38:5:n or 38:2:n:n:n colon separated.
// 38:2::n:n:n colon separated with double colon.
//
// The first return value is the color name.
// The second return value is the number of parameter index increments.
// The last return value is an error.
// Illegal characters will result in an error, while others will be ignored without error.
func convertSGRColor(sgr sgrParams) (string, int, error) {
	if len(sgr.params) == 0 {
		return "", 0, nil
	}
	inc := 1
	ex, err := sgrNumber(sgr.params[0])
	if err != nil {
		return "", inc, err
	}
	switch ex {
	case Colors256: // 38:5:n
		if len(sgr.params) < 2 {
			return "", inc, nil
		}
		color, err := parse256Color(sgr.params[1])
		if err != nil {
			return color, inc, err
		}
		inc++
		return color, inc, nil
	case ColorsRGB: // 38:2:r:g:b
		if len(sgr.params) < 4 {
			return "", len(sgr.params), nil
		}
		rgb := sgr.params[1:4] // 38:2:r:g:b
		// The colon(colonF) parameter allows two colons(::) to be done before the RGB is specified.
		if sgr.colonF && sgr.params[1] == "" && len(sgr.params) > 4 {
			rgb = sgr.params[2:5] // 38:2::r:g:b
		}
		color, err := parseRGBColor(rgb[0], rgb[1], rgb[2])
		if err != nil {
			return color, inc, err
		}
		inc += 3
		return color, inc, nil
	}
	return "", inc, nil
}

// parse256Color parses the 8-bit color.
func parse256Color(param string) (string, error) {
	if param == "" {
		return "", nil
	}
	c, err := sgrNumber(param)
	if err != nil {
		return "", err
	}
	color := colorName(c)
	return color, nil
}

// colorName returns a string that can be used to specify the color of tcell.
func colorName(colorNumber int) string {
	if colorNumber < 0 || colorNumber > 255 {
		return ""
	}
	return tcell.PaletteColor(colorNumber).String()
}

// parseRGBColor parses the RGB color.
func parseRGBColor(red string, green string, blue string) (string, error) {
	if red == "" || green == "" || blue == "" {
		return "", nil
	}
	r, err1 := sgrNumber(red)
	g, err2 := sgrNumber(green)
	b, err3 := sgrNumber(blue)
	if err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("invalid RGB color values: %v, %v, %v", red, green, blue)
	}
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return "", nil
	}
	color := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	return color, nil
}
