/*
 Copyright 2022 Qiniu Limited (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmdline

// ---------------------------------------------------------------------------

const (
	escTableBaseChar = '0'
	escTableLen      = ('z' - escTableBaseChar + 1)
)

var escTable = []byte{
	0,    // 0 [48]
	49,   // 1 [49]
	50,   // 2 [50]
	51,   // 3 [51]
	52,   // 4 [52]
	53,   // 5 [53]
	54,   // 6 [54]
	55,   // 7 [55]
	56,   // 8 [56]
	57,   // 9 [57]
	58,   // : [58]
	59,   // ; [59]
	60,   // < [60]
	61,   // = [61]
	62,   // > [62]
	63,   // ? [63]
	64,   // @ [64]
	65,   // A [65]
	66,   // B [66]
	67,   // C [67]
	68,   // D [68]
	69,   // E [69]
	70,   // F [70]
	71,   // G [71]
	72,   // H [72]
	73,   // I [73]
	74,   // J [74]
	75,   // K [75]
	76,   // L [76]
	77,   // M [77]
	78,   // N [78]
	79,   // O [79]
	80,   // P [80]
	81,   // Q [81]
	82,   // R [82]
	83,   // S [83]
	84,   // T [84]
	85,   // U [85]
	86,   // V [86]
	87,   // W [87]
	88,   // X [88]
	89,   // Y [89]
	90,   // Z [90]
	91,   // [ [91]
	92,   // \ [92]
	93,   // ] [93]
	94,   // ^ [94]
	95,   // _ [95]
	96,   // ` [96]
	97,   // a [97]
	98,   // b [98]
	99,   // c [99]
	100,  // d [100]
	101,  // e [101]
	102,  // f [102]
	103,  // g [103]
	104,  // h [104]
	105,  // i [105]
	106,  // j [106]
	107,  // k [107]
	108,  // l [108]
	109,  // m [109]
	'\n', // n [110]
	111,  // o [111]
	112,  // p [112]
	113,  // q [113]
	'\r', // r [114]
	115,  // s [115]
	'\t', // t [116]
	117,  // u [117]
	118,  // v [118]
	119,  // w [119]
	120,  // x [120]
	121,  // y [121]
	122,  // z [122]
	123,  // { [123]
}

func defaultEscape(c byte) string {

	if c-escTableBaseChar < escTableLen {
		c = escTable[c-escTableBaseChar]
	}
	return string(c)
}

// ---------------------------------------------------------------------------
