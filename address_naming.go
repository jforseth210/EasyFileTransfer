package main

import (
	"net"
	"net/netip"
	"strings"
)

// wordfrequency.info 256 most common nouns
var wordList = []string{
	"time", "people", "year", "way", "thing", "man", "day", "life", "woman", "world", "child", "school", "state", "family", "president", "house", "student", "part", "place", "problem", "country", "week", "point", "hand", "group", "guy", "case", "question", "work", "night", "game", "number", "money", "lot", "book", "system", "government", "city", "company", "story", "job", "friend", "word", "fact", "right", "month", "program", "business", "home", "kind", "study", "issue", "name", "idea", "room", "percent", "law", "power", "kid", "war", "head", "mother", "team", "eye", "side", "water", "service", "area", "person", "end", "hour", "line", "girl", "father", "information", "car", "minute", "party", "back", "health", "reason", "member", "community", "news", "body", "level", "boy", "university", "change", "center", "face", "food", "history", "result", "morning", "parent", "office", "research", "door", "court", "moment", "street", "policy", "table", "care", "process", "teacher", "data", "death", "experience", "plan", "education", "age", "sense", "show", "college", "music", "mind", "class", "police", "use", "effect", "season", "tax", "heart", "son", "art", "market", "air", "force", "foot", "baby", "love", "republican", "interest", "security", "control", "rate", "report", "nation", "action", "wife", "decision", "value", "phone", "thanks", "event", "site", "church", "model", "relationship", "movie", "field", "player", "couple", "record", "difference", "light", "development", "role", "view", "price", "effort", "voice", "department", "leader", "photo", "space", "project", "position", "million", "film", "need", "type", "town", "article", "road", "form", "chance", "drug", "situation", "practice", "science", "brother", "matter", "image", "star", "cost", "post", "society", "picture", "piece", "paper", "energy", "building", "doctor", "activity", "american", "media", "evidence", "product", "arm", "technology", "comment", "look", "term", "color", "choice", "source", "mom", "director", "rule", "campaign", "ground", "election", "page", "test", "patient", "video", "support", "rest", "step", "opportunity", "official", "oil", "call", "organization", "character", "county", "future", "dad", "industry", "second", "list", "stuff", "figure", "attention", "risk", "fire", "dog", "hair", "condition", "wall", "daughter", "deal", "author", "truth", "husband", "period", "series", "order", "officer", "land", "computer", "thought", "economy"}

// Convert an IP address into four common English nouns.
func EncodeIPToWords(ip net.IP) string {
	// Get digits of address
	digits := ip.To4()

	// Look up the word at that index and append it to the list
	var digitStrings []string
	for _, digit := range digits {
		digitStrings = append(digitStrings, wordList[digit])
	}
	
	// Join the words together
	return strings.Join(digitStrings, " ")
}

// Convert four common English nouns into an IP address.
func DecodeIPFromWords(str string) net.IP {
	// Split the string into words
	digitWords := strings.Split(str, " ")
	// Make sure it's four words
	if len(digitWords) != 4 {
		return nil
	}
	actualDigits := []byte{}
	// Find each word in the word list
	for _, digitWord := range digitWords {
		for j, word := range wordList {
			if digitWord == word {
				actualDigits = append(actualDigits, byte(j))
			}
		}
	}
	// Try to convert to an IP
	ip, ok := netip.AddrFromSlice(actualDigits)
	if !ok {
		// Not a valid IP
		return net.IP{}
	}
	// Convert netip.Addr to net.IP
	return net.IP(ip.AsSlice())
}
