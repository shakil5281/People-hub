package utils

import (
	"strings"
)

// UnicodeToBijoy converts Unicode Bangla text to Bijoy 52 / SutonnyMJ ANSI text.
// If the input string is already ASCII / SutonnyMJ ANSI or empty, it is returned as-is.
func UnicodeToBijoy(input string) string {
	if input == "" {
		return ""
	}
	// If no Unicode Bangla characters present, return as-is
	hasBangla := false
	for _, r := range input {
		if r >= 0x0980 && r <= 0x09FF {
			hasBangla = true
			break
		}
	}
	if !hasBangla {
		return input
	}

	// Step 1: Pre-process specific well-known words / phrases for high fidelity
	knownMap := map[string]string{
		"জব কার্ড":                               "PvKzix KvW©",
		"চাকুরী কার্ড":                            "PvKzix KvW©",
		"সময়কাল":                                 "mgqt",
		"কর্মচারীর তথ্য":                           "Kg©Pvixi Z_¨",
		"নাম":                                   "bvg",
		"কর্মচারী আইডি":                          "Kg©x AvBWw",
		"কর্মী আইডি":                             "Kg©x AvBWw",
		"পদবি":                                  "c`ex",
		"পদবী":                                  "c`ex",
		"বিভাগ":                                 "wefvM",
		"ফোন":                                   "gvevBj",
		"মোবাইল":                                 "gvevBj",
		"যোগদানের তারিখ":                         "teveMv‡bi ZvwiL",
		"শিফট":                                  "wkdU",
		"ক্রম":                                   "µt",
		"ক্রমিক":                                 "µt",
		"তারিখ":                                 "ZvwiL",
		"বার":                                   "evi",
		"প্রবেশ":                                 "cÖ‡ek",
		"প্রস্থান":                               "cÖ¯’vb",
		"ঘণ্টা":                                 "Kvvh©NÈv",
		"কর্মঘণ্টা":                             "Kvvh©NÈv",
		"ওটি":                                   "IwU",
		"বিলম্ব (মি.)":                          "wewj¤^",
		"বিলম্ব":                                 "wewj¤^",
		"অবস্থা":                                 "Ae¯’v",
		"সারসংক্ষেপ":                             "mviwms‡¶c",
		"মোট দিন":                               "tgvU Kvvh©wjevm",
		"মোট কার্যদিবস":                         "tgvU Kvvh©wjevm",
		"উপস্থিত":                                "Dcw¯’Z",
		"বিলম্বে":                                 "wewj¤^",
		"অনুপস্থিত":                               "Abycw¯’Z",
		"অর্ধদিবস":                               "AR©w`vem",
		"ছুটি":                                   "QywU",
		"সাপ্তাহিক ছুটি":                         "mvßvwnK QywU",
		"উৎসব ছুটি":                              "miKvix QywU",
		"সরকারি ছুটি":                            "miKvix QywU",
		"মোট বিলম্ব":                             "tgvU wewj¤^",
		"মোট ওভারটাইম":                           "tgvU IfviUvBg",
		"পিপলহাব এইচআর অ্যান্ড পেঅরোল সিস্টেম দ্বারা তৈরি": "wccevnve GBPAvi A¨vÛ ceIivj wm‡÷g `viv ‰Zix",
		"প্রিন্টের তারিখ":                         "wcÖ‡›Ui ZvwiL",
		"রবিবার":                                 "ewebvi",
		"রবি":                                   "ewebvi",
		"সোমবার":                                 "‡mvgevi",
		"সোম":                                   "‡mvgevi",
		"मंगलবার":                                "g½jevi",
		"মঙ্গল":                                  "g½jevi",
		"বুধবার":                                 "evaevi",
		"বুধ":                                   "eva",
		"বৃহস্পতিবার":                             "ewn¯úwZevi",
		"বৃহস্পতি":                               "ewn¯úwZ",
		"শুক্রবার":                               "ïµevi",
		"শুক্র":                                 "ïµevi",
		"শনিবার":                                 "twbvevi",
		"শনি":                                   "twbvevi",
		"স্টাফ":                                 "÷vd",
		"এডমিন":                                 "G¨vWwgb",
		"এডমিন (এ.জি.এম)":                       "G¨vWwgb (G.wR.Gg)",
		"মাষ্টারবাড়ি, গাজীপুর সিটি, গাজীপুর":       "gv÷vevwo, MvRxcyi wmwU, MvRxcyi",
	}

	for k, v := range knownMap {
		if strings.Contains(input, k) {
			input = strings.ReplaceAll(input, k, v)
		}
	}

	if !containsBanglaRune(input) {
		return input
	}

	return convertUnicodeToBijoyStream(input)
}

func containsBanglaRune(s string) bool {
	for _, r := range s {
		if r >= 0x0980 && r <= 0x09FF {
			return true
		}
	}
	return false
}

func convertUnicodeToBijoyStream(input string) string {
	runes := []rune(input)
	n := len(runes)
	var out []string

	i := 0
	for i < n {
		r := runes[i]
		if r < 0x0980 || r > 0x09FF {
			out = append(out, string(r))
			i++
			continue
		}

		// Reph: U+09B0 (র) + U+09CD (্) + Consonant
		if r == 'র' && i+1 < n && runes[i+1] == 0x09CD {
			j := i + 2
			for j < n && (runes[j] == 0x09CD || isBanglaConsonant(runes[j])) {
				j++
			}
			consStr := convertCluster(runes[i+2 : j])
			out = append(out, consStr+"©")
			i = j
			continue
		}

		// Pre-position Vowels: ি (U+09BF), ে (U+09C7), ৈ (U+09C8)
		if isBanglaConsonant(r) {
			j := i + 1
			for j < n && runes[j] == 0x09CD && j+1 < n && isBanglaConsonant(runes[j+1]) {
				j += 2
			}
			consStr := convertCluster(runes[i:j])
			if j < n {
				v := runes[j]
				switch v {
				case 0x09BF: // ি
					out = append(out, "w"+consStr)
					i = j + 1
					continue
				case 0x09C7: // ে
					out = append(out, "‡"+consStr)
					i = j + 1
					continue
				case 0x09C8: // ৈ
					out = append(out, "‰"+consStr)
					i = j + 1
					continue
				case 0x09CB: // ো
					out = append(out, "‡"+consStr+"v")
					i = j + 1
					continue
				case 0x09CC: // ৌ
					out = append(out, "‡"+consStr+"Š")
					i = j + 1
					continue
				}
			}
			out = append(out, consStr)
			i = j
			continue
		}

		out = append(out, mapRuneToBijoy(r))
		i++
	}

	return strings.Join(out, "")
}

func isBanglaConsonant(r rune) bool {
	return (r >= 0x0995 && r <= 0x09B9) || r == 0x09DC || r == 0x09DD || r == 0x09DF
}

func convertCluster(cluster []rune) string {
	if len(cluster) == 1 {
		return mapRuneToBijoy(cluster[0])
	}
	s := string(cluster)
	conjuncts := map[string]string{
		"ক্ষ": "¶", "জ্ঞ": "Á", "ঞ্চ": "Â", "ঞ্ছ": "Ã", "ঞ্জ": "Ä", "ঙ্ক": "•", "ঙ্গ": "½",
		"ষ্ঠ": "ô", "ষ্ণ": "ò", "ষ্ট": "ó", "স্থ": "¯’", "স্প": "¯ú", "স্ত": "¯Í", "ন্ত": "šÍ",
		"ন্দ": "›", "ন্ধ": "Ü", "ম্প": "¤ú", "ম্ভ": "¤¢", "ম্ম": "¤§", "স্ক": "¯‹", "স্ট": "÷",
		"ক্র": "µ", "গ্র": "gÖ", "প্র": "cÖ", "ফ্র": "dÖ", "ব্র": "eÖ", "ভ্র": "fÖ", "ম্র": "gÖ",
		"ত্র": "Î", "দ্র": "æ", "শ্র": "kÖ", "স্ন": "¯œ", "স্ম": "¯§", "ন্ড": "Û", "ন্ট": "›U",
	}
	if v, ok := conjuncts[s]; ok {
		return v
	}
	var b strings.Builder
	for _, r := range cluster {
		b.WriteString(mapRuneToBijoy(r))
	}
	return b.String()
}

func mapRuneToBijoy(r rune) string {
	switch r {
	case 'অ':
		return "A"
	case 'আ':
		return "Av"
	case 'ই':
		return "B"
	case 'ঈ':
		return "C"
	case 'উ':
		return "D"
	case 'ঊ':
		return "E"
	case 'ঋ':
		return "F"
	case 'এ':
		return "G"
	case 'ঐ':
		return "H"
	case 'ও':
		return "I"
	case 'ঔ':
		return "J"
	case 'ক':
		return "k"
	case 'খ':
		return "K"
	case 'গ':
		return "g"
	case 'ঘ':
		return "G"
	case 'ঙ':
		return "q"
	case 'চ':
		return "c"
	case 'ছ':
		return "C"
	case 'জ':
		return "j"
	case 'ঝ':
		return "J"
	case 'ঞ':
		return "N"
	case 'ট':
		return "t"
	case 'ঠ':
		return "T"
	case 'ড':
		return "d"
	case 'ঢ':
		return "D"
	case 'ণ':
		return "Y"
	case 'ত':
		return "Z"
	case 'থ':
		return "_"
	case 'দ':
		return "b"
	case 'ধ':
		return "B"
	case 'ন':
		return "b"
	case 'প':
		return "c"
	case 'ফ':
		return "d"
	case 'ব':
		return "e"
	case 'ভ':
		return "f"
	case 'ম':
		return "g"
	case 'য':
		return "h"
	case 'র':
		return "r"
	case 'ল':
		return "l"
	case 'শ':
		return "k"
	case 'ষ':
		return "l"
	case 'স':
		return "m"
	case 'হ':
		return "n"
	case 0x09DC: // ড়
		return "o"
	case 0x09DD: // ঢ়
		return "p"
	case 0x09DF: // য়
		return "q"
	case 'ৎ':
		return "R"
	case 'ং':
		return "s"
	case 'ঃ':
		return "t"
	case 'ঁ':
		return "u"
	case 0x09BE: // া
		return "v"
	case 0x09BF: // ি
		return "w"
	case 0x09C0: // ী
		return "x"
	case 0x09C1: // ু
		return "y"
	case 0x09C2: // ূ
		return "z"
	case 0x09C3: // ৃ
		return "~"
	case 0x09C7: // ে
		return "‡"
	case 0x09C8: // ৈ
		return "‰"
	case 0x09CD: // ્
		return ""
	case '০':
		return "0"
	case '১':
		return "1"
	case '২':
		return "2"
	case '৩':
		return "3"
	case '৪':
		return "4"
	case '৫':
		return "5"
	case '৬':
		return "6"
	case '৭':
		return "7"
	case '৮':
		return "8"
	case '৯':
		return "9"
	default:
		return string(r)
	}
}
